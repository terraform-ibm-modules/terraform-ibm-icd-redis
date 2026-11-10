##############################################################################
# Resource Group
##############################################################################

module "resource_group" {
  source  = "terraform-ibm-modules/resource-group/ibm"
  version = "1.6.1"
  # if an existing resource group is not set (null) create a new one using prefix
  resource_group_name          = var.resource_group == null ? "${var.prefix}-resource-group" : null
  existing_resource_group_name = var.resource_group
}

locals {
  is_gen2             = can(regex("-gen2$", var.plan))
  default_host_flavor = local.is_gen2 ? "bx3d.4x20" : "multitenant"
  default_disk_mb     = local.is_gen2 ? 15360 : 1024
}

data "ibm_database_backups" "backup_database" {
  deployment_id = var.existing_database_crn
}

# New redis instance pointing to the backup instance
module "restored_icd_redis" {
  source = "../../"
  # remove the above line and uncomment the below 2 lines to consume the module from the registry
  # source            = "terraform-ibm-modules/icd-redis/ibm"
  # version           = "X.Y.Z" # Replace "X.Y.Z" with a release version to lock into a specific release
  resource_group_id   = module.resource_group.resource_group_id
  name                = "${var.prefix}-redis-restored"
  plan                = var.plan
  redis_version       = var.redis_version
  region              = var.region
  resource_tags       = var.resource_tags
  access_tags         = var.access_tags
  member_host_flavor  = var.member_host_flavor != null ? var.member_host_flavor : local.default_host_flavor
  disk_mb             = var.disk_mb != null ? var.disk_mb : local.default_disk_mb
  deletion_protection = false
  backup_crn          = data.ibm_database_backups.backup_database.backups[0].backup_id
}
