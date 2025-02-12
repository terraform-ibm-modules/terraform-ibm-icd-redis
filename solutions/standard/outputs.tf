##############################################################################
# Outputs
##############################################################################

output "id" {
  description = "Redis instance id"
  value       = local.redis_id
}

output "version" {
  description = "Redis instance version"
  value       = local.redis_version
}

output "guid" {
  description = "Redis instance guid"
  value       = local.redis_guid
}

output "crn" {
  description = "Redis instance crn"
  value       = local.redis_crn
}

output "cbr_rule_ids" {
  description = "CBR rule ids created to restrict Redis"
  value       = module.redis.cbr_rule_ids
}

output "service_credentials_json" {
  description = "Service credentials json map"
  value       = module.redis.service_credentials_json
  sensitive   = true
}

output "service_credentials_object" {
  description = "Service credentials object"
  value       = module.redis.service_credentials_object
  sensitive   = true
}

output "adminuser" {
  description = "Database admin user name"
  value       = module.redis.adminuser
}

output "hostname" {
  description = "Database connection hostname"
  value       = local.redis_hostname
}

output "port" {
  description = "Database connection port"
  value       = local.redis_port
}

output "certificate_base64" {
  description = "Database connection certificate"
  value       = module.redis.certificate_base64
  sensitive   = true
}

output "secrets_manager_secrets" {
  description = "Service credential secrets"
  value       = length(local.service_credential_secrets) > 0 ? module.secrets_manager_service_credentials[0].secrets : null
}
