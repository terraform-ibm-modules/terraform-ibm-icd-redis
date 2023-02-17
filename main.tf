/********************************************************************
This file is used to implement the ROOT module.
*********************************************************************/

locals {
  kp_backup_crn = var.backup_encryption_key_crn != null ? var.backup_encryption_key_crn : var.key_protect_key_crn
}

resource "ibm_database" "redis_database" {
  name              = var.instance_name
  plan              = "standard"
  location          = var.region
  service           = "databases-for-redis"
  version           = var.redis_version
  resource_group_id = var.resource_group_id
  service_endpoints = var.endpoints
  tags              = var.tags

  key_protect_key           = var.key_protect_key_crn
  backup_encryption_key_crn = local.kp_backup_crn

  configuration = var.configuration != null ? jsonencode(var.configuration) : null

  auto_scaling {
    cpu {
      rate_increase_percent       = var.auto_scaling.cpu.rate_increase_percent
      rate_limit_count_per_member = var.auto_scaling.cpu.rate_limit_count_per_member
      rate_period_seconds         = var.auto_scaling.cpu.rate_period_seconds
      rate_units                  = var.auto_scaling.cpu.rate_units
    }
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

  dynamic "allowlist" {
    for_each = { for ipaddress in var.allowlist : ipaddress.address => ipaddress }
    content {
      address     = allowlist.value.address
      description = allowlist.value.description
    }
  }

  lifecycle {
    ignore_changes = [
      # Ignore changes to version because a change here will destroy and recreate the instance
      version,
      key_protect_key,
      backup_encryption_key_crn,
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
  source           = "git::https://github.com/terraform-ibm-modules/terraform-ibm-cbr//cbr-rule-module?ref=v1.1.2"
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
        value    = ibm_database.redis_database.guid
        operator = "stringEquals"
      },
      {
        name     = "serviceName"
        value    = "databases-for-redis"
        operator = "stringEquals"
      }
    ],
    tags = var.cbr_rules[count.index].tags
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
