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

# Get all available data quality summaries
data "galaxy_data_quality_summaries" "all" {}

# Local values for handling empty results with null safety
locals {
  catalog_summaries = try(data.galaxy_data_quality_summaries.all.catalog_summaries, [])
  category_counts   = try(data.galaxy_data_quality_summaries.all.category_counts, [])
  has_summaries     = length(local.catalog_summaries) > 0
}

# Diagnostic outputs
output "catalog_summaries_count" {
  value       = length(local.catalog_summaries)
  description = "Total number of catalog summaries"
}

output "category_counts_count" {
  value       = length(local.category_counts)
  description = "Total number of category counts"
}

output "has_summaries" {
  value       = local.has_summaries
  description = "Whether any catalog summaries exist"
}

# Show details of catalog summaries (or message if none exist)
output "catalog_summaries_details" {
  value = local.has_summaries ? [
    for summary in local.catalog_summaries : {
      catalog_id        = try(summary.catalog_id, "unknown")
      catalog_name      = try(summary.catalog_name, "unknown")
      evaluated_checks  = try(summary.evaluated_checks, 0)
      failed_checks     = try(summary.failed_checks, 0)
      successful_checks = try(summary.successful_checks, 0)
      number_of_checks  = try(summary.number_of_checks, 0)
    }
  ] : []
  description = "Details of all available catalog summaries"
}

# Show category counts
output "category_counts_details" {
  value = length(local.category_counts) > 0 ? [
    for count in local.category_counts : {
      category               = try(count.category, "unknown")
      failed_evaluations     = try(count.failed_evaluations, 0)
      successful_evaluations = try(count.successful_evaluations, 0)
      total_evaluations      = try(count.total_evaluations, 0)
    }
  ] : []
  description = "Details of category counts"
}

# Show first catalog summary details for reference
output "first_catalog_summary_info" {
  value = local.has_summaries ? {
    catalog_id        = try(local.catalog_summaries[0].catalog_id, "unknown")
    catalog_name      = try(local.catalog_summaries[0].catalog_name, "unknown")
    evaluated_checks  = try(local.catalog_summaries[0].evaluated_checks, 0)
    failed_checks     = try(local.catalog_summaries[0].failed_checks, 0)
    successful_checks = try(local.catalog_summaries[0].successful_checks, 0)
  } : {
    catalog_id        = "No summaries available"
    catalog_name      = "none"
    evaluated_checks  = 0
    failed_checks     = 0
    successful_checks = 0
  }
  description = "Information about the first catalog summary (if any exists)"
}

# Summary output
output "data_quality_summaries_summary" {
  value = {
    catalog_summaries_count = length(local.catalog_summaries)
    category_counts_count   = length(local.category_counts)
    has_summaries           = local.has_summaries
    unique_catalogs         = length(distinct([for summary in local.catalog_summaries : try(summary.catalog_id, "unknown")]))
    status                  = local.has_summaries ? "success" : "no_summaries_available"
  }
  description = "Summary of data quality summaries data source test"
}