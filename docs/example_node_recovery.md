# Copy BDR Broken Databases from SC2 to SC3 in R2
The following example shows how to move databases from one Service Cluster to another Service Cluster managed by the same Management Cluster.  In this scenario Service Cluster 2 (SC2) has one of the two nodes with a corrupted PostgreSQL install which has resulted in broken BDR replication.  The good node is still responsive but unable to perform DDL changes.

## Get db backups working, run on SC2/0

```
DBS=(
d97ae77e47f1240eda971e3b627e3ba83
d6bf9c759043f4b8b943a6af8a74b0d7c
df183b04c0f944e73b9ba3cfdddfba863
da617481ff55449188c3d4f72a80df61d
d4385c93ff0744b1c9b9a979ab7110821
d8d58910939244064a3f27937f2c12b1e
d7de4a32de58740e1aac8cff449f22a45
d06da85abb7c14e959108da8cbfa2415a
dac77084be7b2431f8674c8d5b0a2dccb
d6714a3740a5742a18b1bdc9a25ccbaa1
d272b0fbce69c4af5909dde67b36038f2
dcc6db01069524d9b82e5644f36c718c8
d45dee6c0c56f41d58df3f1224921e044
d8082e6ec290c4322a804f7d9aca26c73
df7fe15bb18df4b219e83cbfcd559e00a
d799f7cafbe274f56ba0933753c5ad2cd
d3ddf4382a2ae47288e0c5dedc1aef166
d26cd6c10fdd64ecbb345ab354a3fb5b8
dc241d23d45d34c06857cb3982e026b89
d393605fe8fbd4339b0454efc4fd24fb2
d2c4cf043d29e464cabaa44808d801687
dd87fb183ca12421488b49582df82f63b
de30f4b3406f045b88a9dde474db057db
ddc9dea0822324dd79e87c4fbf8f552bd
d085059bd4b0841be9df4008012070be1
d4c8ac9e2a1a148fcb78997c03c4838f7
d6c4297de738b4f1086e9ea8454b284ba
d44423b7bc7ed438ab68c373cedde5c44
)
--removed d089ae9d9ac1a4f52acbc8b4288db5835, only had 1 table called `test`


for dbname in ${DBS[@]}; do
  mkdir -p /var/vcap/store/pgbdr/recover/
  PGDUMP="/var/vcap/packages/pgbdr/bin/pg_dump -p7432 -U vcap"
  echo $dbname
  $PGDUMP -c -s -N "bdr" $dbname > /var/vcap/store/pgbdr/recover/${dbname}.final.schema
  $PGDUMP -a  -N "bdr" $dbname > /var/vcap/store/pgbdr/recover/${dbname}.final.data
  sed "s/DROP EXTENSION/--DROP EXTENSION/g" /var/vcap/store/pgbdr/recover/${dbname}.final.schema > /var/vcap/store/pgbdr/recover/${dbname}.final.schema.exclude_extensions
done

#compress files
tar -cvzf /var/vcap/store/pgbdr/final1.tgz /var/vcap/store/pgbdr/recover/
```

## Copy Backups to SC3/0 using netcat
sending box sc2/0:
`cat /var/vcap/store/pgbdr/final1.tgz | nc 10.202.84.38 54444`

Receiving box sc3/0:
`nc -w 30 -l 54444 > /var/vcap/store/pgbdr/final1.tgz`

Back on SC3 unzip
```
tar xvzf /var/vcap/store/pgbdr/final1.tgz  -C /var/vcap/store/pgbdr/
```

# Switch to SC3/0

```
PSQL="/var/vcap/packages/pgbdr/bin/psql -p7432 -U vcap"
```

## Get list of datatabase

