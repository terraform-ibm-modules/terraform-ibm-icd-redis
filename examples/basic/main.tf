locals {
  # Determine if gen2 plan is being used
  is_gen2 = can(regex("-gen2$", var.plan))

  gen2_host_flavor    = "bx3d.4x20"
  classic_host_flavor = "multitenant"

  endpoint_type = var.service_endpoints == "public-and-private" ? "private" : var.service_endpoints

  gen2_service_credential_names = [
    {
      name     = "redis_manager"
      role     = "Manager"
      endpoint = local.endpoint_type
    },
    {
      name     = "redis_writer"
      role     = "Writer"
      endpoint = local.endpoint_type
    }
  ]
  classic_service_credential_names = [
    {
      name     = "redis_admin"
      role     = "Administrator"
      endpoint = local.endpoint_type
    },
    {
      name     = "redis_operator"
      role     = "Operator"
      endpoint = local.endpoint_type
    },
    {
      name     = "redis_viewer"
      role     = "Viewer"
      endpoint = local.endpoint_type
    },
    {
      name     = "redis_editor"
      role     = "Editor"
      endpoint = local.endpoint_type
    }
  ]
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

##############################################################################
# Redis
##############################################################################

module "database" {
  source = "../.."
  # remove the above line and uncomment the below 2 lines to consume the module from the registry
  # source            = "terraform-ibm-modules/icd-redis/ibm"
  # version           = "X.Y.Z" # Replace "X.Y.Z" with a release version to lock into a specific release
  resource_group_id        = module.resource_group.resource_group_id
  name                     = "${var.prefix}-data-store"
  region                   = var.region
  plan                     = var.plan
  redis_version            = var.redis_version
  access_tags              = var.access_tags
  resource_tags            = var.resource_tags
  service_endpoints        = var.service_endpoints
  member_host_flavor       = local.is_gen2 ? local.gen2_host_flavor : local.classic_host_flavor
  disk_mb                  = local.is_gen2 ? 10240 : 1024
  deletion_protection      = false
  service_credential_names = local.is_gen2 ? local.gen2_service_credential_names : local.classic_service_credential_names
}
