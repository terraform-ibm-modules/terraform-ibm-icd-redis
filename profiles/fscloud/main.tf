locals {
  # tflint-ignore: terraform_unused_declarations
  validate_restrictions_set = (length(var.allowlist) == 0 && length(var.cbr_rules) == 0) ? tobool("Allow list and/or CBR Rules must be set") : true
}

module "redis" {
  source                    = "../.."
  resource_group_id         = var.resource_group_id
  redis_version             = var.redis_version
  region                    = var.region
  instance_name             = var.instance_name
  endpoints                 = "private"
  allowlist                 = var.allowlist
  cbr_rules                 = var.cbr_rules
  configuration             = var.configuration
  cpu_count                 = var.cpu_count
  memory_mb                 = var.memory_mb
  disk_mb                   = var.disk_mb
  key_protect_key_crn       = var.key_protect_key_crn
  backup_encryption_key_crn = var.backup_encryption_key_crn
  auto_scaling              = var.auto_scaling
  tags                      = var.tags

}
