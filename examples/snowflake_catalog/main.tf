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

# Create a Snowflake catalog
locals {
  timestamp = formatdate("MMDDhhmmss", timestamp())
}

resource "galaxy_snowflake_catalog" "test" {
  name               = "sfcat${local.timestamp}"
  account_identifier = var.TESTING_SNOWFLAKE_ACCOUNT_ID
  username           = var.TESTING_SNOWFLAKE_USER
  password           = var.TESTING_SNOWFLAKE_PASSWORD
  database_name      = var.TESTING_SNOWFLAKE_DATABASE
  read_only          = false
  description        = "E2E testing Snowflake data warehouse catalog"
  # Removed role and warehouse to match working curl script
}

# Data source to read the catalog
data "galaxy_snowflake_catalog" "test" {
  depends_on = [galaxy_snowflake_catalog.test]
  id         = galaxy_snowflake_catalog.test.id
}

output "snowflake_catalog_id" {
  value = galaxy_snowflake_catalog.test.id
}

output "snowflake_catalog_database" {
  value     = galaxy_snowflake_catalog.test.database_name
  sensitive = true
}
