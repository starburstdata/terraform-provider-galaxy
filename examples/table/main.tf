terraform {
  required_providers {
    galaxy = {
      source = "starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
}

# Get all available catalogs first to find one to test with
data "galaxy_catalogs" "all" {}

# Local values for handling empty results with null safety
locals {
  catalogs_result = try(data.galaxy_catalogs.all.result, [])

  # Find first non-null catalog
  first_catalog = try([for catalog in local.catalogs_result : catalog if catalog != null][0], null)

  has_catalogs = local.first_catalog != null

  # Use a common schema ID for testing - information_schema is typically available
  test_schema_id = "information_schema"
}

# Test the table data source (only if we have catalogs available)
data "galaxy_table" "test" {
  count      = local.has_catalogs ? 1 : 0
  catalog_id = local.first_catalog.catalog_id
  schema_id  = local.test_schema_id
}

# Local processing of table results
locals {
  table_result = local.has_catalogs ? try(data.galaxy_table.test[0].result, []) : []
  has_tables   = length(local.table_result) > 0
}

# Diagnostic outputs
output "catalogs_available" {
  value = local.has_catalogs
}

output "catalogs_count" {
  value = length(local.catalogs_result)
}

output "tables_count" {
  value = length(local.table_result)
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
  description = "Parameters used for testing table data source"
}

# Table data source test outputs (conditional)
output "table_test" {
  value = local.has_catalogs ? {
    catalog_id   = data.galaxy_table.test[0].catalog_id
    schema_id    = data.galaxy_table.test[0].schema_id
    tables_count = length(local.table_result)
    has_tables   = local.has_tables
    test_status  = "success"
  } : {
    catalog_id   = "not_available"
    schema_id    = "No catalogs available for testing"
    tables_count = 0
    has_tables   = false
    test_status  = "failed"
  }
}

# Show details of tables (if available)
output "tables_details" {
  value = local.has_tables ? [
    for table in local.table_result : {
      table_name = try(table.table_name, "unknown")
      owner      = try(table.owner, "unknown")
      type       = try(table.type, "unknown")
      contacts   = length(try(table.contacts, []))
    }
  ] : []
  description = "Details of tables in the schema"
}

# Show first table details for reference
output "first_table_info" {
  value = local.has_tables ? {
    table_name = try(local.table_result[0].table_name, "unknown")
    owner      = try(local.table_result[0].owner, "unknown")
    type       = try(local.table_result[0].type, "unknown")
    contacts   = length(try(local.table_result[0].contacts, []))
  } : {
    table_name = "No tables available"
    owner      = "none"
    type       = "none"
    contacts   = 0
  }
  description = "Information about the first table (if any exists)"
}

# Summary statistics
output "tables_summary" {
  value = local.has_tables ? {
    total_tables         = length(local.table_result)
    unique_owners        = length(distinct([for table in local.table_result : try(table.owner, "unknown")]))
    unique_types         = length(distinct([for table in local.table_result : try(table.type, "unknown")]))
    tables_with_contacts = length([for table in local.table_result : table if length(try(table.contacts, [])) > 0])
    status              = "success"
  } : {
    total_tables         = 0
    unique_owners        = 0
    unique_types         = 0
    tables_with_contacts = 0
    status              = "no_catalogs_available"
  }
  description = "Summary statistics of tables"
}
