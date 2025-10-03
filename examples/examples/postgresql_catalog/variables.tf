variable "TESTING_POSTGRESQL_AWS_HOST" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL AWS host from integration secrets"
  default     = ""
}

variable "TESTING_POSTGRESQL_AWS_DATABASE" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL AWS database from integration secrets"
  default     = ""
}

variable "TESTING_POSTGRESQL_AWS_USERNAME" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL AWS username from integration secrets"
  default     = ""
}

variable "TESTING_POSTGRESQL_AWS_PASSWORD" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL AWS password from integration secrets"
  default     = ""
}