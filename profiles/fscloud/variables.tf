##############################################################################
# Input Variables
##############################################################################

variable "resource_group_id" {
  description = "ID of resource group to use when creating the redis database"
  type        = string
}

variable "redis_version" {
  description = "The version of redis. If null, the current default ICD redis version is used."
  type        = string
  default     = null
}

variable "region" {
  description = "The region redis is to be created on. The region must support BYOK ( us-south, us-east, and eu-de)"
  type        = string
  default     = "us-south"
  validation {
    condition = anytrue([
      var.region == "us-south",
      var.region == "us-east",
      var.region == "eu-de"
    ])
    error_message = "region must be in a BYOK location, us-south, us-east or eu-de"
  }
}

variable "allowlist" {
  description = "(Optional, List of Objects) A list of allowed IP addresses for the database."
  type = list(object({
    address     = string
    description = string
  }))
  default = []
}

variable "configuration" {
  description = "Database Configuration in JSON format."
  type = object({
    maxmemory                   = optional(number)
    maxmemory-policy            = optional(string)
    appendonly                  = optional(string)
    maxmemory-samples           = optional(number)
    stop-writes-on-bgsave-error = optional(string)
  })
  default = null
}

variable "cpu_count" {
  description = "Number of CPU cores available to the redis instance"
  type        = number
  default     = 3
}

variable "memory_mb" {
  description = "Memory available to the redis instance"
  type        = number
  default     = 1024
}

variable "disk_mb" {
  description = "Disk space available to the redis instance"
  type        = number
  default     = 20480
}


variable "instance_name" {
  description = "Name of the new redis instance"
  type        = string
}

variable "tags" {
  type        = list(any)
  description = "Optional, Array of Strings) A list of tags that you want to add to your instance."
  default     = []
}

variable "key_protect_key_crn" {
  type        = string
  description = "The root key CRN of a Key Management Service like Key Protect or Hyper Protect Crypto Service (HPCS) that you want to use for disk encryption. See https://cloud.ibm.com/docs/cloud-databases?topic=cloud-databases-key-protect&interface=ui#key-byok for current list of supported regions for BYOK"
}

variable "backup_encryption_key_crn" {
  type        = string
  description = "The CRN of a key protect key, that you want to use for encrypting disk that holds deployment backups. If null, will use 'key_protect_key_crn' as encryption key. If 'key_protect_key_crn' is also null database is encrypted by using randomly generated keys."
  default     = null
}

variable "auto_scaling" {
  type = object({
    cpu = object({
      rate_increase_percent       = optional(number)
      rate_limit_count_per_member = optional(number)
      rate_period_seconds         = optional(number)
      rate_units                  = optional(string)
    })
    disk = object({
      capacity_enabled             = optional(bool)
      free_space_less_than_percent = optional(number)
      io_above_percent             = optional(number)
      io_enabled                   = optional(bool)
      io_over_period               = optional(string)
      rate_increase_percent        = optional(number)
      rate_limit_mb_per_member     = optional(number)
      rate_period_seconds          = optional(number)
      rate_units                   = optional(string)
    })
    memory = object({
      io_above_percent         = optional(number)
      io_enabled               = optional(bool)
      io_over_period           = optional(string)
      rate_increase_percent    = optional(number)
      rate_limit_mb_per_member = optional(number)
      rate_period_seconds      = optional(number)
      rate_units               = optional(string)
    })
  })
  description = "(Optional) Configure rules to allow your database to automatically increase its resources. Single block of autoscaling is allowed at once."
  default = {
    cpu    = {}
    disk   = {}
    memory = {}
  }
}

##############################################################
# Context-based restriction (CBR)
##############################################################

variable "cbr_rules" {
  type = list(object({
    description = string
    account_id  = string
    rule_contexts = list(object({
      attributes = optional(list(object({
        name  = string
        value = string
    }))) }))
    enforcement_mode = string
  }))
  description = "(Optional, list) List of CBR rules to create"
  default     = []
  # Validation happens in the rule module
}
