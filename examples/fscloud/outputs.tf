##############################################################################
# Outputs
##############################################################################

output "id" {
  description = "Redis instance id"
  value       = module.redis.id
}

output "guid" {
  description = "Redis instance guid"
  value       = module.redis.guid
}
