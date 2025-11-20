terraform {
  required_providers {
    galaxy = {
      source = "starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials are provided via environment variables
}

# Create a Redshift catalog
locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

variable "TESTING_REDSHIFT_ENDPOINT" {
  type        = string
  sensitive   = true
  description = "Testing Redshift endpoint from integration secrets"
  default     = ""
}

variable "TESTING_REDSHIFT_USER" {
  type        = string
  sensitive   = true
  description = "Testing Redshift user from integration secrets"
  default     = ""
}

variable "TESTING_REDSHIFT_PASSWORD" {
  type        = string
  sensitive   = true
  description = "Testing Redshift password from integration secrets"
  default     = ""
}

resource "galaxy_redshift_catalog" "test" {
  name        = "rscat${local.test_suffix}"
  endpoint    = var.TESTING_REDSHIFT_ENDPOINT
  username    = var.TESTING_REDSHIFT_USER
  password    = var.TESTING_REDSHIFT_PASSWORD
  auth_type   = "basic"
  read_only   = false
  description = "E2E testing Redshift catalog"
}

# Data source to read the catalog
data "galaxy_redshift_catalog" "test" {
  depends_on = [galaxy_redshift_catalog.test]
  catalog_id = galaxy_redshift_catalog.test.catalog_id
}

output "redshift_catalog_id" {
  value = galaxy_redshift_catalog.test.catalog_id
}
