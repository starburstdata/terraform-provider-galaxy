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

# Create a cross-account IAM role for AWS access
locals {
  timestamp = formatdate("YYYYMMDDhhmmss", timestamp())
}

resource "galaxy_cross_account_iam_role" "example" {
  alias_name  = "s3access${local.timestamp}"
  aws_iam_arn = var.TESTING_CROSS_ACCOUNT_ACCESS_ROLE
}

# List all cross-account IAM roles
data "galaxy_cross_account_iam_roles" "all" {
  depends_on = [galaxy_cross_account_iam_role.example]
}

output "cross_account_iam_role_alias" {
  value = galaxy_cross_account_iam_role.example.alias_name
}

output "cross_account_iam_role_arn" {
  value     = galaxy_cross_account_iam_role.example.aws_iam_arn
  sensitive = true
}

variable "TESTING_CROSS_ACCOUNT_ACCESS_ROLE" {
  type        = string
  sensitive   = true
  description = "AWS IAM role ARN for testing from integration secrets"
}
