terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials are provided via environment variables
}

# Create an S3 catalog
locals {
  timestamp = formatdate("YYYYMMDDhhmmss", timestamp())
}

# S3 catalog with Glue metastore (requires region)
resource "galaxy_s3_catalog" "example" {
  name           = "s3cat${local.timestamp}"
  description    = "E2E testing S3 catalog with Glue metastore"
  metastore_type = "glue"
  read_only      = false

  # AWS authentication
  access_key = var.TESTING_AWS_ACCESS_KEY
  secret_key = var.TESTING_AWS_SECRET_KEY

  # Required fields for Glue metastore
  region                = "us-east-1"
  default_bucket        = "e2e-testing-us-east-1"
  default_data_location = "glue/"

  # Optional: IAM role for cross-account access
  # role_arn = "arn:aws:iam::123456789012:role/DataLakeRole"

  default_table_format            = "ICEBERG"
  external_table_creation_enabled = true
  external_table_writes_enabled   = true
}

# S3 catalog with Galaxy metastore (region NOT supported - should be omitted)
resource "galaxy_s3_catalog" "galaxy_example" {
  name           = "s3catgalaxy${local.timestamp}"
  description    = "E2E testing S3 catalog with Galaxy metastore"
  metastore_type = "galaxy"
  read_only      = false

  # AWS authentication
  access_key = var.TESTING_AWS_ACCESS_KEY
  secret_key = var.TESTING_AWS_SECRET_KEY

  # Required fields for Galaxy metastore
  default_bucket        = "e2e-testing-us-east-1"
  default_data_location = "galaxy/"
  # Note: region is NOT included for galaxy metastore type

  default_table_format            = "ICEBERG"
  external_table_creation_enabled = false
  external_table_writes_enabled   = false
}

# Validate S3 catalog configurations
data "galaxy_s3_catalog_validation" "example" {
  catalog_id = galaxy_s3_catalog.example.catalog_id
}

data "galaxy_s3_catalog_validation" "galaxy_example" {
  catalog_id = galaxy_s3_catalog.galaxy_example.catalog_id
}

# Read S3 catalogs by ID
data "galaxy_s3_catalog" "example" {
  depends_on = [galaxy_s3_catalog.example]
  catalog_id = galaxy_s3_catalog.example.catalog_id
}

data "galaxy_s3_catalog" "galaxy_example" {
  depends_on = [galaxy_s3_catalog.galaxy_example]
  catalog_id = galaxy_s3_catalog.galaxy_example.catalog_id
}


data "galaxy_s3_catalogs" "all" {}
# List all S3 catalogs
output "all_s3_catalog_names" {
  description = "Names of all S3 catalogs"
  value       = [for catalog in data.galaxy_s3_catalogs.all.result : catalog.name]
}

output "glue_s3_catalog_id" {
  description = "Glue S3 catalog ID"
  value       = galaxy_s3_catalog.example.catalog_id
}

output "galaxy_s3_catalog_id" {
  description = "Galaxy S3 catalog ID"
  value       = galaxy_s3_catalog.galaxy_example.catalog_id
}

output "glue_s3_catalog_validation" {
  description = "Glue S3 catalog validation result"
  value       = data.galaxy_s3_catalog_validation.example.validation_successful
}

output "galaxy_s3_catalog_validation" {
  description = "Galaxy S3 catalog validation result"
  value       = data.galaxy_s3_catalog_validation.galaxy_example.validation_successful
}
