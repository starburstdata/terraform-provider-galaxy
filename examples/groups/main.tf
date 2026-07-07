terraform {
  required_providers {
    galaxy = {
      source = "starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials from environment variables
}

# List all groups
data "galaxy_groups" "all" {}

output "groups_count" {
  value = length(try(data.galaxy_groups.all.result, []))
}

output "groups_result" {
  value = try(data.galaxy_groups.all.result, [])
}
