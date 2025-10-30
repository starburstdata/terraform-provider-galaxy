terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials from environment variables
}

locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

resource "galaxy_bigquery_catalog" "test" {
  name            = "bqcat${local.test_suffix}"
  credentials_key = var.TESTING_GCS_JSON_KEY
  read_only       = false
  description     = "BigQuery data warehouse catalog"
}

# Data source to read the catalog
data "galaxy_bigquery_catalog" "test" {
  depends_on = [galaxy_bigquery_catalog.test]
  catalog_id = galaxy_bigquery_catalog.test.catalog_id
}

output "bigquery_catalog_id" {
  value = galaxy_bigquery_catalog.test.catalog_id
}

output "bigquery_catalog_name" {
  value = galaxy_bigquery_catalog.test.name
}
