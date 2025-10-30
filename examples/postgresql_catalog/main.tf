terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials are provided via environment variables
}

# Create a PostgreSQL catalog
locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

resource "galaxy_postgresql_catalog" "test" {
  name          = "pgcat${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = false
  description   = "E2E testing PostgreSQL catalog"
}

# Data source to read the catalog
data "galaxy_postgresql_catalog" "test" {
  depends_on = [galaxy_postgresql_catalog.test]
  catalog_id = galaxy_postgresql_catalog.test.catalog_id
}

data "galaxy_postgresql_catalog_validation" "test" {
  depends_on = [galaxy_postgresql_catalog.test]
  id = galaxy_postgresql_catalog.test.catalog_id
}

output "postgresql_catalog_id" {
  value = galaxy_postgresql_catalog.test.catalog_id
}

output "postgresql_catalog_validation" {
  value = data.galaxy_postgresql_catalog_validation.test.validation_successful
}
