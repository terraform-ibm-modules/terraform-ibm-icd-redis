
locals {
  validate_sm_region_cnd = var.existing_sm_instance_guid != null && var.existing_sm_instance_region == null
  validate_sm_region_msg = "existing_sm_instance_region must also be set when value given for existing_sm_instance_guid."
  # tflint-ignore: terraform_unused_declarations
  validate_sm_region_chk = regex(
    "^${local.validate_sm_region_msg}$",
    (!local.validate_sm_region_cnd
      ? local.validate_sm_region_msg
  : ""))

  sm_guid   = var.existing_sm_instance_guid == null ? ibm_resource_instance.secrets_manager[0].guid : var.existing_sm_instance_guid
  sm_region = var.existing_sm_instance_region == null ? var.region : var.existing_sm_instance_region
}

##############################################################################
# Resource Group
##############################################################################

module "resource_group" {
  source = "git::https://github.com/terraform-ibm-modules/terraform-ibm-resource-group.git?ref=v1.0.5"
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
# Key Protect All Inclusive
##############################################################################

module "key_protect_all_inclusive" {
  providers = {
    restapi = restapi.kp
  }
  source            = "git::https://github.com/terraform-ibm-modules/terraform-ibm-key-protect-all-inclusive.git?ref=v3.0.2"
  resource_group_id = module.resource_group.resource_group_id
  # Note: Database instance and Key Protect must be created in the same region when using BYOK
  # See https://cloud.ibm.com/docs/cloud-databases?topic=cloud-databases-key-protect&interface=ui#key-byok
  region                    = var.region
  key_protect_instance_name = "${var.prefix}-kp"
  resource_tags             = var.resource_tags
  key_map                   = { "icd" = ["${var.prefix}-redis"] }
}

# Create IAM Access Policy to allow Key protect to access Redis instance
resource "ibm_iam_authorization_policy" "policy" {
  source_service_name         = "databases-for-redis"
  source_resource_group_id    = module.resource_group.resource_group_id
  target_service_name         = "kms"
  target_resource_instance_id = module.key_protect_all_inclusive.key_protect_guid
  roles                       = ["Reader"]
}

########################################
## Create Secrets Manager layer
########################################

# Create Secrets Manager Instance
resource "ibm_resource_instance" "secrets_manager" {
  count             = var.existing_sm_instance_guid == null ? 1 : 0
  name              = "${var.prefix}-sm" #checkov:skip=CKV_SECRET_6: does not require high entropy string as is static value
  service           = "secrets-manager"
  service_endpoints = "public-and-private"
  plan              = var.sm_service_plan
  location          = var.region
  resource_group_id = module.resource_group.resource_group_id

  timeouts {
    create = "30m" # Extending provisioning time to 30 minutes
  }
}

# Add a Secrets Group to the secret manager instance
module "secrets_manager_secrets_group" {
  providers = {
    restapi = restapi.sm
  }
  source               = "git::https://github.ibm.com/GoldenEye/secrets-manager-secret-group-module.git?ref=1.5.2"
  region               = local.sm_region
  secrets_manager_guid = local.sm_guid
  #tfsec:ignore:general-secrets-no-plaintext-exposure
  secret_group_name        = "${var.prefix}-redis-secrets"
  secret_group_description = "service secret-group" #tfsec:ignore:general-secrets-no-plaintext-exposure
}

##############################################################################
# Service Credentials
##############################################################################

resource "ibm_resource_key" "service_credentials" {
  count                = length(var.service_credentials)
  name                 = var.service_credentials[count.index]
  resource_instance_id = module.icd_redis.id
  tags                 = var.resource_tags
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
  source           = "git::https://github.com/terraform-ibm-modules/terraform-ibm-cbr//cbr-zone-module?ref=v1.1.2"
  name             = "${var.prefix}-VPC-network-zone"
  zone_description = "CBR Network zone containing VPC"
  account_id       = data.ibm_iam_account_settings.iam_account_settings.account_id
  addresses = [{
    type  = "vpc", # to bind a specific vpc to the zone
    value = ibm_is_vpc.example_vpc.crn,
  }]
}

##############################################################################
# Redis Instance
##############################################################################

module "icd_redis" {
  source              = "../../"
  resource_group_id   = module.resource_group.resource_group_id
  redis_version       = var.redis_version
  instance_name       = "${var.prefix}-redis"
  endpoints           = "private"
  region              = var.region
  key_protect_key_crn = module.key_protect_all_inclusive.keys["icd.${var.prefix}-redis"].crn
  tags                = var.resource_tags
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

# The current implementation of the secrets manager module will not allow us to generate secrets from values that are unknown in advance
# The Secrets manager provider support is coming soon and the module can be refactored to handle this.
# Until then you need to apply without the following blocks then uncomment them and apply in a followup apply.
# This will not do in a production environment and is just here for inspiration

#locals {
#  # tflint-ignore: terraform_unused_declarations
#  users = [for index, user in var.service_credentials :
#    {
#      name : user
#      username : ibm_resource_key.service_credentials[index].credentials["connection.https.authentication.username"],
#      password : sensitive(ibm_resource_key.service_credentials[index].credentials["connection.https.authentication.password"]),
#      cert : sensitive(base64decode(ibm_resource_key.service_credentials[index].credentials["connection.https.certificate.certificate_base64"]))
#    }
#  ]
#}

#module "secrets_manager_service_credentials_user_pass" {
#  providers = {
#    restapi = restapi.sm
#  }
#  source                  = "git::https://github.ibm.com/GoldenEye/secrets-manager-secret-module?ref=2.3.6"
#  count                   = length(local.users)
#  region                  = var.region
#  secrets_manager_guid    = ibm_resource_instance.secrets_manager.secrets_manager_guid
#  secret_group_id         = module.secrets_manager_secrets_group.secret_group_id
#  secret_name             = local.users[count.index].name
#  secret_description      = "Redis Service Credential"
#  secret_username         = local.users[count.index].username
#  secret_payload_password = local.users[count.index].password
#}
#
#module "secrets_manager_service_credentials_cert" {
#  providers = {
#    restapi = restapi.sm
#  }
#  depends_on = ["ibm_resource_key.service_credentials"]
#  source                  = "git::https://github.ibm.com/GoldenEye/secrets-manager-secret-module?ref=2.3.6"
#  count                   = length(var.service_credentials)
#  region                  = var.region
#  secrets_manager_guid    = ibm_resource_instance.secrets_manager.secrets_manager_guid
#  secret_group_id         = module.secrets_manager_secrets_group.secret_group_id
#  secret_name             = local.users[count.index].name
#  secret_description      = "Redis Service Credential Certificate"
#  imported_cert              = true
#  imported_cert_certificate  =  local.users[count.index].cert
#}
