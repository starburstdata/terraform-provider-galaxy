terraform {
  required_providers {
    galaxy = {
      source = "starburstdata/galaxy"
    }
  }
}

provider "galaxy" {}

locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

resource "galaxy_role" "domain_owner" {
  role_name              = "domowner${local.test_suffix}"
  grant_to_creating_role = true
  role_description       = "Data domain owner"
}

resource "galaxy_data_domain" "example" {
  name        = "dom${local.test_suffix}"
  description = "E2E data domain"
  role_id     = galaxy_role.domain_owner.role_id
}

data "galaxy_data_domain" "example" {
  data_domain_id = galaxy_data_domain.example.data_domain_id
}

output "data_domain_id" {
  value = galaxy_data_domain.example.data_domain_id
}

output "data_domain_name" {
  value = data.galaxy_data_domain.example.name
}
