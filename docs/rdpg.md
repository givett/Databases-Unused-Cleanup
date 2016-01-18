# rdpg Project Specific Notes

## In the bosh-lite repository directory

For the first time to run bosh-lite or after reboot the laptop, do 
```sh
./bin/add-route in bosh-lite
vagrant destroy --force # destroy bosh-lite vms
vagrant up # run bosh-lite vms
bosh login admin admin # log in as admin
```

Remove old VM ssh keys when destroyed and delete the lines with `10.244...` at 
the bottom of your `known_hosts` file: `vim ~/.ssh/known_hosts`

## In the BOSH Release directory

After vagrant destroy and up, do 

```sh
./rdpg-dev prepare warden ; bosh -n deploy
```

If it is on a new pull of boshrelease repo which you only have a rdpgd code change, do 

```sh
./rdpg-dev release ; bosh -n deploy
```

To delete a deployment

```sh
bosh -n delete deployment rdpg --force
```

SSH To a BOSH Deployed VM: bosh ssh {vm name} {vm index}, eg:

```sh
bosh ssh rdpg_manager/0 
bosh ssh rdpg_manager/1 
bosh ssh rdpg_cluster_1/0 
bosh ssh rdpg_cluster_1/1 
bosh ssh rdpg_cluster_1/2 
```

Tail on manager or service:

```sh
less /var/vcap/sys/log/rdpgd-manager/rdpgd-manager.log
less /var/vcap/sys/log/rdpgd-service/rdpgd-service.log
```

Postgres specific logs (we don’t need these usually)

```sh
less /var/vcap/sys/log/pgbdr/pgbdr.log
```

## On a VM


Tail on manager or service:

```sh
less /var/vcap/sys/log/rdpgd-manager/rdpgd-manager.log
less /var/vcap/sys/log/rdpgd-service/rdpgd-service.log
```

Postgres specific logs (we don’t need these usually)

```sh
less /var/vcap/sys/log/pgbdr/pgbdr.log
head -n 100 /var/vcap/sys/log/pgbdr/pgbdr.log # first 100 lines
tail -n 100 /var/vcap/sys/log/pgbdr/pgbdr.log # last 100 lines
tail -f /var/vcap/sys/log/pgbdr/pgbdr.log # follow
```
Controlling BOSH service jobs,

```sh
/var/vcap/bosh/bin/monit restart consul # restart consul : {start|stop|restart}
/var/vcap/bosh/bin/monit summary
```

## In PostgreSQL

Connect to PG DB:

```sh
/var/vcap/packages/pgbdr/bin/psql -U postgres --port 7432 rdpg
```

Display Extensions loaded in the current database,

```psql
\dx
```

See all connected nodes in master-master replication for the current database,
```psql
select * from bdr.bdr_nodes; 
```

## Consul
consul members # show all members in the consul cluster

Consul UI URL 

```
http://10.244.2.2:8500/ui/#/rdpg/kv/rdpg/rdpgmc/bdr/join/ip/edit
```

## Tuning based on machine sizing and usage

For a reference starting point read the 
[pgtune blog post](http://leopard.in.ua/2014/03/24/pgtune-for-postgresql/) 
and use the [website tool](http://pgtune.leopard.in.ua)

