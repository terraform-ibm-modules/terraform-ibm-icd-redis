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
