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

# Test the SQL job history data source (only if we have SQL jobs available)
data "galaxy_sql_job_history" "test" {
  count = local.has_sql_jobs ? 1 : 0
  id    = local.first_sql_job.sql_job_id
}

# Local processing of history results
locals {
  history_result = local.has_sql_jobs ? try(data.galaxy_sql_job_history.test[0].result, []) : []
  has_history    = length(local.history_result) > 0
}

# Diagnostic outputs
output "sql_jobs_available" {
  value = local.has_sql_jobs
}

output "sql_jobs_count" {
  value = length(local.sql_jobs_result)
}

output "history_entries_count" {
  value = length(local.history_result)
}

output "first_sql_job_debug" {
  value = {
    found      = local.first_sql_job != null
    sql_job_id = try(local.first_sql_job.sql_job_id, "none")
    name       = try(local.first_sql_job.name, "none")
  }
}

# SQL Job History data source test outputs (conditional)
output "sql_job_history_test" {
  value = local.has_sql_jobs ? {
    id              = data.galaxy_sql_job_history.test[0].id
    history_count   = length(local.history_result)
    has_history     = local.has_history
    test_status     = "success"
  } : {
    id              = "not_available"
    history_count   = 0
    has_history     = false
    test_status     = "failed"
  }
}

# Show details of history entries (if available)
output "history_details" {
  value = local.has_history ? [
    for entry in local.history_result : {
      query_id            = try(entry.query_id, "unknown")
      status              = try(entry.status, "unknown")
      progress_percentage = try(entry.progress_percentage, 0)
      started_at          = try(entry.started_at, "")
      updated_at          = try(entry.updated_at, "")
      has_error           = try(entry.error_message, "") != ""
      has_query           = try(entry.query, "") != ""
    }
  ] : []
  description = "Details of SQL job history entries"
}

# Show first history entry details for reference
output "first_history_entry" {
  value = local.has_history ? {
    query_id            = try(local.history_result[0].query_id, "unknown")
    status              = try(local.history_result[0].status, "unknown")
    progress_percentage = try(local.history_result[0].progress_percentage, 0)
    started_at          = try(local.history_result[0].started_at, "")
    updated_at          = try(local.history_result[0].updated_at, "")
    error_message       = try(local.history_result[0].error_message, "")
    query_preview       = try(substr(local.history_result[0].query, 0, 100), "")
  } : {
    query_id            = "No history available"
    status              = "none"
    progress_percentage = 0
    started_at          = "none"
    updated_at          = "none"
    error_message       = "none"
    query_preview       = "none"
  }
  description = "Information about the first history entry (if any exists)"
}

# Summary statistics
output "history_summary" {
  value = local.has_history ? {
    total_entries       = length(local.history_result)
    unique_statuses     = length(distinct([for entry in local.history_result : try(entry.status, "unknown")]))
    completed_entries   = length([for entry in local.history_result : entry if try(entry.progress_percentage, 0) == 100])
    entries_with_errors = length([for entry in local.history_result : entry if try(entry.error_message, "") != ""])
    status             = "success"
  } : {
    total_entries       = 0
    unique_statuses     = 0
    completed_entries   = 0
    entries_with_errors = 0
    status             = "no_history_available"
  }
  description = "Summary of SQL job history statistics"
}