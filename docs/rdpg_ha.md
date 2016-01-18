# RDPG HA Options

This document describes the manifest changes to allow High Availibility (HA) for your RDPG deployment in CF Service Broker bindings and host values returned in the DSNs.

## PGBDR host values

In the manifest file, look for the following code block in the manifest:

```
jobs:
- instances: 3
  name: rdpgmc
  ...
  properties:
  ...
    pgbdr:
```
You would then indent properly under the pgbdr section and add the following option:

```
    dsn_host: "<option>"
```

## DSN Host Options

_**manifestIP**_ - default behavior of RDPG - use the master service IP of the corresponding SC cluster

_**consulDNS**_ - Return a FQDN resolvable using the RDPG deployed DNS services in Consul

_**X.X.X.X**_- User provided IPv4 address for a load balancer to be placed in front of RDPG. PGBDR will perform a IPv4 validity check

_**{DNS name}**_ - User provided & controlled DNS name for a load balancer in front of RDPG. *NOTE* There is no DNS check to validate provided entry

##Project Notes

* manifestIP is the default value in the RDPG deployment if no user manifest changes are made.
* In the event the manifest value is blank or the IP address provided is not a valid IPv4, the default option of manifestIP will be used.
