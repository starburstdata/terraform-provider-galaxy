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

locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

# Create a role and list all grants for it
resource "galaxy_role" "example" {
  role_name              = "rgsrc${local.test_suffix}"
  grant_to_creating_role = true
  role_description       = "Example role for rolegrants data source"
}

# List all role grants (role memberships) for this role
data "galaxy_rolegrants" "example" {
  role_id = galaxy_role.example.role_id
}

output "role_id" {
  value = galaxy_role.example.role_id
}

output "rolegrants_count" {
  value = length(data.galaxy_rolegrants.example.result)
}
