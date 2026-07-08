terraform {
  required_providers {
    galaxy = {
      source = "starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials from environment variables
}

locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

variable "TESTING_POSTGRESQL_AWS_HOST" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_DATABASE" {
  type = string
}

variable "TESTING_POSTGRESQL_AWS_USERNAME" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_PASSWORD" {
  type      = string
  sensitive = true
}

# Create a catalog for testing
resource "galaxy_postgresql_catalog" "test" {
  name          = "dqsc${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = false
  description   = "Catalog for data quality schedule example"
}

# Read data quality schedule for the table
data "galaxy_data_quality_schedule" "test" {
  catalog_id = galaxy_postgresql_catalog.test.catalog_id
  schema_id  = "information_schema"
  table_id   = "tables"
}

output "schedule_id" {
  value = try(data.galaxy_data_quality_schedule.test.data_quality_schedule_id, "none")
}

output "schedule_enabled" {
  value = try(data.galaxy_data_quality_schedule.test.enabled, false)
}
