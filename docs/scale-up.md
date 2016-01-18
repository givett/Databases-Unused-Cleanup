# Scaling Up RDPG

This document describes the manifest changes needed to scale up compute and disk for the RDPG Management Cluster (MC) nodes and the RDPG Service Cluster (SC) nodes running on AWS under BOSH.

The RDPG release will support scaling up compute and disk pool size/type by deploying over an existing deployment and preserve existing databases.

## Compute Scale Up

In the AWS manifest file, look for the following code block in the manifest for both MC and SC nodes to make the compute changes:

```
resource_pools:
- name: rdpg
  cloud_properties:
    instance_type: m3.large
```

_**EXAMPLE**_

So to move up to the top m3 instance m3.2xlarge:

```
resource_pools:
- name: rdpg
  cloud_properties:
    instance_type: m3.2xlarge
```

[Reference - AWS Instance Types](http://aws.amazon.com/ec2/instance-types/)

## Disk Space and Type Scale Up

AWS offers two types of EBS disk pools that can be used for storage with BOSH.  The default is standard disk _(type: property left blank and/or removed from the manifest)._ The higher performance SSD disk is type gp2 and needs to be specified in the manifest.

_Standard EBS Disk_
```
disk_pools:
- name: rdpgsc_disk
  disk_size: 524_288
  cloud_properties:
```
_or_
```
disk_pools:
- name: rdpgsc_disk
  disk_size: 524_288
  cloud_properties:
  type:
```
This configuration is setting a disk pool name to me rdpgsc_disk, using standard EBS disk _(type left blank)_, and setting that disk size to 512 GB.

_**EXAMPLES**_

So to increase the disk from 512 GB to 750 GB and leaving disk as type standard:

```
disk_pools:
- name: rdpgsc_disk
  disk_size: 768_432
  cloud_properties:
    type:
```

To change from standard to SSD EBS disk (gp2):

```
disk_pools:
- name: rdpgsc_disk
  disk_size: 768_432
  cloud_properties:
    type: gp2
```

##Using named disk_pools in jobs

To use a disk_pool name as a reference for a build job, the syntax is as follows:

```
- name: rdpgsc1
  resource_pool: rdpg
  persistent_disk_pool: rdpgsc_disk
```

[Reference - Bosh.io - AWS Properties](https://bosh.io/docs/aws-cpi.html)

##Project Notes

* For scale up purposes there should not be a need to scale up the compute nor disk for the MC nodes as database compute and storage is handled by the SC nodes.
* BOSH deploying to AWS has a limit of 1TB (10_240_000) for each disk pool.
* It is recommended that you perform **only one** scale up change at a time in your deployment. The **exception** is disk changes.  A size increase and disk type change can be done in the same deployment.
