# Backups

## Default Backups
For now backups of user databases are scheduled automatically without need for user interaction.  Whenever a user database is created a scheduled task is automatically created to perform a daily backup.

## How it Works
Whenever a new user database is created by the RDPG daemon it is inserted into the cfsb.instances table in the rdpg database.  There is a task called `ScheduleNewDatabaseBackups` in tasks.schedules which runs every minute and compares the cfsb.instances table to tasks.schedules.  If there are user databases without a scheduled backup task one is created.

By default backups of user databases occur once per day.  The schedule can be modified by an administrator by updating the corresponding row in tasks.schedules table in the rdpg database on the Service Cluster the user database is on.

When the backup tasks for a particular database is performed, a history of this is recorded in the backups.file_history table in the rdpg database.  The following information is recorded:
 - Database Name
 - Cluster ID of the database backup
 - Status of whether the backup was successfully created
 - Timestamp of when the backup started
 - Duration in seconds the backup took
 - The location of the backup file, both the node the file is on and the absolute path to the file.

## Backup Files

There are two types of backup files for each database:
 - *schema* - Contains all the commands needed to recreate all the schemas, tables, functions, views and extensions to recreate the structure of a database.  Note that it does not include CREATE/DROP DATABASE statements nor does it contain users or roles. The equivalent command line used to create the file is:
 ```bash
 database=d087628179e054079b701a95f63297467
 PGDUMP="/var/vcap/packages/pgbdr/bin/pg_dump -p7432 -U vcap"
 $PGDUMP -c -s -N "bdr" $database > x.schema
 ```
 - *data* - A series of COPY statements which will reinsert rows into tables created by the `schema` file.
 ```bash
 database=d087628179e054079b701a95f63297467
 PGDUMP="/var/vcap/packages/pgbdr/bin/pg_dump -p7432 -U vcap"
 $PGDUMP -a -N "bdr" $database > x.data
 ```
There is a third type of backup which is only performed when the `rdpg` database is backed up called `globals` which contain all the user roles.
```bash
database=rdpg
PGDUMP="/var/vcap/packages/pgbdr/bin/pg_dumpall -p7432 -U vcap"
$PGDUMP --globals-only > x.globals
```

## Location and Naming Conventions

There is an entry in rdpg.config which controls the relative location of the backups.  As of this writing backups are written to `/var/vcap/store/pgbdr/backups/<database_name>/YYMMDDHHMMSS.{schema,file}` on a node.

Whichever node in the Service Cluster is not the write master will perform the `pg_dump` locally and will only be written to that server.  If the write master role switches between the servers you will see different backup files on both servers in the Service Cluster.  See the section `Find Backup Files` to determine which server has the most recent backup file for a particular database.

## Performing Restores
Restores are currently a manual process.  There is a backlog of stories to schedule restores using dashboards but for now an administrator will need to ssh onto one of the nodes in the Service Cluster which contains the user database that needs to be restored.

### Find Backup Files
The rdpg database on the service cluster the user database resides on contains a table called `backups.file_history`.  From this table select the record you would like, it will contain the node and file location the backup file is on.

### Use psql to Perform Restore
You can use the `psql` command to restore the database.

The `DROP EXTENSION` command in the *schema* file will often fail if they are not done in the correct order.  Since the bootstrap process creates databases with the extensions before the user are given access to the database, the extensions do not need to be dropped.

A `sed` command is used to remove the `DROP EXTENSION` statements without modifying the original backup file.

```bash
PSQL="/var/vcap/packages/pgbdr/bin/psql -p7432 -U vcap"
sed "s/DROP EXTENSION/--DROP EXTENSION/g" x.schema > x.schema.exclude_extensions
$PSQL $database < x.schema.exclude_extensions
$PSQL $database < x.data
```

## History Maintenance

Information kept in backup.file_history is only kept for a default time period defined in `rdpg.config`.  A scheduled task called `blarg` is run every hour to delete rows in the table for data retention.

