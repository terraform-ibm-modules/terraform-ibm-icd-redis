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

output "hostname" {
  description = "Database hostname. Only contains value when var.service_credential_names or var.users are set."
  value       = length(var.service_credential_names) > 0 ? nonsensitive(ibm_resource_key.service_credentials[keys(var.service_credential_names)[0]].credentials["connection.rediss.hosts.0.hostname"]) : length(var.users) > 0 ? nonsensitive(flatten(data.ibm_database_connection.database_connection[0].rediss[0].hosts[0].hostname)) : null
}

output "port" {
  description = "Database port. Only contains value when var.service_credential_names or var.users are set."
  value       = length(var.service_credential_names) > 0 ? nonsensitive(ibm_resource_key.service_credentials[keys(var.service_credential_names)[0]].credentials["connection.rediss.hosts.0.port"]) : length(var.users) > 0 ? nonsensitive(flatten(data.ibm_database_connection.database_connection[0].rediss[0].hosts[0].port)) : null
}
