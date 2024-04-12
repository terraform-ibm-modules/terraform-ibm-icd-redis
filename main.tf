##############################################################################
# ICD Redis module
##############################################################################

locals {
  # Validation (approach based on https://github.com/hashicorp/terraform/issues/25609#issuecomment-1057614400)
  # tflint-ignore: terraform_unused_declarations
  validate_kms_values = !var.kms_encryption_enabled && (var.kms_key_crn != null || var.backup_encryption_key_crn != null) ? tobool("When passing values for var.backup_encryption_key_crn or var.kms_key_crn, you must set var.kms_encryption_enabled to true. Otherwise unset them to use default encryption") : true
  # tflint-ignore: terraform_unused_declarations
  validate_kms_vars = var.kms_encryption_enabled && var.kms_key_crn == null && var.backup_encryption_key_crn == null ? tobool("When setting var.kms_encryption_enabled to true, a value must be passed for var.kms_key_crn and/or var.backup_encryption_key_crn") : true
  # tflint-ignore: terraform_unused_declarations
  validate_auth_policy = var.kms_encryption_enabled && var.skip_iam_authorization_policy == false && var.existing_kms_instance_guid == null ? tobool("When var.skip_iam_authorization_policy is set to false, and var.kms_encryption_enabled to true, a value must be passed for var.existing_kms_instance_guid in order to create the auth policy.") : true
  # tflint-ignore: terraform_unused_declarations
  validate_backup_key = var.backup_encryption_key_crn != null && var.use_default_backup_encryption_key == true ? tobool("When passing a value for 'backup_encryption_key_crn' you cannot set 'use_default_backup_encryption_key' to 'true'") : true

  # If no value passed for 'backup_encryption_key_crn' use the value of 'kms_key_crn'. If this is a HPCS key (which is not currently supported for backup encryption), default to 'null' meaning encryption is done using randomly generated keys
  # More info https://cloud.ibm.com/docs/cloud-databases?topic=cloud-databases-hpcs
  backup_encryption_key_crn = var.use_default_backup_encryption_key == true ? null : (var.backup_encryption_key_crn != null ? var.backup_encryption_key_crn : (can(regex(".*kms.*", var.kms_key_crn)) ? var.kms_key_crn : null))

  # Determine if auto scaling is enabled
  auto_scaling_enabled = var.auto_scaling == null ? [] : [1]

  # Determine what KMS service is being used for database encryption
  kms_service = var.kms_key_crn != null ? (
    can(regex(".*kms.*", var.kms_key_crn)) ? "kms" : (
      can(regex(".*hs-crypto.*", var.kms_key_crn)) ? "hs-crypto" : null
    )
  ) : null

  # maxmemory configuration should 80% of the deployment's memory.
  calculate_config_maxmemory = tonumber(format("%.0f", var.memory_mb * 0.8))
}

# Create IAM Authorization Policies to allow Redis to access KMS for the encryption key
resource "ibm_iam_authorization_policy" "kms_policy" {
  count                       = var.kms_encryption_enabled == false || var.skip_iam_authorization_policy ? 0 : 1
  source_service_name         = "databases-for-redis"
  source_resource_group_id    = var.resource_group_id
  target_service_name         = local.kms_service
  target_resource_instance_id = var.existing_kms_instance_guid
  roles                       = ["Reader"]
  description                 = "Allow all redis instances in the resource group ${var.resource_group_id} to read from the ${local.kms_service} instance GUID ${var.existing_kms_instance_guid}"
}

# workaround for https://github.com/IBM-Cloud/terraform-provider-ibm/issues/4478
resource "time_sleep" "wait_for_authorization_policy" {
  depends_on = [ibm_iam_authorization_policy.kms_policy]

  create_duration = "30s"
}

