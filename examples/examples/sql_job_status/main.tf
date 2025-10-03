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

# Get all available SQL jobs first to find one to test with
data "galaxy_sql_jobs" "all" {}

# Local values for handling empty results with null safety
locals {
  sql_jobs_result = try(data.galaxy_sql_jobs.all.result, [])

  # Find first non-null SQL job
  first_sql_job = try([for job in local.sql_jobs_result : job if job != null][0], null)

  has_sql_jobs = local.first_sql_job != null
}

# Test the SQL job status data source (only if we have SQL jobs available)
data "galaxy_sql_job_status" "test" {
  count = local.has_sql_jobs ? 1 : 0
  id    = local.first_sql_job.sql_job_id
}

# Diagnostic outputs
output "sql_jobs_available" {
  value = local.has_sql_jobs
}

output "sql_jobs_count" {
  value = length(local.sql_jobs_result)
}

output "first_sql_job_debug" {
  value = {
    found      = local.first_sql_job != null
    sql_job_id = try(local.first_sql_job.sql_job_id, "none")
    name       = try(local.first_sql_job.name, "none")
  }
}

# SQL Job Status data source test outputs (conditional)
output "sql_job_status_test" {
  value = local.has_sql_jobs ? {
    id                  = data.galaxy_sql_job_status.test[0].id
    sql_job_id         = data.galaxy_sql_job_status.test[0].sql_job_id
    status             = data.galaxy_sql_job_status.test[0].status
    progress_percentage = data.galaxy_sql_job_status.test[0].progress_percentage
    query_id           = try(data.galaxy_sql_job_status.test[0].query_id, "")
    error_message      = try(data.galaxy_sql_job_status.test[0].error_message, "")
    updated_at         = try(data.galaxy_sql_job_status.test[0].updated_at, "")
    test_status        = "success"
  } : {
    id                  = "not_available"
    sql_job_id         = "No SQL jobs available for testing"
    status             = "none"
    progress_percentage = 0
    query_id           = "none"
    error_message      = "none"
    updated_at         = "none"
    test_status        = "failed"
  }
}

# Show the first few SQL jobs (up to 3) for debugging
output "sample_sql_jobs" {
  value = [
    for i, sql_job in slice(local.sql_jobs_result, 0, min(3, length(local.sql_jobs_result))) :
    {
      index      = i
      name       = try(sql_job.name, "unknown")
      sql_job_id = try(sql_job.sql_job_id, "unknown")
    }
  ]
}

# Additional status information (if status is available)
output "status_details" {
  value = local.has_sql_jobs ? {
    has_error           = try(data.galaxy_sql_job_status.test[0].error_message, "") != ""
    is_complete         = try(data.galaxy_sql_job_status.test[0].progress_percentage, 0) == 100
    has_query_id        = try(data.galaxy_sql_job_status.test[0].query_id, "") != ""
    last_updated_exists = try(data.galaxy_sql_job_status.test[0].updated_at, "") != ""
  } : {
    has_error           = false
    is_complete         = false
    has_query_id        = false
    last_updated_exists = false
  }
  description = "Additional status analysis"
}