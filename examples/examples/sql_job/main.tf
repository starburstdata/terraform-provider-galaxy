terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  domain        = var.galaxy_domain
  client_id     = var.galaxy_client_id
  client_secret = var.galaxy_client_secret
}

# Get available clusters and roles for testing
data "galaxy_clusters" "all" {}
data "galaxy_roles" "all" {}

# Local values for handling empty results with null safety
locals {
  clusters_result = try(data.galaxy_clusters.all.result, [])
  roles_result    = try(data.galaxy_roles.all.result, [])

  # Find first non-null cluster and role
  first_cluster = try([for c in local.clusters_result : c if c != null][0], null)
  first_role    = try([for r in local.roles_result : r if r != null][0], null)

  has_clusters = local.first_cluster != null
  has_roles    = local.first_role != null
  can_create_sql_job = local.has_clusters && local.has_roles
}

# Create a SQL job only if we have clusters and roles available
resource "galaxy_sql_job" "example" {
  count = local.can_create_sql_job ? 1 : 0

  name            = "example-sql-job-${var.test_suffix}"
  description     = "Example SQL job for testing"
  cluster_id      = local.first_cluster.cluster_id
  role_id         = local.first_role.role_id
  query           = "SELECT 1 as test_column"
  cron_expression = "0 */6 * * *"  # Every 6 hours
  timezone        = "UTC"
}

# Test the data source (only if SQL job was created)
data "galaxy_sql_job" "test" {
  count = local.can_create_sql_job ? 1 : 0
  id    = galaxy_sql_job.example[0].sql_job_id

  depends_on = [galaxy_sql_job.example]
}

# Diagnostic outputs
output "clusters_available" {
  value = local.has_clusters
}

output "roles_available" {
  value = local.has_roles
}

output "clusters_count" {
  value = length(local.clusters_result)
}

output "roles_count" {
  value = length(local.roles_result)
}

output "sql_job_created" {
  value = local.can_create_sql_job
}

# Debug outputs
output "first_cluster_debug" {
  value = {
    found = local.first_cluster != null
    cluster_id = try(local.first_cluster.cluster_id, "none")
  }
}

output "first_role_debug" {
  value = {
    found = local.first_role != null
    role_id = try(local.first_role.role_id, "none")
  }
}

# SQL Job outputs (conditional)
output "sql_job_id" {
  value = local.can_create_sql_job ? galaxy_sql_job.example[0].sql_job_id : "No SQL job created - missing clusters or roles"
}

output "sql_job_name" {
  value = local.can_create_sql_job ? galaxy_sql_job.example[0].name : "No SQL job created - missing clusters or roles"
}

output "data_source_test" {
  value = local.can_create_sql_job ? {
    name        = data.galaxy_sql_job.test[0].name
    description = data.galaxy_sql_job.test[0].description
    query       = data.galaxy_sql_job.test[0].query
    status      = "success"
  } : {
    name        = "not_created"
    description = "No SQL job created - missing clusters or roles"
    query       = "none"
    status      = "failed"
  }
}