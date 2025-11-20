terraform {
  required_providers {
    galaxy = {
      source = "starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials from environment variables
}

# Use TEST_SUFFIX environment variable for unique naming
locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

# Create an OpenSearch catalog with basic auth
resource "galaxy_opensearch_catalog" "example" {
  name        = "opensearchcatalog${local.test_suffix}"
  endpoint    = "https://opensearch.us-east-1.amazonaws.com"
  port        = 443
  auth_type   = "basic"
  username    = "opensearchuser${local.test_suffix}"
  password    = "opensearchpassword"
  read_only   = false
  description = "Example OpenSearch catalog"
}

# Data source to read the catalog
data "galaxy_opensearch_catalog" "example" {
  depends_on = [galaxy_opensearch_catalog.example]
  catalog_id = galaxy_opensearch_catalog.example.catalog_id
}

# List all OpenSearch catalogs
data "galaxy_opensearch_catalogs" "all" {
  depends_on = [galaxy_opensearch_catalog.example]
}

output "opensearch_catalog_id" {
  value = galaxy_opensearch_catalog.example.catalog_id
}

output "all_opensearch_catalogs" {
  value     = data.galaxy_opensearch_catalogs.all
  sensitive = true
}
