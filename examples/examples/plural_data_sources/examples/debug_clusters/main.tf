terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials from environment variables:
  # GALAXY_CLIENT_ID, GALAXY_CLIENT_SECRET, GALAXY_DOMAIN
}

# Test to see if there are any clusters at all
# Create a simple cluster first
resource "galaxy_cluster" "test" {
  name                 = "debugcluster12345"
  cloud_region_id      = "aws-us-east1"
  min_workers          = 1
  max_workers          = 2
  idle_stop_minutes    = 5
  private_link_cluster = false
  catalog_refs         = []
}

# Then test the data source
data "galaxy_clusters" "all" {
  depends_on = [galaxy_cluster.test]
}

output "clusters_result" {
  value       = data.galaxy_clusters.all.result
  description = "Raw result from clusters data source"
}

output "clusters_count" {
  value       = length(data.galaxy_clusters.all.result)
  description = "Count of clusters returned"
}

output "created_cluster_id" {
  value       = galaxy_cluster.test.id
  description = "ID of the cluster we just created"
}
