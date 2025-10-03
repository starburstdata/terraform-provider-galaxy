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

# Create a catalog to apply filters to
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

resource "galaxy_postgresql_catalog" "test" {
  name          = "filterdb${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = true
}

# Create roles to apply filters to
resource "galaxy_role" "regional_analyst" {
  role_name              = "analyst${local.test_suffix}"
  grant_to_creating_role = true
  role_description       = "Analyst with access to specific regions only"
}

resource "galaxy_role" "limited_viewer" {
  role_name              = "viewer${local.test_suffix}"
  grant_to_creating_role = true
  role_description       = "Viewer with limited data access"
}

# Create row filter for regional data access
resource "galaxy_row_filter" "region_filter" {
  name        = "regionfilter${local.test_suffix}"
  description = "Restrict access to US East region data only"
  expression  = "region = 'US-EAST'"
}

# Create row filter for time-based access
resource "galaxy_row_filter" "time_filter" {
  name        = "timefilter${local.test_suffix}"
  description = "Only show data from last 30 days"
  expression  = "event_date >= CURRENT_DATE - INTERVAL '30' DAY"
}

# Create row filter for customer visibility
resource "galaxy_row_filter" "customer_filter" {
  name        = "customerfilter${local.test_suffix}"
  description = "Users can only see their own customer data"
  expression  = "account_manager = CURRENT_USER"
}

# Data source to read row filter
data "galaxy_row_filter" "region" {
  depends_on = [galaxy_row_filter.region_filter]
  id         = galaxy_row_filter.region_filter.id
}

output "region_filter_data" {
  value = data.galaxy_row_filter.region
}

output "region_filter_id" {
  value = galaxy_row_filter.region_filter.id
}

output "region_filter_expression" {
  value = galaxy_row_filter.region_filter.expression
}