## Backing up Administrative Components

The `rdpg` database is scheduled for backup on all clusters and is backed up to the same location as user database backups.  The data and schema dumps are done identically to the user backups but there is a third file called `YYMMDDHHMMSS.globals` which runs the equivalent of a `pg_dumpall --globals` so all of the user roles are also available to help recover a cluster if it is lost.

## S3 File Copy

Backup files which have been successfully generated to the file system can be copied to AWS S3 storage with the correct configuration in the BOSH deployment manifest and scheduled tasks in the RDPG daemon.

There are 3 pieces of information which are needed from AWS:
 - An AWS Access Key
 - An AWS Secret Key
 - A unique bucket name. Note that you need to create the bucket through the AWS Console, API or other similar tool.  The RDPG daemon will not create the bucket or manage permissions.

To enable the copying of backup files to S3 the BOSH deployment manifest needs to be modified.

```yml
rdpgd_service:
  ...
  backups_s3_access_key: "AKIAJWSZHUQ3HYHMK88B"
  backups_s3_secret_key: "cIK9Y1IAlr3adyChang3dTh1sSoN1c3TryBMQ8"
  backups_s3_bucket_name: "some_unique_bucket_name"
  backups_s3: "ENABLED"
```
This configuration is per cluster, swap `rdpgd_service` with `rdpgd_manager` in the above example to copy backups of the rdpg database for the management cluster.

If you are using an S3 location other than `us-east-1` also add the following to the deployment manifest:
```yml
  backups_s3_region: "us-west-2"
```

To have the task scheduler schedule the file copy task make sure that `enabled` is set to `true` for both tasks `FindFilesToCopyToS3` for node_type `read` and `write`.

## Data Retention

There are two default policies which are enabled by default for the retention of database backups.  BOSH has a limitation of mounting a single disk so the database data and log files as well as the backups all consume the same disk leaving the potential for too many backups using the available disk space.

The ad-hoc task `DeleteFile` is responsible for the removal of files.

The job `DeleteBackupHistory` keeps only 180 days of history so even if the files are copied to S3 you will lose the meta information after this time period.

### Retention Policy 1

If S3 Copy Backups are **enabled** and the backup files are copied to S3 they can be removed from the local file system.

Backup files are deleted if the following criteria are met:
 - The scheduled task `EnforceFileRetention` is enabled for node_type `read` and `write`
 - A backup file was created and logged in `backups.file_history` for the activity `CreateBackup` and status `ok`
 - The file was copied a S3 bucket and logged in `backups.file_history` for the activity `CopyToS3` and status `ok`

When the deletion from the local file system occurs the `removed_at` column for the original `CreateBackup` is updated.

### Retention Policy 2

If S3 Copy Backups are **disabled** keep the last 48 hours worth of successful backups for each database.

Backup files are deleted if the following criteria are met:
 - The scheduled task `EnforceFileRetention` is enabled for node_type `read` and `write`
 - A backup file was created and logged in `backups.file_history` for the activity `CreateBackup` and status `ok`
 - The record was created more than 48 hours ago.

When the deletion from the local file system occurs the `removed_at` column for the original `CreateBackup` is updated.



## Future Improvements

### File Storage
 - Compression of backup files

### Data Retention Policies
 - Define how many backups to keep on local disk
 - Define the age of files to keep in S3 storage

### Adminstrator Dashboards
 - Administrators can view all backup files available
 - Ability to view and modify configuration defaults
 - Restoration of a database to a different cluster

### User Dashboards
 - Users can specify alternate backup frequencies
 - Users can create an ad-hoc backup
 - Users can view a list of available database restore files
 - Users can request a particular backup file to use to preform a restore
 - Users can schedule a restore
 - Users can download a backup file

### End Game
 - Complete Service Cluster Migration - Allow an adminstrator to create a new Service Cluster and use information in S3 to restore rdpg, recreate all the user databases and then restore all user databases to their most recent backup files.
