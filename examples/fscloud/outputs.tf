##############################################################################
# Outputs
##############################################################################

output "id" {
  description = "redis instance id"
  value       = module.redis.id
}

output "version" {
  description = "redis instance version"
  value       = module.redis.version
}
