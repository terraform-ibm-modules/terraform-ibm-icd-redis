 # IBM Cloud Databases for Redis

This architecture creates an instance of IBM Cloud Databases for Redis and supports provisioning of the following resources:

- A resource group, if one is not passed in.
- A KMS root key, if one is not passed in.
- An IBM Cloud Databases for Redis instance with KMS encryption.
- Autoscaling rules for the database instance, if provided.

![fscloud-elastic-search](../../reference-architecture/deployable-architecture-redis.svg)

:exclamation: **Important:** This solution is not intended to be called by other modules because it contains a provider configuration and is not compatible with the `for_each`, `count`, and `depends_on` arguments. For more information, see [Providers Within Modules](https://developer.hashicorp.com/terraform/language/modules/develop/providers).