```
dbname="d97ae77e47f1240eda971e3b627e3ba83"; dbuser="u97ae77e47f1240eda971e3b627e3ba83"; dbpass="6a0234ee622c462d9627825ebec89399"
dbname="d6bf9c759043f4b8b943a6af8a74b0d7c"; dbuser="u6bf9c759043f4b8b943a6af8a74b0d7c"; dbpass="6a686da4d59d4dc385e045660bc09365"
dbname="df183b04c0f944e73b9ba3cfdddfba863"; dbuser="uf183b04c0f944e73b9ba3cfdddfba863"; dbpass="c834f139145849978f1cfaa3448b0a7e"
dbname="da617481ff55449188c3d4f72a80df61d"; dbuser="ua617481ff55449188c3d4f72a80df61d"; dbpass="4a5777cf555a46a6a219e869e1d9e8c5"
dbname="d4385c93ff0744b1c9b9a979ab7110821"; dbuser="u4385c93ff0744b1c9b9a979ab7110821"; dbpass="b8409d1a1afe465a8bf595ea13492e74"
dbname="d8d58910939244064a3f27937f2c12b1e"; dbuser="u8d58910939244064a3f27937f2c12b1e"; dbpass="6276e1f231464fffa0fcc9629e8d03ad"
dbname="d7de4a32de58740e1aac8cff449f22a45"; dbuser="u7de4a32de58740e1aac8cff449f22a45"; dbpass="872ab3b63d614736943e24e326f393c3"
dbname="d06da85abb7c14e959108da8cbfa2415a"; dbuser="u06da85abb7c14e959108da8cbfa2415a"; dbpass="40b4430983144d239482b277feffc61e"
dbname="dac77084be7b2431f8674c8d5b0a2dccb"; dbuser="uac77084be7b2431f8674c8d5b0a2dccb"; dbpass="bfae1516dd0c43c484f8e1d27e4fa967"
dbname="d6714a3740a5742a18b1bdc9a25ccbaa1"; dbuser="u6714a3740a5742a18b1bdc9a25ccbaa1"; dbpass="3fc154bd060d4f7c9ef80093f17bb721"
dbname="d272b0fbce69c4af5909dde67b36038f2"; dbuser="u272b0fbce69c4af5909dde67b36038f2"; dbpass="b99fb1fe38e24088a3327b0064ce04c2"
dbname="dcc6db01069524d9b82e5644f36c718c8"; dbuser="ucc6db01069524d9b82e5644f36c718c8"; dbpass="435f2787bd614a86b74c04b7b2ab13fe"
dbname="d45dee6c0c56f41d58df3f1224921e044"; dbuser="u45dee6c0c56f41d58df3f1224921e044"; dbpass="099c3a8527614bfba986601529619a82"
dbname="d8082e6ec290c4322a804f7d9aca26c73"; dbuser="u8082e6ec290c4322a804f7d9aca26c73"; dbpass="d7f0162fead04f9eaf9a4df9b9be6326"
dbname="df7fe15bb18df4b219e83cbfcd559e00a"; dbuser="uf7fe15bb18df4b219e83cbfcd559e00a"; dbpass="225b2e539aec4551a2aec7625cbe8d7a"
dbname="d799f7cafbe274f56ba0933753c5ad2cd"; dbuser="u799f7cafbe274f56ba0933753c5ad2cd"; dbpass="6cd5835928294661b7bd25ce7d5d2c25"
dbname="d089ae9d9ac1a4f52acbc8b4288db5835"; dbuser="u089ae9d9ac1a4f52acbc8b4288db5835"; dbpass="35c6fe9fa149462484c4eb63cb2fcc2e"
dbname="d3ddf4382a2ae47288e0c5dedc1aef166"; dbuser="u3ddf4382a2ae47288e0c5dedc1aef166"; dbpass="7cf09ee29a644a0a882dd3b4a1f772d3"
dbname="d26cd6c10fdd64ecbb345ab354a3fb5b8"; dbuser="u26cd6c10fdd64ecbb345ab354a3fb5b8"; dbpass="cbadc4b9799d41c1969e3e68d3fa5e06"
dbname="dc241d23d45d34c06857cb3982e026b89"; dbuser="uc241d23d45d34c06857cb3982e026b89"; dbpass="010e205e3e6b4d5aa8c3bba4b4e8e770"
dbname="d393605fe8fbd4339b0454efc4fd24fb2"; dbuser="u393605fe8fbd4339b0454efc4fd24fb2"; dbpass="933cdbefaccc4b41bb7be2b88b3b17fa"
dbname="d2c4cf043d29e464cabaa44808d801687"; dbuser="u2c4cf043d29e464cabaa44808d801687"; dbpass="01cabac6f9d140edb4d0214da2cba4a3"
dbname="dd87fb183ca12421488b49582df82f63b"; dbuser="ud87fb183ca12421488b49582df82f63b"; dbpass="0f80e671f30f4168b6c90b7e52b17a43"
dbname="de30f4b3406f045b88a9dde474db057db"; dbuser="ue30f4b3406f045b88a9dde474db057db"; dbpass="32667429cd84457296bb302eba4d7d8f"
dbname="ddc9dea0822324dd79e87c4fbf8f552bd"; dbuser="udc9dea0822324dd79e87c4fbf8f552bd"; dbpass="0ae8755d79fb4896bae37150ca6444d4"
dbname="d085059bd4b0841be9df4008012070be1"; dbuser="u085059bd4b0841be9df4008012070be1"; dbpass="27a82f6d19a9474ea111724e8d7ecf97"
dbname="d4c8ac9e2a1a148fcb78997c03c4838f7"; dbuser="u4c8ac9e2a1a148fcb78997c03c4838f7"; dbpass="4e00b29a4f3449f5af12745589aeed6d"
dbname="d6c4297de738b4f1086e9ea8454b284ba"; dbuser="u6c4297de738b4f1086e9ea8454b284ba"; dbpass="e48b81cf650c4f18870bc344277be192"
dbname="d44423b7bc7ed438ab68c373cedde5c44"; dbuser="u44423b7bc7ed438ab68c373cedde5c44"; dbpass="50b13baf86254049ada72c6a7701dff4"
```

