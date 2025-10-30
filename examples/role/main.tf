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

# Create different roles for different access levels
resource "galaxy_role" "read_only" {
  role_name              = "readonly${local.test_suffix}"
  grant_to_creating_role = true
  role_description       = "Role with read-only access to data"
}

resource "galaxy_role" "data_engineer" {
  role_name              = "dataeng${local.test_suffix}"
  grant_to_creating_role = true
  role_description       = "Role for data engineers with write access"
}

resource "galaxy_role" "admin" {
  role_name              = "admin${local.test_suffix}"
  grant_to_creating_role = true
  role_description       = "Administrative role with full permissions"
}

output "admin_role_id" {
  value = galaxy_role.admin.role_id
}

output "admin_role_name" {
  value = galaxy_role.admin.role_name
}

output "data_engineer_role_id" {
  value = galaxy_role.data_engineer.role_id
}

output "read_only_role_id" {
  value = galaxy_role.read_only.role_id
}

# Data source example for role
data "galaxy_role" "existing_admin" {
  role_id = galaxy_role.admin.role_id
}

output "admin_role_data" {
  value = data.galaxy_role.existing_admin
}
