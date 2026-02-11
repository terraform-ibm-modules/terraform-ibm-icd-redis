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

output "hostname" {
  description = "Database connection hostname"
  value       = module.redis.hostname
}

output "port" {
  description = "Database connection port"
  value       = module.redis.port
}

output "secrets_manager_secrets" {
  description = "Service credential secrets"
  value       = module.redis.secrets_manager_secrets
}

output "next_steps_text" {
  value       = module.redis.next_steps_text
  description = "Next steps text"
}

output "next_step_primary_label" {
  value       = module.redis.next_step_primary_label
  description = "Primary label"
}

output "next_step_primary_url" {
  value       = module.redis.next_step_primary_url
  description = "Primary URL"
}

output "next_step_secondary_label" {
  value       = module.redis.next_step_secondary_label
  description = "Secondary label"
}

output "next_step_secondary_url" {
  value       = module.redis.next_step_secondary_url
  description = "Secondary URL"
}
