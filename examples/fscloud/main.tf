##############################################################################
# Resource Group
##############################################################################

module "resource_group" {
  source  = "terraform-ibm-modules/resource-group/ibm"
  version = "1.1.6"
  # if an existing resource group is not set (null) create a new one using prefix
  resource_group_name          = var.resource_group == null ? "${var.prefix}-resource-group" : null
  existing_resource_group_name = var.resource_group
}

##############################################################################
# VPC
##############################################################################
resource "ibm_is_vpc" "example_vpc" {
  name           = "${var.prefix}-vpc"
  resource_group = module.resource_group.resource_group_id
  tags           = var.resource_tags
}

resource "ibm_is_subnet" "testacc_subnet" {
  name                     = "${var.prefix}-subnet"
  vpc                      = ibm_is_vpc.example_vpc.id
  zone                     = "${var.region}-1"
  total_ipv4_address_count = 256
  resource_group           = module.resource_group.resource_group_id
}

##############################################################################
# Get Cloud Account ID
##############################################################################

data "ibm_iam_account_settings" "iam_account_settings" {
}

##############################################################################
# Create CBR Zone
##############################################################################
module "cbr_zone" {
  source           = "terraform-ibm-modules/cbr/ibm//modules/cbr-zone-module"
  version          = "1.22.2"
  name             = "${var.prefix}-VPC-network-zone"
  zone_description = "CBR Network zone containing VPC"
  account_id       = data.ibm_iam_account_settings.iam_account_settings.account_id
  addresses = [{
    type  = "vpc", # to bind a specific vpc to the zone
    value = ibm_is_vpc.example_vpc.crn,
  }]
}

##############################################################################
# Redis
##############################################################################

module "redis" {
  source                     = "../../modules/fscloud"
  resource_group_id          = module.resource_group.resource_group_id
  instance_name              = "${var.prefix}-redis"
  region                     = var.region
  redis_version              = var.redis_version
  tags                       = var.resource_tags
  kms_key_crn                = var.kms_key_crn
  existing_kms_instance_guid = var.existing_kms_instance_guid
  service_credential_names   = var.service_credential_names
  auto_scaling               = var.auto_scaling
  cbr_rules = [
    {
      description      = "sample rule"
      enforcement_mode = "enabled"
      account_id       = data.ibm_iam_account_settings.iam_account_settings.account_id
      tags = [
        {
          name  = "environment"
          value = "${var.prefix}-test"
        },
        {
          name  = "terraform-rule"
          value = "allow-${var.prefix}-vpc-to-${var.prefix}-redis"
        }
      ]
      rule_contexts = [{
        attributes = [
          {
            "name" : "endpointType",
            "value" : "private"
          },
          {
            name  = "networkZoneId"
            value = module.cbr_zone.zone_id
        }]
      }]
    }
  ]
}
