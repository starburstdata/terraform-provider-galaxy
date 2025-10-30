terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials are provided via environment variables
}

# Generate shorter timestamp for unique naming (to avoid length limits)
locals {
  timestamp = formatdate("YYMMDDhhmm", timestamp())
}

# Variables for testing credentials (set via TF_VAR_ environment variables)
variable "TESTING_AWS_ACCESS_KEY" {
  description = "AWS access key for testing"
  type        = string
  sensitive   = true
}

variable "TESTING_AWS_SECRET_KEY" {
  description = "AWS secret key for testing"
  type        = string
  sensitive   = true
}

# Create required roles
resource "galaxy_role" "data_analyst" {
  role_name              = "tfdataanalyst${local.timestamp}"
  grant_to_creating_role = true
  role_description       = "Role for data analysts"
}

resource "galaxy_role" "data_engineer" {
  role_name              = "tfdataengineer${local.timestamp}"
  grant_to_creating_role = true
  role_description       = "Role for data engineers"
}

# Note: Using hardcoded user values to avoid data source type conversion issues
# In production, you would use data.galaxy_user to look up existing users
locals {
  admin_user_id = "u-9639270557"
  admin_email   = "admin@example.com"
}

# Create required S3 catalog
resource "galaxy_s3_catalog" "example" {
  name                  = "tfexampledatagovernance${local.timestamp}"
  metastore_type        = "galaxy"
  default_bucket        = "example-bucket"
  default_data_location = "data/"
  default_table_format  = "ICEBERG"
  # region is not supported for galaxy metastore type, only for glue
  access_key = var.TESTING_AWS_ACCESS_KEY
  secret_key = var.TESTING_AWS_SECRET_KEY
  read_only  = false
}

# Create a data product for organizing datasets
resource "galaxy_data_product" "customer_analytics" {
  name        = "customeranalytics${local.timestamp}"
  summary     = "Customer analytics data product for business insights"
  description = "Comprehensive data product containing customer transaction and profile data for analytics and reporting purposes"
  schema_name = "analytics_schema"
  catalog_id  = galaxy_s3_catalog.example.catalog_id

  contacts = [
    {
      user_id = local.admin_user_id
      email   = local.admin_email
    }
  ]
}

# Create tags for data classification
resource "galaxy_tag" "pii" {
  name        = "pii_data${local.timestamp}"
  description = "Personally Identifiable Information"
  color       = "#FF0000"
}

resource "galaxy_tag" "sensitive" {
  name        = "sensitive_data${local.timestamp}"
  description = "Sensitive business data"
  color       = "#FFA500"
}

resource "galaxy_tag" "public" {
  name        = "public_data${local.timestamp}"
  description = "Public data"
  color       = "#00FF00"
}

# Create a row filter for data security
resource "galaxy_row_filter" "customer_region_filter" {
  name        = "filter${local.timestamp}"
  description = "Filter customers by region"

  # SQL expression for filtering
  expression = "region = current_user_attribute('region')"
}

# Create a column mask for PII protection
resource "galaxy_column_mask" "ssn_mask" {
  name             = "ssnmasked${local.timestamp}"
  description      = "Mask SSN columns"
  column_mask_type = "Varchar"
  expression       = "CONCAT('XXX-XX-', SUBSTRING(ssn, 8, 4))"
}
# Create a policy that uses the row filter
resource "galaxy_policy" "read_all" {
  name        = "read_all${local.timestamp}"
  description = "Read all policy for customer data"

  # SQL predicate for policy logic - must be valid SQL expression
  predicate = "true"

  # Role that this policy applies to
  role_id = galaxy_role.data_engineer.role_id

  scopes = [
    {
      entity_id       = galaxy_s3_catalog.example.catalog_id
      entity_kind     = "Column"
      schema_name     = "*"
      table_name      = "*"
      column_name     = "*"
      row_filter_ids  = []
      column_mask_ids = []
      privileges = {
        privilege  = ["Select"]
        grant_kind = "Allow"
      }
    }
  ]
}

# Data sources for governance objects
data "galaxy_data_product" "customer_analytics" {
  depends_on = [galaxy_data_product.customer_analytics]
  data_product_id = galaxy_data_product.customer_analytics.data_product_id
}

data "galaxy_data_products" "all" {}

data "galaxy_tag" "pii" {
  depends_on = [galaxy_tag.pii]
  tag_id = galaxy_tag.pii.tag_id
}

data "galaxy_tags" "all" {}

data "galaxy_row_filter" "region_filter" {
  depends_on = [galaxy_row_filter.customer_region_filter]
  row_filter_id = galaxy_row_filter.customer_region_filter.row_filter_id
}

data "galaxy_row_filters" "all" {}

data "galaxy_column_mask" "ssn" {
  depends_on = [galaxy_column_mask.ssn_mask]
  column_mask_id = galaxy_column_mask.ssn_mask.column_mask_id
}

data "galaxy_column_masks" "all" {}

data "galaxy_policies" "all" {}

output "data_product_id" {
  value = galaxy_data_product.customer_analytics.data_product_id
}

output "pii_tag_id" {
  value = galaxy_tag.pii.tag_id
}

output "row_filter_id" {
  value = galaxy_row_filter.customer_region_filter.row_filter_id
}

output "column_mask_id" {
  value = galaxy_column_mask.ssn_mask.column_mask_id
}
