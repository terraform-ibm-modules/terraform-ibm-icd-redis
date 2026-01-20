# Complete example

<!-- BEGIN SCHEMATICS DEPLOY HOOK -->
<a href="https://cloud.ibm.com/schematics/workspaces/create?workspace_name=icd-redis-complete-example&repository=https://github.com/terraform-ibm-modules/terraform-ibm-icd-redis/tree/main/examples/complete"><img src="https://img.shields.io/badge/Deploy%20with IBM%20Cloud%20Schematics-0f62fe?logo=ibm&logoColor=white&labelColor=0f62fe" alt="Deploy with IBM Cloud Schematics" style="height: 16px; vertical-align: text-bottom;"></a>
<!-- END SCHEMATICS DEPLOY HOOK -->


This example creates an IBM Cloud Database for Redis instance with KMS encryption enabled and CBR rules configured.

The following resources are provisioned by this example:

- A new resource group, if an existing one is not passed in.
- A basic VPC and subnet.
- A Key Protect instance with a root key in the given resource group and region.
- An instance of Databases for Redis with KMS encryption enabled.
- Service credentials for the database instance.
- A context-based restriction (CBR) rule to only allow Redis to be accessible from within the VPC.

<!-- BEGIN SCHEMATICS DEPLOY TIP HOOK -->
:information_source: Ctrl/Cmd+Click or right-click on the Schematics deploy button to open in a new tab
<!-- END SCHEMATICS DEPLOY TIP HOOK -->
