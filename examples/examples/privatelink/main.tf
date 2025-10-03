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

# Get all available privatelinks first to find one to test with
data "galaxy_privatelinks" "all" {}

# Local values for handling empty results with null safety
locals {
  privatelinks_result = try(data.galaxy_privatelinks.all.result, [])

  # Find first non-null privatelink
  first_privatelink = try([for p in local.privatelinks_result : p if p != null][0], null)

  has_privatelinks = local.first_privatelink != null
}

# Test the singular privatelink data source (only if we have privatelinks available)
data "galaxy_privatelink" "test" {
  count = local.has_privatelinks ? 1 : 0
  id    = local.first_privatelink.privatelink_id
}

# Diagnostic outputs
output "privatelinks_available" {
  value = local.has_privatelinks
}

output "privatelinks_count" {
  value = length(local.privatelinks_result)
}

output "first_privatelink_debug" {
  value = {
    found           = local.first_privatelink != null
    privatelink_id  = try(local.first_privatelink.privatelink_id, "none")
    name           = try(local.first_privatelink.name, "none")
    cloud_region_id = try(local.first_privatelink.cloud_region_id, "none")
  }
}

# Privatelink data source test outputs (conditional)
output "privatelink_data_source_test" {
  value = local.has_privatelinks ? {
    id              = data.galaxy_privatelink.test[0].id
    name           = data.galaxy_privatelink.test[0].name
    privatelink_id  = data.galaxy_privatelink.test[0].privatelink_id
    cloud_region_id = data.galaxy_privatelink.test[0].cloud_region_id
    status         = "success"
  } : {
    id              = "not_available"
    name           = "No privatelinks available for testing"
    privatelink_id  = "none"
    cloud_region_id = "none"
    status         = "failed"
  }
}

# Show the first few privatelinks (up to 3) for debugging
output "sample_privatelinks" {
  value = [
    for i, privatelink in slice(local.privatelinks_result, 0, min(3, length(local.privatelinks_result))) :
    {
      index           = i
      name           = try(privatelink.name, "unknown")
      privatelink_id  = try(privatelink.privatelink_id, "unknown")
      cloud_region_id = try(privatelink.cloud_region_id, "unknown")
    }
  ]
}