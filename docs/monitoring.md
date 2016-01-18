# Monitoring

## Client Connection Statistics

To get a listing of all connection statistics (CSV format) from haproxy:
```
echo show stat | /var/vcap/packages/socat/bin/socat /var/vcap/sys/run/haproxy/haproxy.sock stdio
```

[More haproxy Command Information](http://cbonte.github.io/haproxy-dconv/configuration-1.5.html#9.2)

## HAProxy UI

[HAProxy Stats UI](http://10.244.2.2:9999/haproxy/stats) 
The username and password for development is admin/admin. 
For production you can find the configured username and password in your 
deployment manifest.

## PGBouncer

## Service Health

[Consul UI](http://10.244.2.2:8500/ui/#/dc1/nodes/rdpg-rdpg-0)

## Administration 

Connect to the first node postgres, the password for the pgbdr database in 
development is 'pgbdr', for production you can find it configured in your 
deployment manifest.

```
psql -U postgres -H 10.244.2.2 --port 6432 pgbdr
```
