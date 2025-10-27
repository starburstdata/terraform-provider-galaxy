terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
}

# Get all available roles first to find one to test with
data "galaxy_roles" "all" {}

# Local values for handling empty results with null safety
locals {
  roles_result = try(data.galaxy_roles.all.result, [])

  # Find first non-null role
  first_role = try([for role in local.roles_result : role if role != null][0], null)

  has_roles = local.first_role != null
}

# Test the rolegrant data source (only if we have roles available)
data "galaxy_rolegrant" "test" {
  count = local.has_roles ? 1 : 0
  id    = local.first_role.role_id
}

# Local processing of rolegrant results
locals {
  rolegrant_result = local.has_roles ? try(data.galaxy_rolegrant.test[0].result, []) : []
  has_rolegrants   = length(local.rolegrant_result) > 0
}

# Diagnostic outputs
output "roles_available" {
  value = local.has_roles
}

output "roles_count" {
  value = length(local.roles_result)
}

output "rolegrants_count" {
  value = length(local.rolegrant_result)
}

output "first_role_debug" {
  value = {
    found   = local.first_role != null
    role_id = try(local.first_role.role_id, "none")
    name    = try(local.first_role.role_name, "none")
  }
}

# Rolegrant data source test outputs (conditional)
output "rolegrant_test" {
  value = local.has_roles ? {
    id               = data.galaxy_rolegrant.test[0].id
    rolegrants_count = length(local.rolegrant_result)
    has_rolegrants   = local.has_rolegrants
    test_status      = "success"
  } : {
    id               = "not_available"
    rolegrants_count = 0
    has_rolegrants   = false
    test_status      = "failed"
  }
}

# Show details of role grants (if available)
output "rolegrants_details" {
  value = local.has_rolegrants ? [
    for grant in local.rolegrant_result : {
      role_privilege_grant_id = try(grant.role_privilege_grant_id, "unknown")
      privilege               = try(grant.privilege, "unknown")
      catalog_name           = try(grant.catalog_name, "")
      schema_name            = try(grant.schema_name, "")
      table_name             = try(grant.table_name, "")
      grantee_name           = try(grant.grantee_name, "unknown")
    }
  ] : []
  description = "Details of role grants"
}

# Show first role grant details for reference
output "first_rolegrant_info" {
  value = local.has_rolegrants ? {
    role_privilege_grant_id = try(local.rolegrant_result[0].role_privilege_grant_id, "unknown")
    privilege               = try(local.rolegrant_result[0].privilege, "unknown")
    catalog_name           = try(local.rolegrant_result[0].catalog_name, "")
    schema_name            = try(local.rolegrant_result[0].schema_name, "")
    table_name             = try(local.rolegrant_result[0].table_name, "")
    grantee_name           = try(local.rolegrant_result[0].grantee_name, "unknown")
  } : {
    role_privilege_grant_id = "No role grants available"
    privilege               = "none"
    catalog_name           = "none"
    schema_name            = "none"
    table_name             = "none"
    grantee_name           = "none"
  }
  description = "Information about the first role grant (if any exists)"
}

# Summary statistics
output "rolegrants_summary" {
  value = local.has_rolegrants ? {
    total_grants         = length(local.rolegrant_result)
    unique_privileges    = length(distinct([for grant in local.rolegrant_result : try(grant.privilege, "unknown")]))
    unique_catalogs      = length(distinct([for grant in local.rolegrant_result : try(grant.catalog_name, "")]))
    unique_grantees      = length(distinct([for grant in local.rolegrant_result : try(grant.grantee_name, "unknown")]))
    catalog_level_grants = length([for grant in local.rolegrant_result : grant if try(grant.catalog_name, "") != "" && try(grant.schema_name, "") == ""])
    schema_level_grants  = length([for grant in local.rolegrant_result : grant if try(grant.schema_name, "") != "" && try(grant.table_name, "") == ""])
    table_level_grants   = length([for grant in local.rolegrant_result : grant if try(grant.table_name, "") != ""])
    status              = "success"
  } : {
    total_grants         = 0
    unique_privileges    = 0
    unique_catalogs      = 0
    unique_grantees      = 0
    catalog_level_grants = 0
    schema_level_grants  = 0
    table_level_grants   = 0
    status              = "no_roles_available"
  }
  description = "Summary statistics of role grants"
}
