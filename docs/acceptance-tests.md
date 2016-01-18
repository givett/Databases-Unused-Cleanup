# Acceptance Tests

## Overview

Acceptance tests help to assure that recent changes result in rdpg clusters are in a good state.  They are also used by the Concourse build pipeline to create releases for deployments to other deployment pipelines.  The tests are written using the Ginkgo spec tool and thus the order of execution will be random.

## Running Acceptance Tests

Make any changes you would like to `rdpg-boshrelease/src/rdpg-acceptance-tests`.  From the root folder of the `rdpg-boshrelease` run the following:

```bash
bosh create release --force && bosh upload release
bosh -n deploy
bosh run errand acceptance_tests
```

Assuming the errand ran successfully, the tail of the output displayed by the bosh errand should look something like:

```
Errand `acceptance_tests' completed successfully (exit code 0)
```

If the errand wasn't successful, scroll up in the output and start reviewing to see which tests failed.

## Current Tests

The tests are defined in `rdpg-boshrelease/src/rdpg-acceptance-tests/rdpg-service/*` folders with each folder representing logical groups of tests:


Directory | Description
-------------------- | --------------------
`broker/` | Contains Service Broker availability tests
`postgres/` | Validates schemas and tables in the `rdpg` database on each node
`consul/` | Checks warden deployments that the correct number of nodes are available and all services are running
`backups/` | Checks if user database backups and other default backup tasks are scheduled
`user-databases/` | Validates pre-existing user databases
`migrations/` | This maybe removed in future when the two legacy environments of RDPG are upgraded

The checks for each of these directories are detailed below.

### broker/ Checks

Check Name | Description
-------------------- | --------------------
**Check Basic Auth No Creds** | Prompts for Basic Auth creds when they aren't provided
**Check Bad Basic Auth Creds** | Does not accept bad Basic Auth creds
**Check Valid Basic Auth Creds** | Accepts valid Basic Auth creds

### postgres/ Checks

Check Name | Description
-------------------- | --------------------
**Check Schemas Exist** | This test validates that all nodes in all clusters have the following schemas created in the rdpg database: `bdr, rdpg, cfsb, tasks, backups, metrics, audit`. If any are missing, the bootstrapping process for the rdpg daemon was not successful.  Look at the logs at `/var/vcap/sys/log/rdpgd-{manager,service}/rdpg-{manager,service}.log` and search for errors.  These schemas are created during the bootstrap.
**Check cfsb Tables Exist** | This test validates that `cfsb.services`, `cfsb.plans`, `cfsb.instances`, `cfsb.bindings`, and `cfsb.credentials` tables exist in the rdgp database for every management and service cluster node.  These tables are created during the bootstrap.
**Check rdpg Tables Exist** | This test validates that `rdpg.confi`g, `rdpg.consul_watch_notifications` and `rdpg.events` tables exist in the rdgp database for every management and service cluster node.  These tables are created during the bootstrap.
**Check tasks Tables Exist** | This test validates that `tasks.tasks` and `tasks.schedules` tables exist on every management and service cluster node in the rdpg database.  These tables are created during the bootstrap.
**Check Instance Counts** | Every service cluster is supposed to pre-allocate bdr replicated databases and report the existence of these databases to the management cluster.  Meta information about each of thee databases is stored in the rdpg database in `cfsb.instances`. This check validates that each service cluster has created a default minimum (20) of user databases *(hint: user databases names all start with 'd')*, that all nodes in a cluster have the same number of databases and finally that the management cluster matches the sum of all the service clusters' available user databases. Note that when this test is run immediately following a new deployment it may fail the test until all of the databases have been created the first time.  Wait a few minutes and run the test again and only then if the failure persists should you be worried.
**Check Scheduled Tasks Exist** | Tests that each service cluster has at least 3 scheduled tasks (the default) with a role of All or Service and that each cluster in the node has the same number of active scheduled tasks.  For the management cluster there are at least 4 default tasks and the count is compared across all nodes in the cluster.  The scheduled tasks are inserted into the tasks.schedules table in the rdgp database during bootstrap.
**Check for Missed Scheduled Tasks** | Checks for any enabled task which is eligible to be scheduled has been skipped for more than twice the duration.  This validates jobs are being rescheduled correctly and are firing.
**Check for databases known to cfsb.instances but don't exist** | These are databases which do not currently or never have existed within postgres.  If all nodes in a service cluster report that a particular database fails to exist either the database was deleted and the entry in cfsb.instances was not updated correctly to denote it's retirement or an unknown bug during the initial database creation. If at least one node in the service cluster has the database created but the other nodes did not, something failed with the bdr database join function in the "PrecreateDatabases" scheduled task for that service cluster.
**Check for databases which exist and aren't known to cfsb.instances** | These are databases which aren't being managed by the rdpg daemon (but likely should be).  This could be the result of the database being restored manually from another service cluster but not registered with cfsb.instances.  Every effort should be made to determine if the database is a user database and added back to cfsb.instances so that scheduled database maintenance can be performed on it (including backups).

### consul/ Checks

Check Name | Description
-------------------- | --------------------
**Check Node Counts** | For deployments of rdpg to warden, the deployment manifest defines there to be 1 management cluster with 3 nodes, and two service clusters each with 2 nodes.  When consul is queried for services matching "rdpgmc, rdpgsc1, rdpgsc2" the number of nodes returned are compared against the good known values and also validates that a connection to consul can be made.
**Check Data Center Name** | The host name should be matched.
**Check Leader** | The leader should be selected and each node should be able to request leader information.
**Check Peers** | All the nodes should be able to get information for all peers.
**Health check for all the services on each node** | Both management cluster and service cluster nodes should have the corresponding services registered and running.

### backups/ Checks
Check Name | Description
-------------------- | --------------------
**Check backups Tables Exist** | Verifies that there is a backups.file_history table in the rdpg database.  This table stores the history of when and what databases were backed up.  The table is created during boostrap.
**Check all user databases are scheduled for backups** | This verifies that a backup for each user database has been scheduled in tasks.schedules in the rdpg database.  This test *may* fail if it is run against a newly created deployment of rdpg which has not had a chance for the scheduled task `ScheduleNewDatabaseBackups` to run.  There is up to a 1 minute lag between when a database is registered in cfsb.instances and when the backup is scheduled.
**Check all configuration defaults have been configured** | There are 4 key/value pairs in rdpg.config in the rdpg database which are used for backups.  If the keys `pgDumpBinaryLocation`,`BackupPort`,`BackupsPath`,`defaultDaysToKeepFileHistory` are missing for a service cluster the backups will fail to be created.  The key/value pairs are inserted during bootstrap.
**Check task DeleteBackupHistory is scheduled** | Validates that there is an entry in tasks.schedules in the rdpg database for `DeleteBackupHistory`.  This task is responsible for removing rows from the backups.file_history table which have exceeded the data retention duration defined as 180 days in the rdpg.config table.  The scheduled task is inserted during boostrap.
**Check task ScheduleNewDatabaseBackups is scheduled** | Validates that there is an entry in tasks.schedules in the rdpg database for `ScheduleNewDatabaseBackups`.  This task is responsible adding scheduled tasks to perform backups for each user database known to cfsb.instances.  The scheduled task is inserted during boostrap.
**Check backups.file_history truncation is working** | This verifies that the scheduled task `DeleteBackupHistory` is working as expected and there are no rows of data in backups.file_history which exceed the data retention default defined in rdgp.config.
**Check task BackupDatabase for rdpg system database is scheduled** | This verifies that the `rdpg` database is scheduled for a backup on each cluster.

### migrations/ Checks

These are not permanent and will be removed once the two target environments with RDPG deployed are upgraded.  A better solution for this will be created in the future.

Check Name | Description
-------------------- | --------------------
**Check backups.file_history table exists, otherwise create** | For new deployments this table is created during bootstrap.
**Check node_type column in tasks.tasks table exists, otherwise create** | For new deployments this column is included already with the TABLE CREATE for `tasks.tasks` and the table is created during bootstrap.
**Check node_type column in tasks.schedules table exists, otherwise create** | For new deployments this column is included already with the TABLE CREATE for `tasks.schedules` and the table is created during bootstrap.
**Check default for defaultDaysToKeepFileHistory added rdpg.config** | For new deployments this key/value pair is inserted into rdpg.config during bootstrap.
 **Check default for BackupPort added rdpg.config** | For new deployments this key/value pair is inserted into rdpg.config during bootstrap.
**Check default for BackupsPath added rdpg.config** | For new deployments this key/value pair is inserted into rdpg.config during bootstrap.
**Check default for pgDumpBinaryLocation added rdpg.config** | For new deployments this key/value pair is inserted into rdpg.config during bootstrap.
**Check BackupDatabase job exists in tasks.schedules for the rdpg system database** | For existing deployments this adds a backup task for the `rdpg` database

## user-databases/ Checks

This class of tests is used to validate information about each user database.

Check Name | Description
-------------------- | --------------------
**Check all user databases have bdr pairs** | Each user database should have two entries in the `bdr.bdr_nodes` table, if not you will receive `DDL Lock` errors which prevent users from making DDL changes.
