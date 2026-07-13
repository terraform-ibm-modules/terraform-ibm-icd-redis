##############################################################################
# Outputs
##############################################################################

output "id" {
  description = "Redis instance id"
  value       = ibm_database.redis_database.id
}

output "version" {
  description = "Redis instance version"
  value       = ibm_database.redis_database.version
}

output "guid" {
  description = "Redis instance guid"
  value       = ibm_database.redis_database.guid
}

output "crn" {
  description = "Redis instance crn"
  value       = ibm_database.redis_database.resource_crn
}

output "service_credentials_json" {
  description = "Service credentials json map"
  value       = local.service_credentials_json
  sensitive   = true
}

output "service_credentials_object" {
  description = "Service credentials object"
  value       = local.service_credentials_object
  sensitive   = true
}

output "cbr_rule_ids" {
  description = "CBR rule ids created to restrict Redis"
  value       = module.cbr_rule[*].rule_id
}

output "adminuser" {
  description = "Database admin user name"
  value       = ibm_database.redis_database.adminuser
}

output "hostname" {
  description = "Database connection hostname"
  value       = can(data.ibm_database_connection.database_connection[0].rediss[0].hosts[0].hostname) ? data.ibm_database_connection.database_connection[0].rediss[0].hosts[0].hostname : null
}

output "port" {
  description = "Database connection port"
  value       = can(data.ibm_database_connection.database_connection[0].rediss[0].hosts[0].port) ? data.ibm_database_connection.database_connection[0].rediss[0].hosts[0].port : null
}

output "certificate_base64" {
  description = "Database connection certificate"
  value       = can(data.ibm_database_connection.database_connection[0].rediss[0].certificate[0].certificate_base64) ? data.ibm_database_connection.database_connection[0].rediss[0].certificate[0].certificate_base64 : null
  sensitive   = true
}
