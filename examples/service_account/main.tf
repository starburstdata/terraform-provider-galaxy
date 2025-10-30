terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials from environment variables
}

# Use TEST_SUFFIX environment variable for unique naming
locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

# Create a role for the service account
resource "galaxy_role" "automation" {
  role_name              = "tfautomationrole${local.test_suffix}"
  grant_to_creating_role = true
}

# Create a service account for ETL processes
resource "galaxy_service_account" "etl" {
  username              = "tfetlsa${local.test_suffix}"
  with_initial_password = true

  additional_role_ids = [
    galaxy_role.automation.role_id
  ]
}

# Create another service account for reporting
resource "galaxy_service_account" "reporting" {
  username              = "tfreportingsa${local.test_suffix}"
  with_initial_password = true

  additional_role_ids = [
    galaxy_role.automation.role_id
  ]
}

# Get service account password (if needed)
resource "galaxy_service_account_password" "etl_password" {
  service_account_id = galaxy_service_account.etl.service_account_id
}

# Data source to read service account
data "galaxy_service_account" "etl" {
  depends_on = [galaxy_service_account.etl]
  service_account_id = galaxy_service_account.etl.service_account_id
}

output "etl_service_account_id" {
  value = galaxy_service_account.etl.service_account_id
}

output "etl_service_account_username" {
  value = galaxy_service_account.etl.username
}

output "etl_password_id" {
  value = galaxy_service_account_password.etl_password.service_account_password_id
}
