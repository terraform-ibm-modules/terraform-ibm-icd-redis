
locals {
  existing_kms_instance_crn_split = var.existing_kms_instance_crn != null ? split(":", var.existing_kms_instance_crn) : null
  existing_kms_instance_guid      = var.existing_kms_instance_crn != null ? element(local.existing_kms_instance_crn_split, length(local.existing_kms_instance_crn_split) - 3) : null
  existing_kms_instance_region    = var.existing_kms_instance_crn != null ? element(local.existing_kms_instance_crn_split, length(local.existing_kms_instance_crn_split) - 5) : null

  redis_key_name      = var.prefix != null ? "${var.prefix}-${var.redis_key_name}" : var.redis_key_name
  redis_key_ring_name = var.prefix != null ? "${var.prefix}-${var.redis_key_ring_name}" : var.redis_key_ring_name

  kms_key_crn = var.existing_kms_key_crn != null ? var.existing_kms_key_crn : module.kms[0].keys[format("%s.%s", local.redis_key_ring_name, local.redis_key_name)].crn
}

#######################################################################################################################
# Resource Group
#######################################################################################################################

module "resource_group" {
  source                       = "terraform-ibm-modules/resource-group/ibm"
  version                      = "1.1.6"
  resource_group_name          = var.use_existing_resource_group == false ? (var.prefix != null ? "${var.prefix}-${var.resource_group_name}" : var.resource_group_name) : null
  existing_resource_group_name = var.use_existing_resource_group == true ? var.resource_group_name : null
}

#######################################################################################################################
# KMS root key for Elasticsearch
#######################################################################################################################

module "kms" {
  providers = {
    ibm = ibm.kms
  }
  count                       = var.existing_kms_key_crn != null ? 0 : 1 # no need to create any KMS resources if passing an existing key
  source                      = "terraform-ibm-modules/kms-all-inclusive/ibm"
  version                     = "4.13.4"
  create_key_protect_instance = false
  region                      = local.existing_kms_instance_region
  existing_kms_instance_guid  = local.existing_kms_instance_guid
  key_ring_endpoint_type      = var.kms_endpoint_type
  key_endpoint_type           = var.kms_endpoint_type
  keys = [
    {
      key_ring_name         = local.redis_key_ring_name
      existing_key_ring     = false
      force_delete_key_ring = true
      keys = [
        {
          key_name                 = local.redis_key_name
          standard_key             = false
          rotation_interval_month  = 3
          dual_auth_delete_enabled = false
          force_delete             = true
        }
      ]
    }
  ]
}

module "redis"{
  source = "../../modules/fscloud"
  resource_group_id = module.resource_group.resource_group_id
  instance_name = var.prefix == null ? var.name: "${var.prefix}-${var.name}"
  region = var.region
  skip_iam_authorization_policy = var.skip_iam_authorization_policy
  redis_version = var.redis_version

  existing_kms_instance_guid = local.existing_kms_instance_guid
  kms_key_crn = local.kms_key_crn
  tags = var.tags
  admin_pass = var.admin_pass
  users = var.users
  members = var.members
  # member_host_flavor            = var.member_host_flavor
  memory_mb =  var.member_memory_mb
  disk_mb = var.member_disk_mb
  cpu_count = var.member_cpu_count
  auto_scaling = var.auto_scaling
  configuration = var.configuration
  service_credential_names = var.service_credential_names
  }


