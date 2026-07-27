locals {
  is_gen2 = can(regex("-gen2$", var.plan))
  # For classic plans, derive the backup CRN from the ibm_database_backups data source.
  # For gen2 plans, the data source does not work, so var.backup_crn must be supplied directly.
  resolved_backup_crn = local.is_gen2 ? var.backup_crn : data.ibm_database_backups.backup_database[0].backups[0].backup_id
}

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

# Only used for classic plan — ibm_database_backups does not support gen2 deployments
data "ibm_database_backups" "backup_database" {
  count         = local.is_gen2 ? 0 : 1
  deployment_id = var.existing_database_crn
}

# New redis instance restored from a backup
module "restored_icd_redis" {
  source = "../../"
  # remove the above line and uncomment the below 2 lines to consume the module from the registry
  # source            = "terraform-ibm-modules/icd-redis/ibm"
  # version           = "X.Y.Z" # Replace "X.Y.Z" with a release version to lock into a specific release
  resource_group_id   = module.resource_group.resource_group_id
  name                = "${var.prefix}-redis-restored"
  redis_version       = var.redis_version
  region              = var.region
  plan                = var.plan
  resource_tags       = var.resource_tags
  access_tags         = var.access_tags
  member_host_flavor  = local.is_gen2 ? "bx3d.4x20" : "multitenant"
  disk_mb             = local.is_gen2 ? 15360 : 1024
  deletion_protection = false
  backup_crn          = local.resolved_backup_crn
}
