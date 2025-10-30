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

# Create a catalog to apply masks to
locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

# Variables for PostgreSQL connection
variable "TESTING_POSTGRESQL_AWS_HOST" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL host from integration secrets"
  default     = ""
}

variable "TESTING_POSTGRESQL_AWS_DATABASE" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL database from integration secrets"
  default     = ""
}

variable "TESTING_POSTGRESQL_AWS_USERNAME" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL username from integration secrets"
  default     = ""
}

variable "TESTING_POSTGRESQL_AWS_PASSWORD" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL password from integration secrets"
  default     = ""
}

resource "galaxy_postgresql_catalog" "test" {
  name          = "maskdb${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = true
}

# Create roles to apply masks to
resource "galaxy_role" "analyst" {
  role_name              = "analyst${local.test_suffix}"
  grant_to_creating_role = true
  role_description       = "Analyst with masked PII access"
}

resource "galaxy_role" "support" {
  role_name              = "support${local.test_suffix}"
  grant_to_creating_role = true
  role_description       = "Support staff with limited PII visibility"
}

# Mask SSN - show only last 4 digits
resource "galaxy_column_mask" "ssn_mask" {
  name             = "ssnmasker${local.test_suffix}"
  description      = "Mask SSN showing only last 4 digits"
  column_mask_type = "Varchar"
  expression       = "CONCAT('XXX-XX-', SUBSTRING(ssn, 8, 4))"
}

# Mask email - show only domain
resource "galaxy_column_mask" "email_mask" {
  name             = "emailmask${local.test_suffix}"
  description      = "Show only email domain"
  column_mask_type = "Varchar"
  expression       = "CONCAT('***@', SPLIT_PART(email, '@', 2))"
}

# Mask credit card - show only last 4 digits
resource "galaxy_column_mask" "credit_card_mask" {
  name             = "cardmask${local.test_suffix}"
  description      = "Mask credit card showing only last 4 digits"
  column_mask_type = "Varchar"
  expression       = "CONCAT('****-****-****-', SUBSTRING(card_number, 13, 4))"
}

# Hash sensitive data
resource "galaxy_column_mask" "phone_hash" {
  name             = "phonemask${local.test_suffix}"
  description      = "Hash phone numbers for privacy"
  column_mask_type = "Varchar"
  expression       = "MD5(phone_number)"
}

# Data source to read column mask
data "galaxy_column_mask" "ssn" {
  depends_on = [galaxy_column_mask.ssn_mask]
  column_mask_id = galaxy_column_mask.ssn_mask.column_mask_id
}

output "ssn_mask_id" {
  value = galaxy_column_mask.ssn_mask.column_mask_id
}
output "ssn_mask_data" {
  value = data.galaxy_column_mask.ssn
}

output "ssn_mask_expression" {
  value = galaxy_column_mask.ssn_mask.expression
}

