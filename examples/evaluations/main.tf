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

resource "galaxy_postgresql_catalog" "example" {
  name          = "evlsds${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = false
  description   = "Catalog for evaluations data source example"
}

resource "galaxy_cluster" "example" {
  name                 = "evlsds${local.test_suffix}"
  cloud_region_id      = "aws-us-east1"
  catalog_refs         = [galaxy_postgresql_catalog.example.catalog_id]
  idle_stop_minutes    = 5
  min_workers          = 1
  max_workers          = 1
  result_cache_enabled = false
  private_link_cluster = false
}

resource "galaxy_role" "eval_grant" {
  role_name              = "evlsdsgrant${local.test_suffix}"
  role_description       = "Role granting access for evaluations example"
  grant_to_creating_role = true
}

resource "galaxy_role_privilege_grant" "eval_grant" {
  role_id      = galaxy_role.eval_grant.role_id
  entity_id    = galaxy_postgresql_catalog.example.catalog_id
  entity_kind  = "Column"
  privilege    = "Select"
  grant_kind   = "Allow"
  grant_option = false
  schema_name  = "*"
  table_name   = "*"
  column_name  = "*"
}

resource "galaxy_data_quality_check" "example" {
  name        = "evlsdscheck_${local.test_suffix}"
  description = "Data quality check for evaluations data source example"
  catalog_id  = galaxy_postgresql_catalog.example.catalog_id
  schema_id   = "anu_test"
  table_id    = "employees"
  severity    = "Low"
  category    = "Completeness"
  kind        = "SqlQuery"
  cluster_id  = galaxy_cluster.example.cluster_id
  query       = "SELECT EXISTS(SELECT * FROM ${galaxy_postgresql_catalog.example.name}.anu_test.employees)"
  depends_on  = [galaxy_role_privilege_grant.eval_grant]
}

# Read evaluation results for the data quality check
data "galaxy_evaluations" "example" {
  data_quality_check_id = galaxy_data_quality_check.example.data_quality_check_id
}

output "check_id" {
  value = data.galaxy_evaluations.example.data_quality_check_id
}

output "evaluations_count" {
  value = length(try(data.galaxy_evaluations.example.evaluations, []))
}
