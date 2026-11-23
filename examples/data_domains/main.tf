terraform {
  required_providers {
    galaxy = {
      source = "starburstdata/galaxy"
    }
  }
}

provider "galaxy" {}

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
  default   = ""
}

variable "TESTING_POSTGRESQL_AWS_DATABASE" {
  type    = string
  default = ""
}

variable "TESTING_POSTGRESQL_AWS_USERNAME" {
  type      = string
  sensitive = true
  default   = ""
}

variable "TESTING_POSTGRESQL_AWS_PASSWORD" {
  type      = string
  sensitive = true
  default   = ""
}

resource "galaxy_postgresql_catalog" "example" {
  name          = "ddc${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = true
  description   = "Catalog for data_domains example"
}

data "galaxy_users" "all" {}

resource "galaxy_data_product" "example" {
  name        = "ddp${local.test_suffix}"
  description = "Data product for data_domains example"
  summary     = "Summary"
  catalog_id  = galaxy_postgresql_catalog.example.catalog_id
  schema_name = "public"
  contacts = [
    {
      email   = data.galaxy_users.all.result[0].email
      user_id = data.galaxy_users.all.result[0].user_id
    }
  ]
}

data "galaxy_data_domains" "example" {
  data_product_id = galaxy_data_product.example.data_product_id
}

output "assigned_domain" {
  value = data.galaxy_data_domains.example.data_domain_id
}
