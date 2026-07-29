# IBM Cloud Databases Gen 2 (VPC) for Redis

This deployable architecture provides a fully configurable solution for IBM Cloud Databases Gen 2 (VPC) for Redis. For more information about Gen 2, see [Databases for Redis Gen 2](https://cloud.ibm.com/docs/cloud-databases-gen2?topic=cloud-databases-gen2-overview-gen1-gen2).

:exclamation: **Important:** This solution is not intended to be called by other modules because it contains a provider configuration and is not compatible with the `for_each`, `count`, and `depends_on` arguments. For more information, see [Providers Within Modules](https://developer.hashicorp.com/terraform/language/modules/develop/providers).
