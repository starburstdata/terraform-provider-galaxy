variable "TESTING_SNOWFLAKE_ACCOUNT_ID" {
  type        = string
  sensitive   = true
  description = "Testing Snowflake account ID from integration secrets"
  default     = ""
}

variable "TESTING_SNOWFLAKE_USER" {
  type        = string
  sensitive   = true
  description = "Testing Snowflake user from integration secrets"
  default     = ""
}

variable "TESTING_SNOWFLAKE_PASSWORD" {
  type        = string
  sensitive   = true
  description = "Testing Snowflake password from integration secrets"
  default     = ""
}

variable "TESTING_SNOWFLAKE_DATABASE" {
  type        = string
  sensitive   = true
  description = "Testing Snowflake database from integration secrets"
  default     = ""
}

variable "TESTING_SNOWFLAKE_WAREHOUSE" {
  type        = string
  sensitive   = true
  description = "Testing Snowflake warehouse from integration secrets"
  default     = ""
}

variable "TESTING_SNOWFLAKE_ROLE" {
  type        = string
  sensitive   = true
  description = "Testing Snowflake role from integration secrets"
  default     = ""
}