variable "TESTING_GCS_JSON_KEY" {
  type        = string
  sensitive   = true
  description = "Testing GCS JSON key from integration secrets"
  default     = ""
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}