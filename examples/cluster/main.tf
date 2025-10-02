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

# Use TEST_SUFFIX environment variable for unique naming
locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

# Create a cluster with basic configuration
resource "galaxy_cluster" "test" {
  name                                    = "cluster${local.test_suffix}"
  cloud_region_id                         = "aws-us-east1"
  min_workers                             = 1
  max_workers                             = 2
  idle_stop_minutes                       = 15
  private_link_cluster                    = false
  result_cache_enabled                    = true
  result_cache_default_visibility_seconds = 3600
  warp_resiliency_enabled                 = false
  # processing_mode can be "Batch" or "WarpSpeed" (optional)

  # Associate with catalogs (need to create catalogs first)
  catalog_refs = []
}


output "cluster_id" {
  value = galaxy_cluster.test.id
}

output "cluster_state" {
  value = galaxy_cluster.test.cluster_state
}

output "trino_uri" {
  value = galaxy_cluster.test.trino_uri
}

# Data source example for cluster
data "galaxy_cluster" "existing_cluster" {
  id = galaxy_cluster.test.id
}

output "cluster_data" {
  value = data.galaxy_cluster.existing_cluster
}
