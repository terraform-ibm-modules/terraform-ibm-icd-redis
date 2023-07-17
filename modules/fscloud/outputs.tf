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

output "hostname" {
  description = "Postgresql instance hostname"
  value       = module.redis.hostname
}

output "port" {
  description = "Postgresql instance port"
  value       = module.redis.port
}
