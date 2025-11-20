terraform {
  required_providers {
    galaxy = {
      source = "starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials are provided via environment variables
}

# Create a SQL Server catalog
locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

variable "TESTING_SQLSERVER_AWS_HOST" {
  type        = string
  sensitive   = true
  description = "Testing SQL Server host from integration secrets"
  default     = ""
}

variable "TESTING_SQLSERVER_AWS_DATABASE" {
  type        = string
  sensitive   = true
  description = "Testing SQL Server database from integration secrets"
  default     = ""
}

variable "TESTING_SQLSERVER_AWS_USERNAME" {
  type        = string
  sensitive   = true
  description = "Testing SQL Server username from integration secrets"
  default     = ""
}

variable "TESTING_SQLSERVER_AWS_PASSWORD" {
  type        = string
  sensitive   = true
  description = "Testing SQL Server password from integration secrets"
  default     = ""
}

resource "galaxy_sqlserver_catalog" "test" {
  name          = "sqlcat${local.test_suffix}"
  endpoint      = var.TESTING_SQLSERVER_AWS_HOST
  port          = 1433
  database_name = var.TESTING_SQLSERVER_AWS_DATABASE
  username      = var.TESTING_SQLSERVER_AWS_USERNAME
  password      = var.TESTING_SQLSERVER_AWS_PASSWORD
  read_only     = false
  cloud_kind    = "AWS"
}

# Data source to read the catalog
data "galaxy_sqlserver_catalog" "test" {
  depends_on = [galaxy_sqlserver_catalog.test]
  catalog_id = galaxy_sqlserver_catalog.test.catalog_id
}

output "sqlserver_catalog_id" {
  value = galaxy_sqlserver_catalog.test.catalog_id
}
