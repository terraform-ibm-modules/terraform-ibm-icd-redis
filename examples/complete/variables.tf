variable "ibmcloud_api_key" {
  type        = string
  description = "The IBM Cloud API Key"
  sensitive   = true
}

variable "region" {
  type        = string
  description = "The region redis is to be created on. The region must support BYOK if key_protect_key_crn is used"
  default     = "us-south"
}

variable "prefix" {
  type        = string
  description = "Prefix to append to all resources created by this example"
  default     = "redis"
}

variable "resource_group" {
  type        = string
  description = "An existing resource group name to use for this example, if unset a new resource group will be created"
  default     = null
}

variable "redis_version" {
  description = "The version of redis. If null, the current default ICD redis version is used."
  type        = string
  default     = null
}

variable "sm_service_plan" {
  type        = string
  description = "Secrets Manager plan"
  default     = "trial"
}

variable "resource_tags" {
  type        = list(string)
  description = "Optional list of tags to be added to created resources"
  default     = []
}

variable "service_credentials" {
  description = "A list of service credentials that you want to create for the database"
  type        = list(string)
  default     = ["redis_credential_microservices", "redis_credential_dev_1", "redis_credential_dev_2"]
}

variable "existing_sm_instance_guid" {
  type        = string
  description = "Existing Secrets Manager GUID. If not provided an new instance will be provisioned"
  default     = null
}

variable "existing_sm_instance_region" {
  type        = string
  description = "Required if value is passed into var.existing_sm_instance_guid"
  default     = null
}
