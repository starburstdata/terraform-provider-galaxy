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

resource "galaxy_postgresql_catalog" "example" {
  name          = "rpgcat${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = "galaxy_testing"
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = false
}

resource "galaxy_role" "example" {
  role_name              = "rpgrole${local.test_suffix}"
  role_description       = "Example role for privilege grants data source"
  grant_to_creating_role = true
}

resource "galaxy_role_privilege_grant" "example" {
  role_id      = galaxy_role.example.role_id
  entity_id    = galaxy_postgresql_catalog.example.catalog_id
  entity_kind  = "Catalog"
  privilege    = "CreateSchema"
  grant_kind   = "Allow"
  grant_option = false
}

# List all directly granted privileges for the role
data "galaxy_role_privilege_grants" "example" {
  role_id    = galaxy_role.example.role_id
  depends_on = [galaxy_role_privilege_grant.example]
}

# List all privileges including inherited ones
data "galaxy_role_privilege_grants" "all" {
  role_id             = galaxy_role.example.role_id
  list_all_privileges = true
  depends_on          = [galaxy_role_privilege_grant.example]
}

output "direct_grants" {
  value = data.galaxy_role_privilege_grants.example.result
}

output "all_grants_count" {
  value = length(data.galaxy_role_privilege_grants.all.result)
}

variable "TESTING_POSTGRESQL_AWS_HOST" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_USERNAME" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_PASSWORD" {
  type      = string
  sensitive = true
}
