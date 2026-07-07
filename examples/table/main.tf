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

  test_schema_id = "information_schema"
  test_table_id  = "tables"
}

# Example: Reading a table data source
# Note: Replace test schema and table IDs with actual values that exist in your catalog
# Set count = 1 to enable this data source
data "galaxy_table" "test" {
  count      = 0
  catalog_id = local.has_catalogs ? local.first_catalog.catalog_id : ""
  schema_id  = local.test_schema_id
  table_id   = local.test_table_id
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

output "table_example_note" {
  value = "To test the table data source, update test_schema_id and test_table_id to actual values in your catalog and set count=1"
}
