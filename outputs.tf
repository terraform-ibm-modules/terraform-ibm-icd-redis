##############################################################################
# Outputs
##############################################################################

output "id" {
  description = "redis instance id"
  value       = ibm_database.redis_database.id
}

output "version" {
  description = "redis instance version"
  value       = ibm_database.redis_database.version
}
