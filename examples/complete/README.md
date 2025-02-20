# Complete example

This example creates an IBM Cloud Database for Redis instance with KMS encryption enabled and CBR rules configured.

The following resources are provisioned by this example:

- A new resource group, if an existing one is not passed in.
- A basic VPC and subnet.
- A Key Protect instance with a root key in the given resource group and region.
- An instance of Databases for Redis with KMS encryption enabled.
- Service credentials for the database instance.
- A context-based restriction (CBR) rule to only allow Redis to be accessible from within the VPC.
