# Update Existing Deployment

## Update Database

Update the name of the service in the `rdpg` database on one of the Management Cluster nodes.

Connect to PostgreSQL:
```bash
PSQL="/var/vcap/packages/pgbdr/bin/psql -p7432 -U vcap"
$PSQL rdpg
```

```SQL
UPDATE cfsb.services SET name = 'whatever new service name you want';
```

## Update CF Service Broker

Update the existing plan name
```
cf update-service-broker rdpg cfadmin cfadmin http://<ip address of sb>:8888  #10.202.72.32 on R5, run `cf m`
```

Activate the new `postgres` plan
```
cf enable-service-access postgres
```

Show the list of existing plans
```
cf service-access
```
