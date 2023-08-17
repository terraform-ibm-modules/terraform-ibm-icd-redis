module "redis" {
  source                        = "../.."
  resource_group_id             = var.resource_group_id
  redis_version                 = var.redis_version
  region                        = var.region
  skip_iam_authorization_policy = var.skip_iam_authorization_policy
  instance_name                 = var.instance_name
  endpoints                     = "private"
  cbr_rules                     = var.cbr_rules
  configuration                 = var.configuration
  cpu_count                     = var.cpu_count
  memory_mb                     = var.memory_mb
  members                       = var.members
  admin_pass                    = var.admin_pass
  users                         = var.users
  disk_mb                       = var.disk_mb
  kms_encryption_enabled        = true
  existing_kms_instance_guid    = var.existing_kms_instance_guid
  kms_key_crn                   = var.kms_key_crn
  backup_encryption_key_crn     = null # Need to use default encryption until ICD adds HPCS support for backup encryption
  auto_scaling                  = var.auto_scaling
  tags                          = var.tags
  service_credential_names      = var.service_credential_names
}
