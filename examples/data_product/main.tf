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

# First create some catalogs to include in data product
locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

variable "TESTING_POSTGRESQL_AWS_HOST" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL host from integration secrets"
  default     = ""
}

variable "TESTING_POSTGRESQL_AWS_DATABASE" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL database from integration secrets"
  default     = ""
}

variable "TESTING_POSTGRESQL_AWS_USERNAME" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL username from integration secrets"
  default     = ""
}

variable "TESTING_POSTGRESQL_AWS_PASSWORD" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL password from integration secrets"
  default     = ""
}

resource "galaxy_postgresql_catalog" "source1" {
  name          = "pgsrc1${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = true
  description   = "E2E testing PostgreSQL source 1 for data product"
}

resource "galaxy_postgresql_catalog" "source2" {
  name          = "pgsrc2${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = true
  description   = "E2E testing PostgreSQL source 2 for data product"
}

# Create a role first
resource "galaxy_role" "data_product_role" {
  role_name              = "dataprod${local.test_suffix}"
  grant_to_creating_role = true
  role_description       = "Role for data product owner"
}

# Note: Using hardcoded user values to avoid data source type conversion issues
# In production, you would use data.galaxy_user to look up existing users
locals {
  data_product_owner_user_id = "u-9639270557"
  data_product_owner_email   = "dataowner@example.com"
}

# Create tags for the data product
resource "galaxy_tag" "customer_data" {
  name        = "custdata${local.test_suffix}"
  description = "E2E testing - contains customer information"
  color       = "#0000FF"
}

resource "galaxy_tag" "sensitive" {
  name        = "sens${local.test_suffix}"
  description = "E2E testing - contains sensitive information"
  color       = "#FF0000"
}

# Create a data product that combines multiple catalogs
resource "galaxy_data_product" "customer_360" {
  name        = "cust360${local.test_suffix}"
  description = "E2E testing - complete 360-degree view of customer data across all systems"
  summary     = "Customer 360 data product for testing"
  catalog_id  = galaxy_postgresql_catalog.source1.id
  schema_name = "public"

  contacts = [
    {
      email   = local.data_product_owner_email
      user_id = local.data_product_owner_user_id
    }
  ]
}

# Data source to read the data product
data "galaxy_data_product" "customer_360" {
  depends_on = [galaxy_data_product.customer_360]
  id         = galaxy_data_product.customer_360.id
}

output "data_product_id" {
  value = galaxy_data_product.customer_360.id
}

output "data_product_name" {
  value = galaxy_data_product.customer_360.name
}

output "data_product_catalogs" {
  value = galaxy_data_product.customer_360.catalog_id
}

