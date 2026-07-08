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
}

# Example: Reading a schema data source
# Note: Replace "information_schema" with an actual schema that exists in your catalog
# To find available schemas, users can use the Galaxy API directly
# Set count = 1 to enable this data source with an actual schema ID
data "galaxy_schema" "test" {
  count      = 0
  catalog_id = local.has_catalogs ? local.first_catalog.catalog_id : ""
  schema_id  = "information_schema"
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

output "schema_example_note" {
  value = "To test the schema data source, update the schema_id to an actual schema in your catalog and set count=1"
}
