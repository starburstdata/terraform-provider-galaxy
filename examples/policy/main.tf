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
  timestamp = formatdate("YYYYMMDDHHmmss", timestamp())
}

# Create a PostgreSQL catalog (needed for policy scopes)
resource "galaxy_postgresql_catalog" "example" {
  name          = "polcat${local.timestamp}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  description   = "Catalog for policy testing"
  read_only     = false
}

# Create a role for the policy
resource "galaxy_role" "example" {
  role_name              = "polrole${local.timestamp}"
  role_description       = "Role for policy testing"
  grant_to_creating_role = true
}

# Create a row filter first (needed for policy)
resource "galaxy_row_filter" "example" {
  name        = "rowfilter${local.timestamp}"
  description = "Row filter for customer data security"

  # SQL filter expression
  expression = "region = 'US'"
}

# Create a column mask (needed for policy)
resource "galaxy_column_mask" "example" {
  name             = "columnmask${local.timestamp}"
  description      = "Column mask for data obfuscation"
  column_mask_type = "Varchar"

  # Expression to mask sensitive data
  expression = "'***MASKED***'"
}

# Create a policy that uses the row filter
resource "galaxy_policy" "row_security" {
  name        = "rowsecurity${local.timestamp}"
  description = "Row-level security policy for customer data"

  # SQL predicate for policy logic - must be valid SQL expression
  predicate = "true"

  # Role that this policy applies to
  role_id = galaxy_role.example.role_id

  # Scopes define what entities this policy applies to
  scopes = [
    {
      entity_id       = galaxy_postgresql_catalog.example.catalog_id
      entity_kind     = "Column"
      row_filter_ids  = [galaxy_row_filter.example.row_filter_id]
      column_mask_ids = [galaxy_column_mask.example.column_mask_id]
      schema_name     = "*"
      table_name      = "*"
      column_name     = "*"
    }
  ]
}

# Data source to read the policy
data "galaxy_policy" "example" {
  depends_on = [galaxy_policy.row_security]
  policy_id = galaxy_policy.row_security.policy_id
}

output "catalog_id" {
  value = galaxy_postgresql_catalog.example.catalog_id
}

output "role_id" {
  value = galaxy_role.example.role_id
}

output "row_filter_id" {
  value = galaxy_row_filter.example.row_filter_id
}

output "column_mask_id" {
  value = galaxy_column_mask.example.column_mask_id
}

output "policy_id" {
  value = galaxy_policy.row_security.policy_id
}

output "policy_predicate" {
  value = galaxy_policy.row_security.predicate
}

output "policy_data" {
  value = data.galaxy_policy.example
}

# Variable declarations
variable "TESTING_POSTGRESQL_AWS_HOST" {
  type        = string
  sensitive   = true
  description = "PostgreSQL host for testing"
}

variable "TESTING_POSTGRESQL_AWS_DATABASE" {
  type        = string
  sensitive   = true
  description = "PostgreSQL database for testing"
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
