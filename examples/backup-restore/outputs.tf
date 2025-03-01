##############################################################################
# Outputs
##############################################################################
output "restored_icd_redis_id" {
  description = "Restored redis instance id"
  value       = module.restored_icd_redis.id
}

output "restored_icd_redis_version" {
  description = "Restored redis instance version"
  value       = module.restored_icd_redis.version
}
