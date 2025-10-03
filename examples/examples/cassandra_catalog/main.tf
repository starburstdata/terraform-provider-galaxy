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

# Variables sourced from environment
variable "TESTING_CASSANDRA_AWS_CONTACT_POINTS" {
  description = "Cassandra contact points"
  type        = string
  default     = null
}

variable "TESTING_CASSANDRA_AWS_LOCAL_DC" {
  description = "Cassandra local datacenter"
  type        = string
  default     = "datacenter1"
}

variable "TESTING_CASSANDRA_AWS_USERNAME" {
  description = "Cassandra username"
  type        = string
  default     = null
}

variable "TESTING_CASSANDRA_AWS_PASSWORD" {
  description = "Cassandra password"
  type        = string
  sensitive   = true
  default     = null
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

# Create a Cassandra catalog
resource "galaxy_cassandra_catalog" "example" {
  name             = "cass${local.test_suffix}"
  deployment_type  = "apacheCassandra"
  contact_points   = var.TESTING_CASSANDRA_AWS_CONTACT_POINTS
  local_datacenter = var.TESTING_CASSANDRA_AWS_LOCAL_DC
  port             = 9042
  username         = var.TESTING_CASSANDRA_AWS_USERNAME
  password         = var.TESTING_CASSANDRA_AWS_PASSWORD
  read_only        = false
  description      = "Example Cassandra catalog using environment variables"
}

# Data source to read the catalog
data "galaxy_cassandra_catalog" "example" {
  depends_on = [galaxy_cassandra_catalog.example]
  id         = galaxy_cassandra_catalog.example.id
}

# List all Cassandra catalogs
data "galaxy_cassandra_catalogs" "all" {
  depends_on = [galaxy_cassandra_catalog.example]
}

output "cassandra_catalog_id" {
  value = galaxy_cassandra_catalog.example.id
}

output "all_cassandra_catalogs" {
  value     = data.galaxy_cassandra_catalogs.all
  sensitive = true
}
