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

# Use timestamps for unique naming
locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

# Create a PostgreSQL catalog first (needed for privilege grants)
resource "galaxy_postgresql_catalog" "example" {
  name          = "privcat${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = "galaxy_testing"
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  description   = "Catalog for privilege grant testing"
  read_only     = false
}

# Create a role to grant privileges to
resource "galaxy_role" "example" {
  role_name              = "examplerole${local.test_suffix}"
  role_description       = "Example role for privilege testing"
  grant_to_creating_role = true
}

# Grant privileges on the catalog to the role
resource "galaxy_role_privilege_grant" "catalog_grant" {
  role_id      = galaxy_role.example.role_id
  entity_id    = galaxy_postgresql_catalog.example.catalog_id
  entity_kind  = "Catalog"
  privilege    = "CreateSchema"
  grant_kind   = "Allow"
  grant_option = false
}

output "catalog_id" {
  value = galaxy_postgresql_catalog.example.catalog_id
}

output "example_role_id" {
  value = galaxy_role.example.role_id
}

output "privilege_grant_info" {
  value = {
    entity_id   = galaxy_role_privilege_grant.catalog_grant.entity_id
    entity_kind = galaxy_role_privilege_grant.catalog_grant.entity_kind
    privilege   = galaxy_role_privilege_grant.catalog_grant.privilege
  }
}

# Variable declarations
variable "TESTING_POSTGRESQL_AWS_HOST" {
  type        = string
  sensitive   = true
  description = "PostgreSQL host for testing"
}

variable "TESTING_POSTGRESQL_AWS_USERNAME" {
  type        = string
  sensitive   = true
  description = "PostgreSQL username for testing"
}

variable "TESTING_POSTGRESQL_AWS_PASSWORD" {
  type        = string
  sensitive   = true
  description = "PostgreSQL password for testing"
}
