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

# Create a MySQL catalog
locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}
resource "galaxy_mysql_catalog" "test" {
  name            = "mycat${local.test_suffix}"
  host            = var.TESTING_MYSQL_AWS_HOST
  port            = 3306
  username        = var.TESTING_MYSQL_AWS_USERNAME
  password        = var.TESTING_MYSQL_AWS_PASSWORD
  read_only       = false
  connection_type = "direct"
  description     = "E2E testing MySQL catalog"
}

# Data source to read the catalog
data "galaxy_mysql_catalog" "test" {
  depends_on = [galaxy_mysql_catalog.test]
  catalog_id = galaxy_mysql_catalog.test.catalog_id
}

data "galaxy_mysql_catalog_validation" "test" {
  depends_on = [galaxy_mysql_catalog.test]
  catalog_id = galaxy_mysql_catalog.test.catalog_id
}

output "mysql_catalog_id" {
  value = galaxy_mysql_catalog.test.catalog_id
}

output "mysql_catalog_validation" {
  value = data.galaxy_mysql_catalog_validation.test.validation_successful
}
