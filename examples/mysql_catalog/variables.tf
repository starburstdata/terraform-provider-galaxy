variable "TESTING_MYSQL_AWS_HOST" {
  type        = string
  sensitive   = true
  description = "Testing MySQL AWS host from integration secrets"
  default     = ""
}

variable "TESTING_MYSQL_AWS_USERNAME" {
  type        = string
  sensitive   = true
  description = "Testing MySQL AWS username from integration secrets"
  default     = ""
}

variable "TESTING_MYSQL_AWS_PASSWORD" {
  type        = string
  sensitive   = true
  description = "Testing MySQL AWS password from integration secrets"
  default     = ""
}
variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}
