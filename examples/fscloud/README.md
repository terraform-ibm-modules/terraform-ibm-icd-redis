# Financial Services compliant example

This example uses the [Profile for IBM Cloud Framework for Financial Services](../../modules/fscloud/) to provision an IBM Cloud Databases for Redis instance.

The following resources are provisioned by this example:

- A new resource group, if an existing one is not passed in.
- An IAM authorization between all Redis database instances in the given resource group, and the IBM Hyper Protect Crypto Services (HPCS) instance that is passed in.
- An IBM Cloud Database for Redis database instance that is encrypted with the Hyper Protect Crypto Services (HPCS) root key that is passed in.
- Autoscaling rules for the IBM Cloud Database for Redis database instance.
- Service Credentials for the Database for Redis instance.
- A basic virtual private cloud (VPC).
- A context-based restriction (CBR) rule to only allow Redis to be accessible from within the VPC.

:exclamation: **Important:** In this example, only the IBM Cloud Database for Redis instance complies with the IBM Cloud Framework for Financial Services. Other parts of the infrastructure do not necessarily comply.

## Before you begin

Before you run the example, make sure that you set up the following prerequisites:

- A Hyper Protect Crypto Services (HPCS) instance and root key available in the region that you want to deploy your Database for Redis instance to.
