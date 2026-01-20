# Financial Services compliant example

<!-- BEGIN SCHEMATICS DEPLOY HOOK -->
<a href="https://cloud.ibm.com/schematics/workspaces/create?workspace_name=icd-redis-fscloud-example&repository=https://github.com/terraform-ibm-modules/terraform-ibm-icd-redis/tree/main/examples/fscloud"><img src="https://img.shields.io/badge/Deploy%20with IBM%20Cloud%20Schematics-0f62fe?logo=ibm&logoColor=white&labelColor=0f62fe" alt="Deploy with IBM Cloud Schematics" style="height: 16px; vertical-align: text-bottom;"></a>
<!-- END SCHEMATICS DEPLOY HOOK -->


This example uses the [Profile for IBM Cloud Framework for Financial Services](https://github.com/terraform-ibm-modules/terraform-ibm-icd-redis/tree/main/modules/fscloud) to provision an IBM Cloud Databases for Redis instance.

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

<!-- BEGIN SCHEMATICS DEPLOY TIP HOOK -->
:information_source: Ctrl/Cmd+Click or right-click on the Schematics deploy button to open in a new tab
<!-- END SCHEMATICS DEPLOY TIP HOOK -->
