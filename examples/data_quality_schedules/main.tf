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
  name          = "dqsched${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = false
  description   = "Catalog for data quality schedules data source example"
}

resource "galaxy_cluster" "example" {
  name                 = "dqsched${local.test_suffix}"
  cloud_region_id      = "aws-us-east1"
  catalog_refs         = [galaxy_postgresql_catalog.example.catalog_id]
  idle_stop_minutes    = 5
  min_workers          = 1
  max_workers          = 1
  result_cache_enabled = false
  private_link_cluster = false
}

resource "galaxy_role" "dq_schedule" {
  role_name              = "dqschedrole${local.test_suffix}"
  role_description       = "Role for data quality schedule example"
  grant_to_creating_role = true
}

resource "galaxy_role_privilege_grant" "dq_schedule" {
  role_id      = galaxy_role.dq_schedule.role_id
  entity_id    = galaxy_postgresql_catalog.example.catalog_id
  entity_kind  = "Column"
  privilege    = "Select"
  grant_kind   = "Allow"
  grant_option = false
  schema_name  = "*"
  table_name   = "*"
  column_name  = "*"
}

# Creating a cluster only grants the creator CREATE_CLUSTER on it - it does NOT grant
# USE_CLUSTER to the role that will execute checks/schedules on the cluster. Both
# galaxy_data_quality_check and galaxy_data_quality_schedule validate USE_CLUSTER on the
# cluster for the given role_id independently, so that role needs an explicit grant or
# creation fails with a 403 PERMISSION_DENIED (GET_CLUSTER) error.
resource "galaxy_role_privilege_grant" "dq_schedule_use_cluster" {
  role_id      = galaxy_role.dq_schedule.role_id
  entity_id    = galaxy_cluster.example.cluster_id
  entity_kind  = "Cluster"
  privilege    = "UseCluster"
  grant_kind   = "Allow"
  grant_option = false
}

resource "galaxy_data_quality_check" "example" {
  name        = "dqschedcheck_${local.test_suffix}"
  description = "Data quality check for schedule example"
  catalog_id  = galaxy_postgresql_catalog.example.catalog_id
  schema_id   = "anu_test"
  table_id    = "employees"
  severity    = "Low"
  category    = "Completeness"
  kind        = "SqlQuery"
  cluster_id  = galaxy_cluster.example.cluster_id
  query       = "SELECT EXISTS(SELECT * FROM ${galaxy_postgresql_catalog.example.name}.anu_test.employees)"
  depends_on  = [galaxy_role_privilege_grant.dq_schedule, galaxy_role_privilege_grant.dq_schedule_use_cluster]
}

resource "galaxy_data_quality_schedule" "example" {
  catalog_id      = galaxy_postgresql_catalog.example.catalog_id
  schema_id       = "anu_test"
  table_id        = "employees"
  cluster_id      = galaxy_cluster.example.cluster_id
  cron_expression = "0 0 * * *"
  role_id         = galaxy_role.dq_schedule.role_id
  depends_on      = [galaxy_data_quality_check.example, galaxy_role_privilege_grant.dq_schedule_use_cluster]
}

# Read the data quality schedule for the table
data "galaxy_data_quality_schedules" "example" {
  catalog_id = galaxy_postgresql_catalog.example.catalog_id
  schema_id  = "anu_test"
  table_id   = "employees"
  depends_on = [galaxy_data_quality_schedule.example]
}

output "schedule_id" {
  value = try(data.galaxy_data_quality_schedules.example.data_quality_schedule_id, "none")
}

output "schedule_enabled" {
  value = try(data.galaxy_data_quality_schedules.example.enabled, false)
}
