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

locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

variable "TESTING_POSTGRESQL_AWS_HOST" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_DATABASE" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_USERNAME" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_PASSWORD" {
  type      = string
  sensitive = true
}

resource "galaxy_postgresql_catalog" "example" {
  name          = "uedscat${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = true
}

data "galaxy_users" "all" {}

locals {
  first_user = data.galaxy_users.all.result[0]
}

resource "galaxy_data_product" "example" {
  name        = "uedsdp${local.test_suffix}"
  catalog_id  = galaxy_postgresql_catalog.example.catalog_id
  schema_name = "public"
  summary     = "Data product for usage examples data source"

  contacts = [
    {
      email   = local.first_user.email
      user_id = local.first_user.user_id
    }
  ]
}

# List all usage examples for the data product
data "galaxy_usage_examples" "example" {
  data_product_id = galaxy_data_product.example.data_product_id
}

output "data_product_id" {
  value = galaxy_data_product.example.data_product_id
}

output "usage_examples_count" {
  value = length(data.galaxy_usage_examples.example.result)
}
