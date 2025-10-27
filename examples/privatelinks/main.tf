terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
}

# Get all available privatelinks
data "galaxy_privatelinks" "all" {}

# Local values for handling empty results with null safety
locals {
  privatelinks_result = try(data.galaxy_privatelinks.all.result, [])
  has_privatelinks    = length(local.privatelinks_result) > 0
}

# Diagnostic outputs
output "privatelinks_count" {
  value       = length(local.privatelinks_result)
  description = "Total number of privatelinks"
}

output "has_privatelinks" {
  value       = local.has_privatelinks
  description = "Whether any privatelinks exist"
}

# Show details of all privatelinks (or message if none exist)
output "privatelinks_details" {
  value = local.has_privatelinks ? [
    for privatelink in local.privatelinks_result : {
      name           = try(privatelink.name, "unknown")
      privatelink_id  = try(privatelink.privatelink_id, "unknown")
      cloud_region_id = try(privatelink.cloud_region_id, "unknown")
    }
  ] : []
  description = "Details of all available privatelinks"
}

# Show first privatelink details for reference
output "first_privatelink_info" {
  value = local.has_privatelinks ? {
    name           = try(local.privatelinks_result[0].name, "unknown")
    privatelink_id  = try(local.privatelinks_result[0].privatelink_id, "unknown")
    cloud_region_id = try(local.privatelinks_result[0].cloud_region_id, "unknown")
  } : {
    name           = "No privatelinks available"
    privatelink_id  = "none"
    cloud_region_id = "none"
  }
  description = "Information about the first privatelink (if any exists)"
}

# Summary output
output "privatelinks_summary" {
  value = {
    total_count     = length(local.privatelinks_result)
    has_privatelinks = local.has_privatelinks
    status         = local.has_privatelinks ? "success" : "no_privatelinks_available"
  }
  description = "Summary of privatelinks data source test"
}
