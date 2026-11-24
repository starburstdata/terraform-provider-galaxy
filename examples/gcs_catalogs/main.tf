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

# List all GCS catalogs in the Galaxy account.
data "galaxy_gcs_catalogs" "all" {}

# ENG-32720: reading element attributes (not only length) is the regression
# guard — if the underlying mapper skips a required attribute, every element
# collapses to null and this expression fails with
# "Attempt to get attribute from null value".
output "gcs_catalog_names" {
  value       = [for c in data.galaxy_gcs_catalogs.all.result : c.name]
  description = "Names of all GCS catalogs"
}

output "gcs_catalogs_count" {
  value       = length(data.galaxy_gcs_catalogs.all.result)
  description = "Total number of GCS catalogs"
}
