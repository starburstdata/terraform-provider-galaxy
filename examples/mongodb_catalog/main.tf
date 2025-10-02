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

# Use timestamps for unique naming
locals {
  timestamp = formatdate("MMDDhhmmss", timestamp())
}

# Create a MongoDB catalog
resource "galaxy_mongodb_catalog" "example" {
  name            = "mongodbcatalog${local.timestamp}"
  host            = "mongodb.us-east-1.amazonaws.com"
  database        = "mydatabase"
  username        = "mongodbuser"
  password        = "mongodbpassword"
  read_only       = false
  connection_type = "direct"
  regions         = ["aws-us-east1"]
  description     = "Example MongoDB catalog"
}

# Data source to read the catalog
data "galaxy_mongodb_catalog" "example" {
  depends_on = [galaxy_mongodb_catalog.example]
  id         = galaxy_mongodb_catalog.example.id
}

# List all MongoDB catalogs
data "galaxy_mongodb_catalogs" "all" {
  depends_on = [galaxy_mongodb_catalog.example]
}

output "mongodb_catalog_id" {
  value = galaxy_mongodb_catalog.example.id
}

output "all_mongodb_catalogs" {
  value       = [for catalog in data.galaxy_mongodb_catalogs.all.result : catalog.name]
  description = "Names of all MongoDB catalogs"
}