## Create Database, User and BDR

Loop through the results from the previous step, run the following on SC3/0 and SC3/1:

```
dbname="d44423b7bc7ed438ab68c373cedde5c44"; dbuser="u44423b7bc7ed438ab68c373cedde5c44"; dbpass="50b13baf86254049ada72c6a7701dff4"

$PSQL postgres -c "CREATE USER ${dbuser};"  #On first attempt, the user was already successful, may be able to skip these two steps
$PSQL postgres -c "ALTER USER ${dbuser} ENCRYPTED PASSWORD '"${dbpass}"';"
$PSQL postgres -c "CREATE DATABASE ${dbname} WITH OWNER ${dbuser} TEMPLATE template0 ENCODING 'UTF8';"
$PSQL postgres -c "REVOKE ALL ON DATABASE \"${dbname}\" FROM public;"
$PSQL postgres -c "GRANT ALL PRIVILEGES ON DATABASE ${dbname} TO ${dbuser};"
$PSQL $dbname -c "CREATE EXTENSION IF NOT EXISTS btree_gist;"
$PSQL $dbname -c "CREATE EXTENSION IF NOT EXISTS bdr;"
$PSQL $dbname -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;"
$PSQL $dbname -c "CREATE EXTENSION IF NOT EXISTS pg_stat_statements;"
$PSQL $dbname -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
$PSQL $dbname -c "CREATE EXTENSION IF NOT EXISTS hstore;"
nodeNameSideA="${dbname}_a"
nodeNameSideB="${dbname}_b"
port="7432"
myIPSideA="10.202.84.38"
targetIPSideA="10.202.84.39"
repUser="postgres"

```

### Now peform group create & join
Execute one of the two statements below, run the first one on the first node, the second one on the second node

On first node:

```
$PSQL $dbname -c "SELECT bdr.bdr_group_create( local_node_name := '"${nodeNameSideA}"', node_external_dsn := 'host="${myIPSideA}" port="${port}" user="${repUser}" dbname="${dbname}"'); "
$PSQL $dbname -c "SELECT bdr.bdr_node_join_wait_for_ready();"

```

