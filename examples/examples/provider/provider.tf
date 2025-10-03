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