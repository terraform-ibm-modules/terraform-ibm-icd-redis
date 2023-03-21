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
  validation {
    condition = anytrue([
      var.redis_version == null,
      var.redis_version == "5",
      var.redis_version == "6"
    ])
    error_message = "Version must be 5 or 6. If null, the current default ICD redis version is used"
  }
}

variable "region" {
  description = "The region redis is to be created on. The region must support BYOK if key_protect_key_crn is used"
  type        = string
  default     = "us-south"
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
  validation {
    condition = alltrue([
      var.cpu_count >= 3,
      var.cpu_count <= 28
    ])
    error_message = "cpus must be >= 3 and <= 28 in increments of 1"
  }
}

variable "memory_mb" {
  description = "Memory available to the redis instance"
  type        = number
  default     = 1024
  validation {
    condition = alltrue([
      var.memory_mb >= 1024,
      var.memory_mb <= 114688
    ])
    error_message = "member group memory must be >= 1024 and <= 114688 in increments of 128"
  }
}

variable "disk_mb" {
  description = "Disk space available to the redis instance"
  type        = number
  default     = 20480
  validation {
    condition = alltrue([
      var.disk_mb >= 5120,
      var.disk_mb <= 4194304
    ])
    error_message = "member group disk must be >= 5120 and <= 4194304 in increments of 1024"
  }
}

variable "members" {
  description = "Allocated number of members."
  type        = number
  default     = 2
  validation {
    condition = alltrue([
      var.members == 2
    ])
    error_message = "member group members must be >= 2 and <= 2 in increments of 1"
  }
}

variable "endpoints" {
  description = "Endpoints available to the redis instance (public, private, public-and-private)"
  type        = string
  default     = "private"
  validation {
    condition     = can(regex("public|public-and-private|private", var.endpoints))
    error_message = "Valid values for service_endpoints are 'public', 'public-and-private', and 'private'"
  }
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
  description = "(Optional) The root key CRN of a Key Management Service like Key Protect or Hyper Protect Crypto Service (HPCS) that you want to use for disk encryption. If `null`, database is encrypted by using randomly generated keys. See https://cloud.ibm.com/docs/cloud-databases?topic=cloud-databases-key-protect&interface=ui#key-byok for current list of supported regions for BYOK"
  default     = null
}

variable "backup_encryption_key_crn" {
  type        = string
  description = "(Optional) The CRN of a key protect key, that you want to use for encrypting disk that holds deployment backups. If null, will use 'key_protect_key_crn' as encryption key. If 'key_protect_key_crn' is also null database is encrypted by using randomly generated keys."
  default     = null
}

variable "auto_scaling" {
  type = object({
    cpu = object({
      rate_increase_percent       = optional(number, 10)
      rate_limit_count_per_member = optional(number, 20)
      rate_period_seconds         = optional(number, 900)
      rate_units                  = optional(string, "count")
    })
    disk = object({
      capacity_enabled             = optional(bool, false)
      free_space_less_than_percent = optional(number, 10)
      io_above_percent             = optional(number, 90)
      io_enabled                   = optional(bool, false)
      io_over_period               = optional(string, "15m")
      rate_increase_percent        = optional(number, 10)
      rate_limit_mb_per_member     = optional(number, 3670016)
      rate_period_seconds          = optional(number, 900)
      rate_units                   = optional(string, "mb")
    })
    memory = object({
      io_above_percent         = optional(number, 90)
      io_enabled               = optional(bool, false)
      io_over_period           = optional(string, "15m")
      rate_increase_percent    = optional(number, 10)
      rate_limit_mb_per_member = optional(number, 114688)
      rate_period_seconds      = optional(number, 900)
      rate_units               = optional(string, "mb")
    })
  })
  description = "(Optional) Configure rules to allow your database to automatically increase its resources. Single block of autoscaling is allowed at once."
  default     = null
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
