##############################################################################
# Outputs
##############################################################################

output "id" {
  description = "redis instance id"
  value       = module.redis.id
}


output "guid" {
  description = "redis instance guid"
  value       = module.redis.guid
}


output "version" {
  description = "redis instance version"
  value       = module.redis.version
}