On second node:

```
$PSQL $dbname -c "SELECT bdr.bdr_group_join( local_node_name := '"${nodeNameSideB}"', node_external_dsn := 'host="${targetIPSideA}" port="${port}" user="${repUser}" dbname="${dbname}"', join_using_dsn := 'host=${myIPSideA} port="${port}" user="${repUser}" dbname="$dbname"'); "
$PSQL $dbname -c "SELECT bdr.bdr_node_join_wait_for_ready();"

```



## Once db is up on both nodes then run the following on SC3/0:

### Restore DATABASE

```
$PSQL $dbname < /var/vcap/store/pgbdr/var/vcap/store/pgbdr/recover/${dbname}.final.schema.exclude_extensions
$PSQL $dbname < /var/vcap/store/pgbdr/var/vcap/store/pgbdr/recover/${dbname}.final.data

```

### Register with cfsb.instances in the Service Cluster SC3 & Management Cluster

```
PSQL="/var/vcap/packages/pgbdr/bin/psql -p7432 -U vcap"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc2','abd06ac6-8c2b-4ca5-ad57-fc2a1a54110c','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','6f47470f-134c-42f6-8279-5362421497c5','1f203c4d-2719-4e12-97ab-5508fde6fc31','d97ae77e47f1240eda971e3b627e3ba83','u97ae77e47f1240eda971e3b627e3ba83','6a0234ee622c462d9627825ebec89399');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','e0f88c1b-18e3-40b2-b5be-e414cdb6e2d6','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','e2cf1ec8-df85-45f9-b737-767c599c245d','85e6561e-b8fd-4307-a7e8-fa1b7f6bf1f8','d6bf9c759043f4b8b943a6af8a74b0d7c','u6bf9c759043f4b8b943a6af8a74b0d7c','6a686da4d59d4dc385e045660bc09365');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','62e1e9d1-0ff0-407a-97b5-6499ccdb92d1','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','6f47470f-134c-42f6-8279-5362421497c5','1f203c4d-2719-4e12-97ab-5508fde6fc31','df183b04c0f944e73b9ba3cfdddfba863','uf183b04c0f944e73b9ba3cfdddfba863','c834f139145849978f1cfaa3448b0a7e');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','3055ebcd-3b16-4a11-a28a-ba0c6e258ce3','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','6f47470f-134c-42f6-8279-5362421497c5','1f203c4d-2719-4e12-97ab-5508fde6fc31','da617481ff55449188c3d4f72a80df61d','ua617481ff55449188c3d4f72a80df61d','4a5777cf555a46a6a219e869e1d9e8c5');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','63873feb-449c-4d0c-bb16-cdc26c02a7c7','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','27c2bb71-f1b9-48d1-afe3-96b25d68bb52','1e5b5b48-89dd-4481-9cb7-0694bbdb25ab','d4385c93ff0744b1c9b9a979ab7110821','u4385c93ff0744b1c9b9a979ab7110821','b8409d1a1afe465a8bf595ea13492e74');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','309b68e6-8e5a-42c6-9719-371665673ec3','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','1c95e6bb-d3cc-451c-9567-33d6c7d391af','4b98d204-8b7e-4353-ab8b-907960f8c65e','d8d58910939244064a3f27937f2c12b1e','u8d58910939244064a3f27937f2c12b1e','6276e1f231464fffa0fcc9629e8d03ad');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','7544c667-b932-479f-a2b2-57013f1051c1','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','70d30324-d273-4073-bb84-f6f91ab376e2','bbd19255-3258-462a-81e6-4c5fbb7101aa','d7de4a32de58740e1aac8cff449f22a45','u7de4a32de58740e1aac8cff449f22a45','872ab3b63d614736943e24e326f393c3');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','82d055df-a494-4b7d-b538-9844d921c70e','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','9112ed1a-e5c3-402f-9586-b67788f8a742','f4f7e31c-4f98-4657-9fe0-3be54adbe7a1','d06da85abb7c14e959108da8cbfa2415a','u06da85abb7c14e959108da8cbfa2415a','40b4430983144d239482b277feffc61e');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','6582fed5-8a04-49cb-accd-4368322d7061','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','6f47470f-134c-42f6-8279-5362421497c5','1f203c4d-2719-4e12-97ab-5508fde6fc31','dac77084be7b2431f8674c8d5b0a2dccb','uac77084be7b2431f8674c8d5b0a2dccb','bfae1516dd0c43c484f8e1d27e4fa967');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','bd5f31a8-c4b9-4e0a-b459-a66fe965defb','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','27c2bb71-f1b9-48d1-afe3-96b25d68bb52','1e5b5b48-89dd-4481-9cb7-0694bbdb25ab','d6714a3740a5742a18b1bdc9a25ccbaa1','u6714a3740a5742a18b1bdc9a25ccbaa1','3fc154bd060d4f7c9ef80093f17bb721');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','69b04549-21c6-42df-8952-bb0ec1089ebd','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','8aae213a-30ff-4ea6-962f-5ff2f733acc3','8ae3577b-5c4f-4c30-9885-a47fd75ccefb','d272b0fbce69c4af5909dde67b36038f2','u272b0fbce69c4af5909dde67b36038f2','b99fb1fe38e24088a3327b0064ce04c2');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','f28d19b1-db81-42f4-9bf1-90aad80a66c8','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','0bd74d06-b2ec-4656-9702-19528af00b3f','a5b88cbd-bb23-441a-9fbd-fbf6fbd8bd9d','dcc6db01069524d9b82e5644f36c718c8','ucc6db01069524d9b82e5644f36c718c8','435f2787bd614a86b74c04b7b2ab13fe');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','3492002a-10bd-4dcc-9333-f9ab7c9335a7','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','1c95e6bb-d3cc-451c-9567-33d6c7d391af','2506c93b-73ac-447d-956f-bf6400adfae5','d45dee6c0c56f41d58df3f1224921e044','u45dee6c0c56f41d58df3f1224921e044','099c3a8527614bfba986601529619a82');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','72c2b542-a72f-4349-abba-19c280034e69','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','9a061ac2-3660-485e-b4c5-97e0b0a1cde5','d21f7746-e2b1-428e-b738-12686fea69e2','d8082e6ec290c4322a804f7d9aca26c73','u8082e6ec290c4322a804f7d9aca26c73','d7f0162fead04f9eaf9a4df9b9be6326');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','2a9baf94-3a44-4402-8008-5f12af53937f','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','e2cf1ec8-df85-45f9-b737-767c599c245d','3624cede-4959-480a-8f90-c3da03e0c50d','df7fe15bb18df4b219e83cbfcd559e00a','uf7fe15bb18df4b219e83cbfcd559e00a','225b2e539aec4551a2aec7625cbe8d7a');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','833d5561-bbdc-4d5d-8692-2214ef59fdb0','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','9112ed1a-e5c3-402f-9586-b67788f8a742','f4f7e31c-4f98-4657-9fe0-3be54adbe7a1','d799f7cafbe274f56ba0933753c5ad2cd','u799f7cafbe274f56ba0933753c5ad2cd','6cd5835928294661b7bd25ce7d5d2c25');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','03899d88-9acb-4e24-8077-b918ba60ecdb','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','e2cf1ec8-df85-45f9-b737-767c599c245d','7358ba72-39b5-4eca-bc01-7924b99ba43b','d089ae9d9ac1a4f52acbc8b4288db5835','u089ae9d9ac1a4f52acbc8b4288db5835','35c6fe9fa149462484c4eb63cb2fcc2e');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','06375a58-6fd4-4431-88cc-2cde69cac2de','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','1c95e6bb-d3cc-451c-9567-33d6c7d391af','253e5387-b4d4-4872-a65e-98124dc68c40','d3ddf4382a2ae47288e0c5dedc1aef166','u3ddf4382a2ae47288e0c5dedc1aef166','7cf09ee29a644a0a882dd3b4a1f772d3');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','74b32384-c341-44bd-9427-86bbe3eb0d0b','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','1c95e6bb-d3cc-451c-9567-33d6c7d391af','253e5387-b4d4-4872-a65e-98124dc68c40','d26cd6c10fdd64ecbb345ab354a3fb5b8','u26cd6c10fdd64ecbb345ab354a3fb5b8','cbadc4b9799d41c1969e3e68d3fa5e06');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','128aaa64-164a-440d-bbdf-ab6b87a1a943','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','2c8a3037-7d8e-4ed1-afc2-6d574df111c2','a365f5f9-cead-4a88-86f0-992a227d44e0','dc241d23d45d34c06857cb3982e026b89','uc241d23d45d34c06857cb3982e026b89','010e205e3e6b4d5aa8c3bba4b4e8e770');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','da0277a0-ad21-484d-be4b-e38e6325262b','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','1a4966df-78fb-4196-9dda-240911e93b3c','86ba4332-ec62-4e8f-b5c5-7169b44dcb12','d393605fe8fbd4339b0454efc4fd24fb2','u393605fe8fbd4339b0454efc4fd24fb2','933cdbefaccc4b41bb7be2b88b3b17fa');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','591edcf0-a2a0-45bb-b8d9-db2723e1122b','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','9c7f3f7f-4924-4bb7-a543-dfee544aa2b2','e7cedfaf-c97b-4e53-8042-f3a41624317e','d2c4cf043d29e464cabaa44808d801687','u2c4cf043d29e464cabaa44808d801687','01cabac6f9d140edb4d0214da2cba4a3');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','11e7580d-8bb7-4fbd-8e4c-9964a824b322','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','2c8a3037-7d8e-4ed1-afc2-6d574df111c2','6f68fd43-3f8a-4596-bb1d-fb1c7652e99d','dd87fb183ca12421488b49582df82f63b','ud87fb183ca12421488b49582df82f63b','0f80e671f30f4168b6c90b7e52b17a43');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','ffb1d08f-dae3-4faf-bd2b-2f89589cfd87','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','2c8a3037-7d8e-4ed1-afc2-6d574df111c2','86672290-2288-4dc0-bfaa-00ec33f018a7','de30f4b3406f045b88a9dde474db057db','ue30f4b3406f045b88a9dde474db057db','32667429cd84457296bb302eba4d7d8f');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','d167c3c5-32bd-4275-8f60-70101490bbb6','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','2c8a3037-7d8e-4ed1-afc2-6d574df111c2','d71e2bb9-50ed-4f75-b3fa-cbdd5a38c468','ddc9dea0822324dd79e87c4fbf8f552bd','udc9dea0822324dd79e87c4fbf8f552bd','0ae8755d79fb4896bae37150ca6444d4');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','bacc90c5-d990-4d17-b01a-305e4f19dd1f','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','2c8a3037-7d8e-4ed1-afc2-6d574df111c2','63e8d661-44d0-4094-9546-84896fe1b601','d085059bd4b0841be9df4008012070be1','u085059bd4b0841be9df4008012070be1','27a82f6d19a9474ea111724e8d7ecf97');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','7dea1c06-4476-4bb3-b60a-5ac0499c6e6b','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','0bd74d06-b2ec-4656-9702-19528af00b3f','a5b88cbd-bb23-441a-9fbd-fbf6fbd8bd9d','d4c8ac9e2a1a148fcb78997c03c4838f7','u4c8ac9e2a1a148fcb78997c03c4838f7','4e00b29a4f3449f5af12745589aeed6d');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','7a4bc526-f02b-449d-9bc3-d96848bd6ba0','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','abe08872-4ca8-4793-9e79-85b82eb69e94','630f90fa-d788-460f-bc96-515093923c83','d6c4297de738b4f1086e9ea8454b284ba','u6c4297de738b4f1086e9ea8454b284ba','e48b81cf650c4f18870bc344277be192');"
$PSQL rdpg -c "INSERT INTO cfsb.instances (cluster_id, instance_id, service_id, plan_id, organization_id, space_id, dbname, dbuser, dbpass) VALUES ( 'rdpgsc3','a1414026-fdad-4d30-bd37-1d264e4f2c8f','738508bd-1f4e-4183-b93b-ea7779b15bde','946ae934-d8ed-4220-a79c-d89b12f7fa3d','2c8a3037-7d8e-4ed1-afc2-6d574df111c2','01d161df-f11c-4fa2-be94-05c6e9bd4e4e','d44423b7bc7ed438ab68c373cedde5c44','u44423b7bc7ed438ab68c373cedde5c44','50b13baf86254049ada72c6a7701dff4');"
```

