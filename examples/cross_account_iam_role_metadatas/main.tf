terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
}

# Get all available cross account IAM role metadatas
data "galaxy_cross_account_iam_role_metadatas" "all" {}

# Local values - this data source returns fields directly, not in a result array
locals {
  # Check if the data source has valid data
  has_external_id = data.galaxy_cross_account_iam_role_metadatas.all.external_id != null && data.galaxy_cross_account_iam_role_metadatas.all.external_id != ""
  has_account_id  = data.galaxy_cross_account_iam_role_metadatas.all.starburst_aws_account_id != null && data.galaxy_cross_account_iam_role_metadatas.all.starburst_aws_account_id != ""
  has_metadatas   = local.has_external_id || local.has_account_id
}

# Diagnostic outputs
output "metadatas_count" {
  value       = local.has_metadatas ? 1 : 0
  description = "Whether cross account IAM role metadata is available (1) or not (0)"
}

output "has_metadatas" {
  value       = local.has_metadatas
  description = "Whether any cross account IAM role metadatas exist"
}

# Show details of the metadata (or empty if none exist)
output "metadatas_details" {
  value = local.has_metadatas ? {
    external_id            = try(data.galaxy_cross_account_iam_role_metadatas.all.external_id, "unknown")
    starburst_aws_account_id = try(data.galaxy_cross_account_iam_role_metadatas.all.starburst_aws_account_id, "unknown")
  } : {}
  description = "Details of cross account IAM role metadata"
}

# Show metadata details for reference (same as metadatas_details since there's only one record)
output "first_metadata_info" {
  value = local.has_metadatas ? {
    external_id            = try(data.galaxy_cross_account_iam_role_metadatas.all.external_id, "unknown")
    starburst_aws_account_id = try(data.galaxy_cross_account_iam_role_metadatas.all.starburst_aws_account_id, "unknown")
  } : {
    external_id            = "No metadata available"
    starburst_aws_account_id = "none"
  }
  description = "Information about the cross account IAM role metadata (if any exists)"
}

# Summary by AWS account
output "metadatas_by_account" {
  value = local.has_metadatas ? {
    unique_accounts = 1
    accounts_list   = [try(data.galaxy_cross_account_iam_role_metadatas.all.starburst_aws_account_id, "unknown")]
  } : {
    unique_accounts = 0
    accounts_list   = []
  }
  description = "Summary of metadata grouped by AWS account"
}

# Summary output
output "cross_account_iam_role_metadatas_summary" {
  value = {
    total_count     = local.has_metadatas ? 1 : 0
    has_metadatas   = local.has_metadatas
    unique_accounts = local.has_metadatas ? 1 : 0
    external_id     = try(data.galaxy_cross_account_iam_role_metadatas.all.external_id, "none")
    status         = local.has_metadatas ? "success" : "no_metadatas_available"
  }
  description = "Summary of cross account IAM role metadatas data source test"
}

# Show metadata info (this data source doesn't contain role ARNs, just external_id and account_id)
output "metadata_info" {
  value = local.has_metadatas ? {
    external_id_length      = length(try(data.galaxy_cross_account_iam_role_metadatas.all.external_id, ""))
    starburst_aws_account_id = try(data.galaxy_cross_account_iam_role_metadatas.all.starburst_aws_account_id, "unknown")
  } : {}
  description = "Cross account IAM role metadata information"
}
