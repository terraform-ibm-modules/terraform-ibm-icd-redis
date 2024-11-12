##############################################################################
# Outputs
##############################################################################

output "id" {
  description = "Redis instance id"
  value       = module.redis.id
}

output "version" {
  description = "Redis instance version"
  value       = module.redis.version
}

output "guid" {
  description = "Redis instance guid"
  value       = module.redis.guid
}

output "crn" {
  description = "Redis instance crn"
  value       = module.redis.crn
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
  value       = module.redis.hostname
}

output "port" {
  description = "Database connection port"
  value       = module.redis.port
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
