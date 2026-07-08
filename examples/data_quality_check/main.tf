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
  name          = "dqex${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = false
  description   = "Catalog for data quality check example"
}

# Create a cluster for validating SQL checks
resource "galaxy_cluster" "test" {
  name                 = "dqex${local.test_suffix}"
  cloud_region_id      = "aws-us-east1"
  catalog_refs         = [galaxy_postgresql_catalog.test.catalog_id]
  idle_stop_minutes    = 5
  min_workers          = 1
  max_workers          = 1
  result_cache_enabled = false
  private_link_cluster = false
}

# Grant the cluster access to the catalog
resource "galaxy_role" "dq_check" {
  role_name              = "dqrole_${local.test_suffix}"
  role_description       = "Role granting query access to the data quality check catalog"
  grant_to_creating_role = true
}

resource "galaxy_role_privilege_grant" "dq_check" {
  role_id      = galaxy_role.dq_check.role_id
  entity_id    = galaxy_postgresql_catalog.test.catalog_id
  entity_kind  = "Column"
  privilege    = "Select"
  grant_kind   = "Allow"
  grant_option = false
  schema_name  = "*"
  table_name   = "*"
  column_name  = "*"
}

# Create a data quality check
resource "galaxy_data_quality_check" "example" {
  name        = "dq_check_${local.test_suffix}"
  description = "Example data quality check"
  catalog_id  = galaxy_postgresql_catalog.test.catalog_id
  schema_id   = "anu_test"
  table_id    = "employees"
  query       = "SELECT EXISTS(SELECT * FROM ${galaxy_postgresql_catalog.test.name}.anu_test.employees)"
  kind        = "SqlQuery"
  category    = "Validity"
  severity    = "Low"
  cluster_id  = galaxy_cluster.test.cluster_id
  depends_on  = [galaxy_role_privilege_grant.dq_check]
}

# Read back the data quality check via data source
data "galaxy_data_quality_check" "example" {
  data_quality_check_id = galaxy_data_quality_check.example.data_quality_check_id
}

output "data_quality_check_id" {
  value = galaxy_data_quality_check.example.data_quality_check_id
}

output "data_quality_check_name" {
  value = data.galaxy_data_quality_check.example.name
}
