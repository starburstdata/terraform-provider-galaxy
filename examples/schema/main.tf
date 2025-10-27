terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
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

# Test the schema data source (only if we have catalogs available)
data "galaxy_schema" "test" {
  count = local.has_catalogs ? 1 : 0
  id    = local.first_catalog.catalog_id
}

# Local processing of schema results
locals {
  schema_result = local.has_catalogs ? try(data.galaxy_schema.test[0].result, []) : []
  has_schemas   = length(local.schema_result) > 0
}

# Diagnostic outputs
output "catalogs_available" {
  value = local.has_catalogs
}

output "catalogs_count" {
  value = length(local.catalogs_result)
}

output "schemas_count" {
  value = length(local.schema_result)
}

output "first_catalog_debug" {
  value = {
    found      = local.first_catalog != null
    catalog_id = try(local.first_catalog.catalog_id, "none")
    name       = try(local.first_catalog.catalog_name, "none")
  }
}

# Schema data source test outputs (conditional)
output "schema_test" {
  value = local.has_catalogs ? {
    id            = data.galaxy_schema.test[0].id
    schemas_count = length(local.schema_result)
    has_schemas   = local.has_schemas
    test_status   = "success"
  } : {
    id            = "not_available"
    schemas_count = 0
    has_schemas   = false
    test_status   = "failed"
  }
}

# Show details of schemas (if available)
output "schemas_details" {
  value = local.has_schemas ? [
    for schema in local.schema_result : {
      schema_name = try(schema.schema_name, "unknown")
      owner       = try(schema.owner, "unknown")
      type        = try(schema.type, "unknown")
      contacts    = length(try(schema.contacts, []))
    }
  ] : []
  description = "Details of schemas in the catalog"
}

# Show first schema details for reference
output "first_schema_info" {
  value = local.has_schemas ? {
    schema_name = try(local.schema_result[0].schema_name, "unknown")
    owner       = try(local.schema_result[0].owner, "unknown")
    type        = try(local.schema_result[0].type, "unknown")
    contacts    = length(try(local.schema_result[0].contacts, []))
  } : {
    schema_name = "No schemas available"
    owner       = "none"
    type        = "none"
    contacts    = 0
  }
  description = "Information about the first schema (if any exists)"
}

# Summary statistics
output "schemas_summary" {
  value = local.has_schemas ? {
    total_schemas    = length(local.schema_result)
    unique_owners    = length(distinct([for schema in local.schema_result : try(schema.owner, "unknown")]))
    unique_types     = length(distinct([for schema in local.schema_result : try(schema.type, "unknown")]))
    schemas_with_contacts = length([for schema in local.schema_result : schema if length(try(schema.contacts, [])) > 0])
    status           = "success"
  } : {
    total_schemas    = 0
    unique_owners    = 0
    unique_types     = 0
    schemas_with_contacts = 0
    status           = "no_catalogs_available"
  }
  description = "Summary statistics of schemas"
}
