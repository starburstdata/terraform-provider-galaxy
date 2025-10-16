terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

# Configure the Galaxy provider
provider "galaxy" {
  # Configuration can be provided via environment variables:
  # GALAXY_CLIENT_ID, GALAXY_CLIENT_SECRET, GALAXY_DOMAIN

  # Or explicitly in the configuration (not recommended for production)
  # client_id     = "your-client-id"
  # client_secret = "your-client-secret"
  # domain        = "https://your-account.galaxy.starburst.io"
}

# Create a Starburst Galaxy cluster
resource "galaxy_cluster" "example" {
  name                                    = "example"
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