## Fix up Management Cluster

Connect to a server on the Management Cluster
** Run the above queries if they aren't already sync'd for you **

### Disable `instance_id` references to R2

Verify that we see the 29 databases:
```
SELECT count(*) FROM cfsb.instances WHERE cluster_id = 'rdpgsc2' AND dbname IN ('d97ae77e47f1240eda971e3b627e3ba83','d6bf9c759043f4b8b943a6af8a74b0d7c','df183b04c0f944e73b9ba3cfdddfba863','da617481ff55449188c3d4f72a80df61d','d4385c93ff0744b1c9b9a979ab7110821','d8d58910939244064a3f27937f2c12b1e','d7de4a32de58740e1aac8cff449f22a45','d06da85abb7c14e959108da8cbfa2415a','dac77084be7b2431f8674c8d5b0a2dccb','d6714a3740a5742a18b1bdc9a25ccbaa1','d272b0fbce69c4af5909dde67b36038f2','dcc6db01069524d9b82e5644f36c718c8','d45dee6c0c56f41d58df3f1224921e044','d8082e6ec290c4322a804f7d9aca26c73','df7fe15bb18df4b219e83cbfcd559e00a','d799f7cafbe274f56ba0933753c5ad2cd','d089ae9d9ac1a4f52acbc8b4288db5835','d3ddf4382a2ae47288e0c5dedc1aef166','d26cd6c10fdd64ecbb345ab354a3fb5b8','dc241d23d45d34c06857cb3982e026b89','d393605fe8fbd4339b0454efc4fd24fb2','d2c4cf043d29e464cabaa44808d801687','dd87fb183ca12421488b49582df82f63b','de30f4b3406f045b88a9dde474db057db','ddc9dea0822324dd79e87c4fbf8f552bd','d085059bd4b0841be9df4008012070be1','d4c8ac9e2a1a148fcb78997c03c4838f7','d6c4297de738b4f1086e9ea8454b284ba','d44423b7bc7ed438ab68c373cedde5c44');
```