resource "ibm_database" "redis_database" {
  depends_on                = [time_sleep.wait_for_authorization_policy]
  name                      = var.instance_name
  plan                      = "standard" # Only standard plan is available for redis
  location                  = var.region
  service                   = "databases-for-redis"
  version                   = var.redis_version
  resource_group_id         = var.resource_group_id
  service_endpoints         = var.endpoints
  tags                      = var.tags
  adminpassword             = var.admin_pass
  key_protect_key           = var.kms_key_crn
  backup_encryption_key_crn = local.backup_encryption_key_crn
  backup_id                 = var.backup_crn

  # For default configuration, see here: https://cloud.ibm.com/docs/databases-for-redis?topic=databases-for-redis-changing-configuration&interface=cli
  configuration = var.configuration == null ? null : jsonencode({
    maxmemory                   = var.configuration.maxmemory != null ? var.configuration.maxmemory : local.calculate_config_maxmemory
    maxmemory-policy            = var.configuration.maxmemory-policy != null ? var.configuration.maxmemory-policy : "noeviction"
    appendonly                  = var.configuration.appendonly != null ? var.configuration.appendonly : "yes"
    maxmemory-samples           = var.configuration.maxmemory-samples != null ? var.configuration.maxmemory-samples : 5
    stop-writes-on-bgsave-error = var.configuration.stop-writes-on-bgsave-error != null ? var.configuration.stop-writes-on-bgsave-error : "yes"
  })

  dynamic "users" {
    for_each = nonsensitive(var.users != null ? var.users : [])
    content {
      name     = users.value.name
      password = users.value.password
      type     = users.value.type
      role     = (users.value.role != "" ? users.value.role : null)
    }
  }

  group {
    group_id = "member"
    memory {
      allocation_mb = var.memory_mb
    }
    disk {
      allocation_mb = var.disk_mb
    }
    cpu {
      allocation_count = var.cpu_count
    }
    members {
      allocation_count = var.members
    }
  }

  ## This for_each block is NOT a loop to attach to multiple auto_scaling blocks.
  ## This block is only used to conditionally add auto_scaling block depending on var.auto_scaling
  dynamic "auto_scaling" {
    for_each = local.auto_scaling_enabled
    content {
      disk {
        capacity_enabled             = var.auto_scaling.disk.capacity_enabled
        free_space_less_than_percent = var.auto_scaling.disk.free_space_less_than_percent
        io_above_percent             = var.auto_scaling.disk.io_above_percent
        io_enabled                   = var.auto_scaling.disk.io_enabled
        io_over_period               = var.auto_scaling.disk.io_over_period
        rate_increase_percent        = var.auto_scaling.disk.rate_increase_percent
        rate_limit_mb_per_member     = var.auto_scaling.disk.rate_limit_mb_per_member
        rate_period_seconds          = var.auto_scaling.disk.rate_period_seconds
        rate_units                   = var.auto_scaling.disk.rate_units
      }
      memory {
        io_above_percent         = var.auto_scaling.memory.io_above_percent
        io_enabled               = var.auto_scaling.memory.io_enabled
        io_over_period           = var.auto_scaling.memory.io_over_period
        rate_increase_percent    = var.auto_scaling.memory.rate_increase_percent
        rate_limit_mb_per_member = var.auto_scaling.memory.rate_limit_mb_per_member
        rate_period_seconds      = var.auto_scaling.memory.rate_period_seconds
        rate_units               = var.auto_scaling.memory.rate_units
      }
    }
  }

  lifecycle {
    ignore_changes = [
      # Ignore changes to these because a change will destroy and recreate the instance
      version,
      key_protect_key,
      backup_encryption_key_crn
    ]
  }

  timeouts {
    create = "120m" # Extending provisioning time to 120 minutes
  }
}

##############################################################################
# Context Based Restrictions
##############################################################################

module "cbr_rule" {
  count            = length(var.cbr_rules) > 0 ? length(var.cbr_rules) : 0
  source           = "terraform-ibm-modules/cbr/ibm//modules/cbr-rule-module"
  version          = "1.20.1"
  rule_description = var.cbr_rules[count.index].description
  enforcement_mode = var.cbr_rules[count.index].enforcement_mode
  rule_contexts    = var.cbr_rules[count.index].rule_contexts
  resources = [{
    attributes = [
      {
        name     = "accountId"
        value    = var.cbr_rules[count.index].account_id
        operator = "stringEquals"
      },
      {
        name     = "serviceInstance"
        value    = ibm_database.redis_database.id
        operator = "stringEquals"
      },
      {
        name     = "serviceName"
        value    = "databases-for-redis"
        operator = "stringEquals"
      }
    ]
  }]
  #  There is only 1 operation type for Redis so it is not exposed as a configuration
  operations = [{
    api_types = [
      {
        api_type_id = "crn:v1:bluemix:public:context-based-restrictions::::api-type:data-plane"
      }
    ]
  }]
}

##############################################################################
# Service Credentials
##############################################################################

resource "ibm_resource_key" "service_credentials" {
  for_each             = var.service_credential_names
  name                 = each.key
  role                 = each.value
  resource_instance_id = ibm_database.redis_database.id
}

locals {
  # used for output only
  service_credentials_json = length(var.service_credential_names) > 0 ? {
    for service_credential in ibm_resource_key.service_credentials :
    service_credential["name"] => service_credential["credentials_json"]
  } : null

  service_credentials_object = length(var.service_credential_names) > 0 ? {
    hostname    = ibm_resource_key.service_credentials[keys(var.service_credential_names)[0]].credentials["connection.rediss.hosts.0.hostname"]
    certificate = ibm_resource_key.service_credentials[keys(var.service_credential_names)[0]].credentials["connection.rediss.certificate.certificate_base64"]
    port        = ibm_resource_key.service_credentials[keys(var.service_credential_names)[0]].credentials["connection.rediss.hosts.0.port"]
    credentials = {
      for service_credential in ibm_resource_key.service_credentials :
      service_credential["name"] => {
        username = service_credential.credentials["connection.rediss.authentication.username"]
        password = service_credential.credentials["connection.rediss.authentication.password"]
      }
    }
  } : null
}

data "ibm_database_connection" "database_connection" {
  endpoint_type = var.endpoints
  deployment_id = ibm_database.redis_database.id
  user_id       = ibm_database.redis_database.adminuser
  user_type     = "database"
}
