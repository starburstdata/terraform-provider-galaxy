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

  # Use common schema and table names that often exist
  test_schema_id = "information_schema"
  test_table_id  = "columns" # information_schema.columns is a common system table
}

# Test the column data source (only if we have catalogs available)
data "galaxy_column" "test" {
  count      = local.has_catalogs ? 1 : 0
  catalog_id = local.first_catalog.catalog_id
  schema_id  = local.test_schema_id
  table_id   = local.test_table_id
}

# Local processing of column results
locals {
  column_result = local.has_catalogs ? try(data.galaxy_column.test[0].result, []) : []
  # Ensure column_result is never null by providing a fallback
  safe_column_result = local.column_result != null ? local.column_result : []
  has_columns       = length(local.safe_column_result) > 0
}

# Diagnostic outputs
output "catalogs_available" {
  value = local.has_catalogs
}

output "catalogs_count" {
  value = length(local.catalogs_result)
}

output "columns_count" {
  value = length(local.safe_column_result)
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
    table_id   = local.test_table_id
  }
  description = "Parameters used for testing column data source"
}

# Column data source test outputs (conditional)
output "column_test" {
  value = local.has_catalogs ? {
    catalog_id   = data.galaxy_column.test[0].catalog_id
    schema_id    = data.galaxy_column.test[0].schema_id
    table_id     = data.galaxy_column.test[0].table_id
    columns_count = length(local.safe_column_result)
    has_columns  = local.has_columns
    test_status  = "success"
  } : {
    catalog_id   = "not_available"
    schema_id    = "No catalogs available for testing"
    table_id     = "none"
    columns_count = 0
    has_columns  = false
    test_status  = "failed"
  }
}

# Show details of columns (if available)
output "columns_details" {
  value = local.has_columns ? [
    for column in local.safe_column_result : {
      column_id   = try(column.column_id, "unknown")
      data_type   = try(column.data_type, "unknown")
      nullable    = try(column.nullable, false)
      description = try(column.description, "")
      tags_count  = length(try(column.tags, []))
    }
  ] : []
  description = "Details of columns in the table"
}

# Show first column details for reference
output "first_column_info" {
  value = local.has_columns ? {
    column_id   = try(local.safe_column_result[0].column_id, "unknown")
    data_type   = try(local.safe_column_result[0].data_type, "unknown")
    nullable    = try(local.safe_column_result[0].nullable, false)
    description = try(local.safe_column_result[0].description, "")
    tags_count  = length(try(local.safe_column_result[0].tags, []))
  } : {
    column_id   = "No columns available"
    data_type   = "none"
    nullable    = false
    description = ""
    tags_count  = 0
  }
  description = "Information about the first column (if any exists)"
}

# Summary statistics
output "columns_summary" {
  value = local.has_columns ? {
    total_columns       = length(local.safe_column_result)
    unique_data_types   = length(distinct([for column in local.safe_column_result : try(column.data_type, "unknown")]))
    nullable_columns    = length([for column in local.safe_column_result : column if try(column.nullable, false) == true])
    columns_with_tags   = length([for column in local.safe_column_result : column if length(try(column.tags, [])) > 0])
    status             = "success"
  } : {
    total_columns       = 0
    unique_data_types   = 0
    nullable_columns    = 0
    columns_with_tags   = 0
    status             = "no_catalogs_available"
  }
  description = "Summary statistics of columns"
}