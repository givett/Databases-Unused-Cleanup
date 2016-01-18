# Upgrade Notes

When upgrading from pre v0.2.25 to v0.2.25 the following may need to be done:
 - Have to manually start/stop/restart monit and postgres several times to get past ddl locking issues with the rdpg database (edited)
 - The extensions were added to many of the user databases (maybe +50%) using the bash script shown in the `Scripts` section.  The rest had ddl-lock errors.  Of these I took a peak and the ones having the locking errors only have a single entry in `bdr.bdr_connections`.  If we can get these fixed up I can run the script which will add the extensions against the failed databases.
 - The upgrade to BDR 0.9.2 from 0.9.1 kept generating errors saying `“bdr” has no update path from version "0.9.2" to “0.9.1”` while running through startup.  The only fix we found was to stop Postgres on all 3 servers and start them up one at a time.  It was odd because running `select * from pg_available_extensions` would show only 0.9.2 and all the BOSH packages matched ids (edited)
 - The previous deploy had references to scheduled tasks which never did or will exist.  These were manually deleted.
 - The `max_wal_senders` in `/var/vcap/store/pgbdr/data/postgresql.conf` now is set to equal with  `max_replication_slots`.  The value was at 50 and thus only the first 50 databases were allowed to replicate.  `max_connections` must also be greater than `max_wal_senders`.
 - The number of databases allowed for bdr replication is controlled by `max_worker_processes`.  See the section `Have more that 100 bdr databases` at the bottom of this document to calculate the value which is in `postgres.conf`


## Scripts

### Roll out new extensions
The bash script below is run against one node in each of the clusters.  It will add the 4 new extensions to each existing databases.  New databases will already have these extensions when they are created with the `PrecreateDatabases` task.

```bash
#! /bin/bash

# Add new extensions to existing databases
PSQL="/var/vcap/packages/pgbdr/bin/psql -p7432 -U vcap"

DBS=(
rdpg
$($PSQL -l -t | awk '{print $1}' | awk '/^d/')
)

# now loop through each individual database and create extensions
for database in ${DBS[@]}; do
    echo "Database: $database"
    $PSQL $database -c 'CREATE EXTENSION IF NOT EXISTS pgcrypto;'
    $PSQL $database -c 'CREATE EXTENSION IF NOT EXISTS pg_stat_statements;'
    $PSQL $database -c 'CREATE EXTENSION IF NOT EXISTS "uuid-ossp";'
    $PSQL $database -c 'CREATE EXTENSION IF NOT EXISTS hstore;'

done

```


### Performing Manual Backups
The bash script below is run against one node in each of the clusters.  The globals and users are dumped for the cluster first then each of the non-system databases are backed up into separate files.  See the note at the bottom on why a single `pg_dumpall` does not work.

```bash
#! /bin/bash

# http://serverfault.com/questions/59838/whats-the-best-way-to-automate-backing-up-of-postgresql-databases
# backup-postgresql.sh
# by Craig Sanders
# this script is public domain.  feel free to use or modify as you like.

DUMPALL="/var/vcap/packages/pgbdr/bin/pg_dumpall -p7432 -U vcap"
PGDUMP="/var/vcap/packages/pgbdr/bin/pg_dump -p7432 -U vcap"
PSQL="/var/vcap/packages/pgbdr/bin/psql -p7432 -U vcap"


# directory to save backups in, must be rwx by postgres user
BASE_DIR="/var/vcap/store/pgbdr"
YMD=$(date "+%Y-%m-%d")
DIR="$BASE_DIR/$YMD"
mkdir -p $DIR
cd $DIR

# get list of databases in system , exclude the tempate dbs

DBS=(
rdpg
$($PSQL -l -t | awk '{print $1}' | awk '/^d/')
)

# next dump globals (roles and tablespaces) only
$DUMPALL -g | gzip -9 > "$DIR/globals.gz"

# now loop through each individual database and backup the schema and data separately
for database in ${DBS[@]}; do
    SCHEMA=$DIR/$database.schema.gz
    DATA=$DIR/$database.data.gz

    # export data from postgres databases to plain text
    $PGDUMP -C -c -s -N "bdr" $database | gzip -9 > $SCHEMA

    # dump data
    $PGDUMP -a  -N "bdr" $database | gzip -9 > $DATA
done

```
#### Note:

You cannot run a generic `pg_dumpall` because you receive the following error:
```
pg_dump: [archiver (db)] connection to database "bdr_supervisordb" failed: FATAL:  The BDR extension reserves the database bdr_supervisordb for its own use
HINT:  Use a different database
pg_dumpall: pg_dump failed on database "bdr_supervisordb", exiting
```


### Resync BDR on Databases with Broken replication

