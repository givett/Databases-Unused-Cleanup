# Scale Out

Scaling out allows adding a new service clusters to the existing deployment and adding new node to the management cluster. When new service clusters are added, the new requests for user databases will be distributed among the existing and new service clusters. On one hand, scaling out allows user databases span among more servers; on the other hand, it enables RDPG to flexibly increase user database capacity.

Now with service matrix idea, we can scale out very easily and meanwhile keep the manifest short and neat.

## Add New Service Clusters

Assume we already have two service clusters and we want to add the third one. The steps to modify `rdpg-warden-manifest.yml` and testing are as follows. We introduce the change in manifest, if you use `spiff` to generate manifest, please modify your `templates` and/or `jobs.yml` accordingly.

### Modify the Manifest

Modify the following `sc-pgbdr-m0-r0` job block:  change `- instances: 2` to `- instances: 3`, add one more IP to `static_ips`, in this example, we already use `- 10.244.2.14` and `- 10.244.2.18` for two nodes in the existing service clusters, we will add `- 10.244.2.30`. 
  ```
- instances: 2
   name: sc-pgbdr-m0-r0
   networks:
  - default:
    - dns
    - gateway
    name: rdpg
    static_ips:
    - 10.244.2.14
    - 10.244.2.18
  ```

Modify the following `sc-pgbdr-m0-r1` job block: change `- instances: 2` to `- instances: 3`, add one more IP to `static_ips`, in this example, we already use `- 10.244.2.22` and `- 10.244.2.26` for the two nodes in the existing service clusters, we will add `- 10.244.2.34`. 

```
- instances: 2
  name: sc-pgbdr-m0-r1
  networks:
  - default:
    - dns
    - gateway
    name: rdpg
    static_ips:
    - 10.244.2.22
    - 10.244.2.26
```

Modify networks fields as follows.

```
yaml
   - cloud_properties:
      name: random
    range: 10.244.2.28/30
    reserved:
     - 10.244.2.29
    static:
     - 10.244.2.30
   - cloud_properties:
      name: random
    range: 10.244.2.32/30
    reserved:
     - 10.244.2.33
    static:
     - 10.244.2.34 
```

Increase resources_pools size accordingly, in this example it is 2 (from 7 to 9) since we add one service cluster.

```
yaml
    resource_pools:
     - cloud_properties:
       name: random
     name: rdpg
     network: rdpg
     size: 9
```

### Cheking Cluster Status 

Run `bosh vms`, you should be able to see the new service clusters nodes running.

Go to `10.244.2.2:8500` in your favorite browser, you should be able to see the new service cluster master node and replica node show up. Also the `postgres`, `haproxy` and `pgbouncer` services should have increased four passing count.

Connect to management administrative database, you should be able to see pre-created databases in `cfsb.instances` table from the new added service cluster.

## Add New Node to the Management Cluster

We already have three nodes in the management cluster. This specific example shows you how to more management nodes by adding one more management node. Note that the management cluster should have odd number of nodes. The steps to modify `rdpg-warden-manifest.yml` and testing are as follows.

### Modify the Manifest

Modify the existing rdpgmc job block.

Modify the number of instance from 3 to 4. 

```
  yaml
    - instances: 4
      name: rdpgmc
  ```

Modify Network ips to what you use here, in this example we are adding `10.244.2.38`.
``` 
yaml
    networks:
  - default:
    - dns
    - gateway
    name: rdpg
    static_ips:
     - 10.244.2.2
     - 10.244.2.6
     - 10.244.2.10
     - 10.244.2.38
```

Modify the consul configure by adding the IP for the new node, here it is `10.244.2.38`.
```
yaml
    consul:
      debug: "true"
      join_node: 10.244.2.2
      join_nodes:
       - 10.244.2.2
       - 10.244.2.6
       - 10.244.2.10
       - 10.244.2.38
```

Modify networks fields as follows.

```
yaml
      - cloud_properties:
      name: random
    range: 10.244.2.36/30
    reserved:
     - 10.244.2.37
    static:
     - 10.244.2.38
```

Increase resources_pools size accordingly, in this example we increase one (from 9 to 10) since we add one management node.

```
yaml
    resource_pools:
     - cloud_properties:
       name: random
     name: rdpg
     network: rdpg
     size: 10
```

### Cheking Cluster Status

Run `bosh vms`, you should be able to see the new management node running.

Go to `10.244.2.2:8500` in your favorite browser, you should be able to see the new mangement node. Also the `consul` and `rdpgmc` services should have increaced one passing count.

Connect to management administrative database on the new management node, you should be able to see administrative tables such as `cfsb.instances` and `cfsb.bindings` are populated and have the same content with the corresponding tabled from other management nodes.


########################################

The following are for old way to add more service clusters, we will delete it after all the deployment has service matrix appilied.

## Add New Service Clusters

Assume we already have two service clusters and we want to add the third one. The steps to modify `rdpg-warden-manifest.yml` and testing are as follows.

### Modify the Manifest

Copy and paste `rdpgsc2` job block.

Modify name from `rdpgsc2` to `rdpgsc3`:

  ```
  yaml
    - instances: 2
      name: rdpgsc3
  ```

Modify Network ips to what you use here, in this example we are modifing `10.244.2.22` and `10.244.2.26` into `10.244.2.30` and `10.244.2.34`.

``` 
yaml
    networks:
  - default:
    - dns
    - gateway
    name: rdpg
    static_ips:
    - 10.244.2.30
    - 10.244.2.34
```

Modify the cluster_name for rdpgd_service in properties from `rdpgsc2` into `rdpgsc3`.

```
yaml
    rdpgd_service:
      cluster_name: rdpgsc3
```

Modify networks fields as follows.

```
yaml
   - cloud_properties:
      name: random
    range: 10.244.2.28/30
    reserved:
     - 10.244.2.29
    static:
     - 10.244.2.30
   - cloud_properties:
      name: random
    range: 10.244.2.32/30
    reserved:
     - 10.244.2.33
    static:
     - 10.244.2.34 
```

Increase resources_pools size accordingly, in this example it is 2 (from 7 to 9) since we add one service cluster.

```
yaml
    resource_pools:
     - cloud_properties:
       name: random
     name: rdpg
     network: rdpg
     size: 9
```

### Cheking Cluster Status 

Run `bosh vms`, you should be able to see the new service clusters nodes running.

Go to `10.244.2.2:8500` in your favorite browser, you should be able to see the new service cluster master node and replica node show up. Also the `postgres`, `haproxy` and `pgbouncer` services should have increased four passing count.

Connect to management administrative database, you should be able to see pre-created databases in `cfsb.instances` table from the new added service cluster.


