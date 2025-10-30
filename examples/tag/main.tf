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

# Use TEST_SUFFIX environment variable for unique naming
locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

# Create tags for data classification
resource "galaxy_tag" "pii" {
  name        = "pii${local.test_suffix}"
  description = "Personally Identifiable Information - requires special handling"
  color       = "#FF0000" # Red for high sensitivity
}

resource "galaxy_tag" "public" {
  name        = "public${local.test_suffix}"
  description = "Public data that can be shared freely"
  color       = "#00FF00" # Green for public data
}

resource "galaxy_tag" "confidential" {
  name        = "confidential${local.test_suffix}"
  description = "Confidential business data"
  color       = "#FFA500" # Orange for confidential
}

resource "galaxy_tag" "financial" {
  name        = "financial${local.test_suffix}"
  description = "Financial and payment data"
  color       = "#FFFF00" # Yellow for financial
}

resource "galaxy_tag" "healthcare" {
  name        = "healthcare${local.test_suffix}"
  description = "Healthcare data subject to HIPAA"
  color       = "#800080" # Purple for healthcare
}

# Data source to read a tag
data "galaxy_tag" "pii" {
  tag_id = galaxy_tag.pii.tag_id
}

# List all tags -
data "galaxy_tags" "all" {
  depends_on = [galaxy_tag.pii, galaxy_tag.public, galaxy_tag.confidential, galaxy_tag.financial, galaxy_tag.healthcare]
}

output "pii_tag_id" {
  value = galaxy_tag.pii.tag_id
}

output "all_tag_names" {
  value = [
    for tag in data.galaxy_tags.all.result : tag.name
  ]
}
