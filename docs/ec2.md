# EC2 Notes

First add a `rdpg` security group which allows the following TCP ports:
|port|service|
|6432|PostgreSQL Write|
|7432|PostgreSQL Read|
|9999|HAProxy Statistics & Admin|
|8500|Consul UI|

After generating the manifest, go in and replace the value of `SUBNET_ID_GOES_HERE` with the subnet within your VPC you will be using.

