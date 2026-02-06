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

# Use timestamps for unique naming
locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

# Create the base role whose privileges will be inherited
resource "galaxy_role" "base" {
  role_name              = "baserole${local.test_suffix}"
  role_description       = "Base role with shared privileges"
  grant_to_creating_role = true
}

# Create a child role that will inherit from the base role
resource "galaxy_role" "child" {
  role_name              = "childrole${local.test_suffix}"
  role_description       = "Child role that inherits from the base role"
  grant_to_creating_role = true
}

# Grant the base role to the child role so the child inherits its privileges
resource "galaxy_role_grant" "example" {
  role_id         = galaxy_role.child.role_id
  granted_role_id = galaxy_role.base.role_id
  admin_option    = false
}

# Verify via the rolegrant data source
data "galaxy_rolegrant" "verify" {
  role_id = galaxy_role.child.role_id

  depends_on = [galaxy_role_grant.example]
}

output "base_role_id" {
  value = galaxy_role.base.role_id
}

output "child_role_id" {
  value = galaxy_role.child.role_id
}

output "granted_role_name" {
  value = galaxy_role_grant.example.granted_role_name
}

output "role_grants" {
  value = data.galaxy_rolegrant.verify.result
}
