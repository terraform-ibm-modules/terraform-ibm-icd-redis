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

output "service_credentials_json" {
  description = "Service credentials json map"
  value       = var.existing_redis_instance_crn != null ? null : module.redis[0].service_credentials_json
  sensitive   = true
}

output "service_credentials_object" {
  description = "Service credentials object"
  value       = var.existing_redis_instance_crn != null ? null : module.redis[0].service_credentials_object
  sensitive   = true
}

output "hostname" {
  description = "Database connection hostname"
  value       = local.redis_hostname
}

output "port" {
  description = "Database connection port"
  value       = local.redis_port
}

output "secrets_manager_secrets" {
  description = "Service credential secrets"
  value       = length(local.service_credential_secrets) > 0 ? module.secrets_manager_service_credentials[0].secrets : null
}

output "next_steps_text" {
  value       = "Your Database for Redis instance is ready. You can now take advantage of reduced application response time , achieve cost-optimized performance, low latency, high throughput, in a highly available and scalable solution."
  description = "Next steps text"
}

output "next_step_primary_label" {
  value       = "Deployment Details"
  description = "Primary label"
}

output "next_step_primary_url" {
  value       = "https://cloud.ibm.com/services/databases-for-redis/${local.redis_crn}"
  description = "Primary URL"
}

output "next_step_secondary_label" {
  value       = "Learn more about Databases for Redis"
  description = "Secondary label"
}

output "next_step_secondary_url" {
  value       = "https://cloud.ibm.com/docs/databases-for-redis"
  description = "Secondary URL"
}
