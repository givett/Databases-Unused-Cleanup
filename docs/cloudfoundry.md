# Cloud Foundry

This assumes that you have Cloud Foundry running already and the [cli](https://github.com/cloudfoundry/cli/releases)
installed

Make sure we are logged into an org/space in CF. 
As an example, for develoment when targeting bosh-lite,
```sh
cf login --skip-ssl-validation -u admin -p admin -a https://api.10.244.0.34.xip.io
cf auth admin admin
cf create-org $USER
cf target -o $USER
cf create-space rdpg
cf target -s rdpg
```

Register the rdpg service broker, this example is in development environment,
```sh
CF_TRACE=true cf create-service-broker rdpg cfadmin cfadmin http://10.244.2.2:8888
```

Allow access to the new service,
```sh
cf enable-service-access rdpg -o $USER
cf service-access
cf marketplace
```

Follow the [instructions](https://github.com/wayneeseguin/rdpg-cf-service-checks/blob/master/README.md)
for the `rdpg-cf-service-checks` application to test if the service is functional.

