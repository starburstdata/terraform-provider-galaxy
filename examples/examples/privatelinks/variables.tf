variable "galaxy_domain" {
  description = "Galaxy domain URL"
  type        = string
  default     = ""
}

variable "galaxy_client_id" {
  description = "Galaxy OAuth2 client ID"
  type        = string
  default     = ""
}

variable "galaxy_client_secret" {
  description = "Galaxy OAuth2 client secret"
  type        = string
  sensitive   = true
  default     = ""
}

variable "test_suffix" {
  description = "Suffix for test resources to avoid conflicts"
  type        = string
  default     = "local"
}