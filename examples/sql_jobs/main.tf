terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
}

# Get all available SQL jobs
data "galaxy_sql_jobs" "all" {}

# Local values for handling empty results with null safety
locals {
  sql_jobs_result = try(data.galaxy_sql_jobs.all.result, [])
  has_sql_jobs    = length(local.sql_jobs_result) > 0
}

# Diagnostic outputs
output "sql_jobs_count" {
  value       = length(local.sql_jobs_result)
  description = "Total number of SQL jobs"
}

output "has_sql_jobs" {
  value       = local.has_sql_jobs
  description = "Whether any SQL jobs exist"
}

# Show details of all SQL jobs (or message if none exist)
output "sql_jobs_details" {
  value = local.has_sql_jobs ? [
    for sql_job in local.sql_jobs_result : {
      name            = try(sql_job.name, "unknown")
      sql_job_id      = try(sql_job.sql_job_id, "unknown")
      description     = try(sql_job.description, "")
      role_id         = try(sql_job.role_id, "unknown")
      cron_expression = try(sql_job.cron_expression, "unknown")
      timezone        = try(sql_job.timezone, "UTC")
    }
  ] : []
  description = "Details of all available SQL jobs"
}

# Show first SQL job details for reference
output "first_sql_job_info" {
  value = local.has_sql_jobs ? {
    name            = try(local.sql_jobs_result[0].name, "unknown")
    sql_job_id      = try(local.sql_jobs_result[0].sql_job_id, "unknown")
    description     = try(local.sql_jobs_result[0].description, "")
    role_id         = try(local.sql_jobs_result[0].role_id, "unknown")
    cron_expression = try(local.sql_jobs_result[0].cron_expression, "unknown")
    timezone        = try(local.sql_jobs_result[0].timezone, "UTC")
  } : {
    name            = "No SQL jobs available"
    sql_job_id      = "none"
    description     = ""
    role_id         = "none"
    cron_expression = "none"
    timezone        = "none"
  }
  description = "Information about the first SQL job (if any exists)"
}

# Summary by job characteristics
output "sql_jobs_summary" {
  value = {
    total_count     = length(local.sql_jobs_result)
    has_sql_jobs    = local.has_sql_jobs
    unique_roles    = length(distinct([for job in local.sql_jobs_result : try(job.role_id, "unknown")]))
    unique_timezones = length(distinct([for job in local.sql_jobs_result : try(job.timezone, "UTC")]))
    status          = local.has_sql_jobs ? "success" : "no_sql_jobs_available"
  }
  description = "Summary of SQL jobs data source test"
}

# Statistics about cron expressions (if any jobs exist)
output "cron_expressions_stats" {
  value = local.has_sql_jobs ? {
    unique_expressions = length(distinct([for job in local.sql_jobs_result : try(job.cron_expression, "unknown")]))
    sample_expressions = slice(distinct([for job in local.sql_jobs_result : try(job.cron_expression, "unknown")]), 0, min(3, length(distinct([for job in local.sql_jobs_result : try(job.cron_expression, "unknown")]))))
  } : {
    unique_expressions = 0
    sample_expressions = []
  }
  description = "Statistics about cron expressions used in SQL jobs"
}
