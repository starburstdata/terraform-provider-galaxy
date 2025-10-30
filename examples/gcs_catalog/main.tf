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

# Create a GCS catalog with Galaxy metastore
locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

resource "galaxy_gcs_catalog" "test" {
  name                  = "gcscat${local.test_suffix}"
  metastore_type        = "galaxy"
  credentials_key       = var.TESTING_GCS_JSON_KEY != "" ? var.TESTING_GCS_JSON_KEY : "{\"type\": \"service_account\"}"
  default_bucket        = var.TESTING_GCS_BUCKET != "" ? var.TESTING_GCS_BUCKET : "test-bucket"
  default_data_location = "testdata"
  read_only             = false
  default_table_format  = "ICEBERG"
}

# Data source to read the catalog
data "galaxy_gcs_catalog" "test" {
  depends_on = [galaxy_gcs_catalog.test]
  catalog_id = galaxy_gcs_catalog.test.catalog_id
}

output "gcs_catalog_id" {
  value = galaxy_gcs_catalog.test.catalog_id
}
