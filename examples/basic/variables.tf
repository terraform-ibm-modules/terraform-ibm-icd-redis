variable "ibmcloud_api_key" {
  type        = string
  description = "The IBM Cloud API Key"
  sensitive   = true
}

variable "region" {
  type        = string
  description = "Region to provision all resources created by this example"
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

variable "access_tags" {
  type        = list(string)
  description = "A list of access tags to apply to the Redis instance created by the module, see https://cloud.ibm.com/docs/account?topic=account-access-tags-tutorial for more details"
  default     = []
}

variable "resource_tags" {
  type        = list(string)
  description = "Optional list of tags to be added to created resources"
  default     = []
}

variable "redis_version" {
  description = "Version of the Redis instance to provision. If no value is passed, the current preferred version of IBM Cloud Databases is used."
  type        = string
  default     = null
}

variable "service_endpoints" {
  type        = string
  description = "The type of endpoint of the database instance. Possible values: `public`, `private`, `public-and-private`."
  default     = "public"

  validation {
    condition     = can(regex("public|public-and-private|private", var.service_endpoints))
    error_message = "Valid values for service_endpoints are 'public', 'public-and-private', and 'private'"
  }
}
variable "member_host_flavor" {
  type        = string
  description = "The host flavor per member. [Learn more](https://registry.terraform.io/providers/IBM-Cloud/ibm/latest/docs/resources/database#host_flavor)."
  default     = "multitenant"
  # Validation is done in the Terraform plan phase by the IBM provider, so no need to add extra validation here.
}