Now update the existing rows referencing SC2 so we no longer have duplicate:
```SQL
UPDATE cfsb.instances SET instance_id = ('off' || instance_id) WHERE cluster_id = 'rdpgsc2' and dbname in ('d97ae77e47f1240eda971e3b627e3ba83','d6bf9c759043f4b8b943a6af8a74b0d7c','df183b04c0f944e73b9ba3cfdddfba863','da617481ff55449188c3d4f72a80df61d','d4385c93ff0744b1c9b9a979ab7110821','d8d58910939244064a3f27937f2c12b1e','d7de4a32de58740e1aac8cff449f22a45','d06da85abb7c14e959108da8cbfa2415a','dac77084be7b2431f8674c8d5b0a2dccb','d6714a3740a5742a18b1bdc9a25ccbaa1','d272b0fbce69c4af5909dde67b36038f2','dcc6db01069524d9b82e5644f36c718c8','d45dee6c0c56f41d58df3f1224921e044','d8082e6ec290c4322a804f7d9aca26c73','df7fe15bb18df4b219e83cbfcd559e00a','d799f7cafbe274f56ba0933753c5ad2cd','d089ae9d9ac1a4f52acbc8b4288db5835','d3ddf4382a2ae47288e0c5dedc1aef166','d26cd6c10fdd64ecbb345ab354a3fb5b8','dc241d23d45d34c06857cb3982e026b89','d393605fe8fbd4339b0454efc4fd24fb2','d2c4cf043d29e464cabaa44808d801687','dd87fb183ca12421488b49582df82f63b','de30f4b3406f045b88a9dde474db057db','ddc9dea0822324dd79e87c4fbf8f552bd','d085059bd4b0841be9df4008012070be1','d4c8ac9e2a1a148fcb78997c03c4838f7','d6c4297de738b4f1086e9ea8454b284ba','d44423b7bc7ed438ab68c373cedde5c44');
```

