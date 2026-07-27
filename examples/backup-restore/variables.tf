variable "ibmcloud_api_key" {
  type        = string
  description = "The IBM Cloud API Key"
  sensitive   = true
}

variable "region" {
  type        = string
  description = "Region to provision all resources created by this example."
  default     = "us-south"
}

variable "prefix" {
  type        = string
  description = "Prefix to append to all resources created by this example"
  default     = "backup"
}

variable "redis_version" {
  type        = string
  description = "Version of the redis instance. If no value passed, the current ICD preferred version is used."
  default     = null
}

variable "resource_group" {
  type        = string
  description = "An existing resource group name to use for this example, if unset a new resource group will be created"
  default     = null
}

variable "resource_tags" {
  type        = list(string)
  description = "Optional list of tags to be added to created resources"
  default     = []
}

variable "access_tags" {
  type        = list(string)
  description = "A list of access tags to apply to the redis instance created by the module, see https://cloud.ibm.com/docs/account?topic=account-access-tags-tutorial for more details"
  default     = []
}

variable "plan" {
  type        = string
  description = "The name of the service plan that you choose for your Redis instance"
  default     = "standard"

  validation {
    condition = anytrue([
      var.plan == "standard",
      var.plan == "standard-gen2",
    ])
    error_message = "Only supported plans are standard and standard-gen2"
  }
}

variable "existing_database_crn" {
  type        = string
  description = "The CRN of the existing database deployment whose latest backup will be used to restore from. Used for classic plan instances."
  default     = null
}

variable "backup_crn" {
  type        = string
  description = "The CRN of a specific backup to restore from (crn:v1:<...>:backup:). Required for gen2 plan instances, since the ibm_database_backups data source does not support gen2 deployments."
  default     = null
}