Step 1 - Find user databases with broken replication
```bash
#! /bin/bash

PSQL="/var/vcap/packages/pgbdr/bin/psql -p7432 -U vcap"

DBS=(
$($PSQL -l -t | awk '{print $1}' | awk '/^d/')
)

# On server A (good) select the databases which exist but have no partner
a=0
for database in ${DBS[@]}; do
    x=$($PSQL $database -c 'SELECT count(*) FROM bdr.bdr_nodes;'  )
    y=$(echo $x |  awk '{print $3}')
    if [[ "$y" != "2" ]]; then
      echo "$database"
      let a=a+1
    fi

done
echo "Total databases: $a"

# Against the server which returns the above results
dbname="d8dcf0edd47474fb0a374a9485aad2866"
$PSQL rdpg -c "select cluster_id, dbname, dbuser, dbpass from cfsb.instances where dbname='"${dbname}"';"

# Switch to server A (bad) server, use the info from the last step to populate dbname, dbuser and dbpass below
dbname="d8dcf0edd47474fb0a374a9485aad2866"; dbuser="u8dcf0edd47474fb0a374a9485aad2866"
dbpass="nice_try...."  #Put the correct password in here

$PSQL rdpg -c "CREATE USER ${dbuser};"  #On first attempt, the user was already successful, may be able to skip these two steps
$PSQL rdpg -c "ALTER USER ${dbuser} ENCRYPTED PASSWORD '"${dbpass}"';"

#Create the database
$PSQL postgres -c "CREATE DATABASE ${dbname} WITH OWNER ${dbuser} TEMPLATE template0 ENCODING 'UTF8';"

#Apply user permissions
$PSQL postgres -c "REVOKE ALL ON DATABASE \"${dbname}\" FROM public;"
$PSQL postgres -c "GRANT ALL PRIVILEGES ON DATABASE ${dbname} TO ${dbuser};"

# Now log into that new database and add extensions:
$PSQL $dbname -c "CREATE EXTENSION IF NOT EXISTS btree_gist;"
$PSQL $dbname -c "CREATE EXTENSION IF NOT EXISTS bdr;"

#Finally join the new database to it's partner
nodeName="${dbname}_b"
port="7432"
myIP="10.202.84.36"
targetIP="10.202.84.37"
repUser="postgres"

# Now execute one of the two statements below, run the first one on the first node, the second one on the second node
$PSQL $dbname -c "SELECT bdr.bdr_group_create( local_node_name := '"${nodeName}"', node_external_dsn := 'host="${myIP}" port="${port}" user="${repUser}" dbname="${dbname}"'); "
$PSQL $dbname -c "SELECT bdr.bdr_group_join( local_node_name := '"${nodeName}"', node_external_dsn := 'host="${myIP}" port="${port}" user="${repUser}" dbname="${dbname}"', join_using_dsn := 'host=${targetIP} port="${port}" user="${repUser}" dbname="$dbname"'); "

$PSQL $dbname -c "SELECT bdr.bdr_node_join_wait_for_ready();"

```

### Not enough connection

Make sure that `max_connections` is larger than `max_wal_senders` or postgres won't start if there are more bdr replicated databases consuming replication slots than there are connections available.

```
2015-08-13T19:47:00$> /var/vcap/jobs/pgbdr/bin/control
semmsl: 250 semmns: 32000 semopm: 32 semmni: 128
postgres: max_wal_senders must be less than max_connections
2015-08-13T19:49:10$> /var/vcap/jobs/pgbdr/bin/control
semmsl: 250 semmns: 32000 semopm: 32 semmni: 128
postgres: max_wal_senders must be less than max_connections
```

### Have more that 100 bdr databases

According to the bdr extension c code the maximum number of databases is dictated by the following calculation:
`bdr_max_databases = (max_worker_processes / 2) + 1;`

So make sure the `max_worker_processes` is set correctly in `postgresql.conf`
```
d= p=11798 a=LOCATION:  bdr_perdb_worker_main, bdr_perdb.c:707
d= p=11798 a=ERROR:  53400: Too many databases BDR-enabled for bdr.max_databases
d= p=11798 a=HINT:  Increase bdr.max_databases above the current limit of 101
d= p=11798 a=LOCATION:  bdr_locks_find_database, bdr_locks.c:253
d= p=7055 a=LOG:  00000: worker process: bdr db: d7a74b62d8ecc427aadea45cf3ff86a8a (PID 11798) exited with exit code 1
```

### Drop one of the databases with BDR replication

If you want to drop the database on one of the nodes, you can run the following, substituting in the correct database name.  You need to PSQL into the `rdpg` or other database first, note that you cannot drop a database your are currently logged into.
```
select pg_terminate_backend(pid) from pg_stat_activity where datname='d8dcf0edd47474fb0a374a9485aad2866';
WITH slot AS (select slot_name from pg_replication_slots WHERE database = 'd8dcf0edd47474fb0a374a9485aad2866' limit 1)
SELECT pg_drop_replication_slot(slot_name) FROM slot;
drop database d8dcf0edd47474fb0a374a9485aad2866;
```
