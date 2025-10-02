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

# Get all available catalogs first to find one to test with
data "galaxy_catalogs" "all" {}

# Local values for handling empty results with null safety
locals {
  catalogs_result = try(data.galaxy_catalogs.all.result, [])

  # Find first non-null catalog
  first_catalog = try([for catalog in local.catalogs_result : catalog if catalog != null][0], null)

  has_catalogs = local.first_catalog != null

  # Use a hardcoded schema ID for testing - data quality is typically on specific schemas
  test_schema_id = "information_schema" # Common schema that often exists
}

# Test the data quality summary data source (only if we have catalogs available)
data "galaxy_data_quality_summary" "test" {
  count      = local.has_catalogs ? 1 : 0
  catalog_id = local.first_catalog.catalog_id
  schema_id  = local.test_schema_id
}

# Diagnostic outputs
output "catalogs_available" {
  value = local.has_catalogs
}

output "catalogs_count" {
  value = length(local.catalogs_result)
}

output "first_catalog_debug" {
  value = {
    found      = local.first_catalog != null
    catalog_id = try(local.first_catalog.catalog_id, "none")
    name       = try(local.first_catalog.catalog_name, "none")
  }
}

output "test_parameters" {
  value = {
    catalog_id = local.has_catalogs ? local.first_catalog.catalog_id : "none"
    schema_id  = local.test_schema_id
  }
  description = "Parameters used for testing data quality summary"
}

# Data Quality Summary data source test outputs (conditional)
output "data_quality_summary_test" {
  value = local.has_catalogs ? {
    catalog_id           = data.galaxy_data_quality_summary.test[0].catalog_id
    schema_id           = data.galaxy_data_quality_summary.test[0].schema_id
    category_counts     = length(try(data.galaxy_data_quality_summary.test[0].category_counts, []))
    daily_summaries     = length(try(data.galaxy_data_quality_summary.test[0].daily_summaries, []))
    severity_counts     = length(try(data.galaxy_data_quality_summary.test[0].severity_counts, []))
    test_status         = "success"
  } : {
    catalog_id           = "not_available"
    schema_id           = "No catalogs available for testing"
    category_counts     = 0
    daily_summaries     = 0
    severity_counts     = 0
    test_status         = "failed"
  }
}

# Show sample category counts (if available)
output "sample_category_counts" {
  value = local.has_catalogs ? [
    for i, category in slice(try(data.galaxy_data_quality_summary.test[0].category_counts, []), 0, min(3, length(try(data.galaxy_data_quality_summary.test[0].category_counts, [])))) :
    {
      index                   = i
      category               = try(category.category, "unknown")
      total_evaluations      = try(category.total_evaluations, 0)
      successful_evaluations = try(category.successful_evaluations, 0)
      failed_evaluations     = try(category.failed_evaluations, 0)
    }
  ] : []
  description = "Sample category counts from data quality summary"
}

# Show sample daily summaries (if available)
output "sample_daily_summaries" {
  value = local.has_catalogs ? [
    for i, daily in slice(try(data.galaxy_data_quality_summary.test[0].daily_summaries, []), 0, min(3, length(try(data.galaxy_data_quality_summary.test[0].daily_summaries, [])))) :
    {
      index                   = i
      day                    = try(daily.day, "unknown")
      total_evaluations      = try(daily.total_evaluations, 0)
      successful_evaluations = try(daily.successful_evaluations, 0)
      failed_evaluations     = try(daily.failed_evaluations, 0)
    }
  ] : []
  description = "Sample daily summaries from data quality summary"
}

# Summary statistics
output "data_quality_summary_stats" {
  value = local.has_catalogs ? {
    has_data_quality_info   = length(try(data.galaxy_data_quality_summary.test[0].category_counts, [])) > 0
    categories_tracked      = length(try(data.galaxy_data_quality_summary.test[0].category_counts, []))
    days_with_data         = length(try(data.galaxy_data_quality_summary.test[0].daily_summaries, []))
    severity_levels        = length(try(data.galaxy_data_quality_summary.test[0].severity_counts, []))
    status                 = "success"
  } : {
    has_data_quality_info   = false
    categories_tracked      = 0
    days_with_data         = 0
    severity_levels        = 0
    status                 = "no_catalogs_available"
  }
  description = "Summary statistics of data quality information"
}