## Rebind CF Applications

Login to R2 CF from the Jumpbox
```
cf login -a http://api.grc-apps.svc.ice.ge.com -u admin -p BtIJNGJlVRHqcdgiWFIW
```
### Gather Information
In another terminal session connect to a Management Cluster node to the `instance_id`s from `cfsb.instances` in `rdpg`

```
$PSQL rdpg -c "SELECT instance_id FROM cfsb.instances cluster_id = 'rdpgsc3' and dbname in ('d97ae77e47f1240eda971e3b627e3ba83','d6bf9c759043f4b8b943a6af8a74b0d7c','df183b04c0f944e73b9ba3cfdddfba863','da617481ff55449188c3d4f72a80df61d','d4385c93ff0744b1c9b9a979ab7110821','d8d58910939244064a3f27937f2c12b1e','d7de4a32de58740e1aac8cff449f22a45','d06da85abb7c14e959108da8cbfa2415a','dac77084be7b2431f8674c8d5b0a2dccb','d6714a3740a5742a18b1bdc9a25ccbaa1','d272b0fbce69c4af5909dde67b36038f2','dcc6db01069524d9b82e5644f36c718c8','d45dee6c0c56f41d58df3f1224921e044','d8082e6ec290c4322a804f7d9aca26c73','df7fe15bb18df4b219e83cbfcd559e00a','d799f7cafbe274f56ba0933753c5ad2cd','d089ae9d9ac1a4f52acbc8b4288db5835','d3ddf4382a2ae47288e0c5dedc1aef166','d26cd6c10fdd64ecbb345ab354a3fb5b8','dc241d23d45d34c06857cb3982e026b89','d393605fe8fbd4339b0454efc4fd24fb2','d2c4cf043d29e464cabaa44808d801687','dd87fb183ca12421488b49582df82f63b','de30f4b3406f045b88a9dde474db057db','ddc9dea0822324dd79e87c4fbf8f552bd','d085059bd4b0841be9df4008012070be1','d4c8ac9e2a1a148fcb78997c03c4838f7','d6c4297de738b4f1086e9ea8454b284ba','d44423b7bc7ed438ab68c373cedde5c44');
```

Start with one of the `instance_id`s from the previous step:
 - Find the service tied to the instance ID:  `cf curl /v2/service_instances/<instance id>`  This will give you the name of the service
 - Find the apps tied to the service: `cf curl /v2/service_instances/<instance id>/service_bindings`  This will give app_guids (there can be more than one)
 - Find the app name and space guid the app is in: `cf curl /v2/apps/<app guid>`
 - Find the space name for the app: `cf curl /v2/spaces/<space_guid>`
 - Find the org name for the app: `cf curl /v2/organizations/organization_guid>`

### Perform the Bind/Restage

 - `cf target -o <org name> -s <space name>`
 - `cf stop <app name>`
 - `cf unbind-service <app namne> <service name>`
 - `cf bind-service <app name> <service name>`
 - `cf restage <app name>`

Check that the new IP is in the DSN: `cf curl /v2/service_instances/<instance id>/service_bindings`

For more information on all the `cf curl` options refer to http://apidocs.cloudfoundry.org/217/
