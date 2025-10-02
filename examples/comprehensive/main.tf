terraform {
  required_providers {
    galaxy = {
      source = "hashicorp.com/starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials from environment variables:
  # GALAXY_CLIENT_ID, GALAXY_CLIENT_SECRET, GALAXY_DOMAIN
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

# ==========================================
# CATALOGS
# ==========================================

# MySQL Catalog
resource "galaxy_mysql_catalog" "transactional" {
  name            = "mysqltransactional${local.test_suffix}"
  host            = var.TESTING_MYSQL_AWS_HOST
  port            = 3306
  username        = var.TESTING_MYSQL_AWS_USERNAME
  password        = var.TESTING_MYSQL_AWS_PASSWORD
  read_only       = false
  connection_type = "direct"
  description     = "Main transactional database"
}

# PostgreSQL Catalog
resource "galaxy_postgresql_catalog" "analytics" {
  name          = "postgresanalytics${local.test_suffix}"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = true
  description   = "Analytics read replica"
}

# Redshift Catalog
resource "galaxy_redshift_catalog" "warehouse" {
  name      = "redshiftwarehouse${local.test_suffix}"
  endpoint  = var.TESTING_REDSHIFT_ENDPOINT
  username  = var.TESTING_REDSHIFT_USER
  password  = var.TESTING_REDSHIFT_PASSWORD
  auth_type = "basic"
  read_only = false
}

# S3 Catalog removed due to cross-region querying issues
# Cross-region querying is not supported between cluster and S3 catalog

# ==========================================
# CLUSTER
# ==========================================

resource "galaxy_cluster" "main" {
  name                 = "prod${local.test_suffix}"
  cloud_region_id      = "aws-us-east1"
  min_workers          = 2
  max_workers          = 10
  idle_stop_minutes    = 30
  private_link_cluster = false

  # Performance settings
  result_cache_enabled                    = true
  result_cache_default_visibility_seconds = 3600
  warp_resiliency_enabled                 = false

  # Attach catalogs to cluster (ordered to match API response)
  catalog_refs = [
    galaxy_redshift_catalog.warehouse.id,   # c-6145040459
    galaxy_mysql_catalog.transactional.id,  # c-3661228661
    galaxy_postgresql_catalog.analytics.id, # c-6981255000
  ]
}

# ==========================================
# USERS AND ROLES
# ==========================================

# Create a role for data analysts
resource "galaxy_role" "data_analyst" {
  role_name              = "dataanalyst${local.test_suffix}"
  grant_to_creating_role = true
  role_description       = "Role for data analysts with read access"
}

# Create a role for data engineers
resource "galaxy_role" "data_engineer" {
  role_name              = "dataengineer${local.test_suffix}"
  grant_to_creating_role = true
  role_description       = "Role for data engineers with write access"
}

# Create an admin role
resource "galaxy_role" "admin" {
  role_name              = "admin${local.test_suffix}"
  grant_to_creating_role = true
  role_description       = "Administrative role with full access"
}

# Note: Using hardcoded user values to avoid data source type conversion issues
# In production, you would use data.galaxy_user to look up existing users
locals {
  admin_user_id = "u-9639270557"
  admin_email   = "admin@example.com"
}

# ==========================================
# SERVICE ACCOUNTS
# ==========================================

resource "galaxy_service_account" "etl" {
  username              = "etlservice${local.test_suffix}"
  with_initial_password = true

  additional_role_ids = [
    galaxy_role.data_engineer.id
  ]
}

resource "galaxy_service_account" "reporting" {
  username              = "reportingservice${local.test_suffix}"
  with_initial_password = true

  additional_role_ids = [
    galaxy_role.data_analyst.id
  ]
}

# ==========================================
# DATA GOVERNANCE
# ==========================================

# Create tags for data classification
resource "galaxy_tag" "pii" {
  name        = "pii${local.test_suffix}"
  description = "Personally Identifiable Information"
  color       = "#FF0000"
}

resource "galaxy_tag" "financial" {
  name        = "financial${local.test_suffix}"
  description = "Financial and payment data"
  color       = "#FFA500"
}

resource "galaxy_tag" "public" {
  name        = "public${local.test_suffix}"
  description = "Public data"
  color       = "#00FF00"
}

# Row filter for data security
resource "galaxy_row_filter" "region_filter" {
  name        = "regionfilter${local.test_suffix}"
  description = "Filter data by region"
  expression  = "region = current_user_region()"
}

# Column mask for PII protection
resource "galaxy_column_mask" "ssn_mask" {
  name             = "ssnmasking${local.test_suffix}"
  description      = "Mask SSN for non-admin users"
  column_mask_type = "Varchar"
  expression       = "CASE WHEN current_role() = 'admin' THEN ssn ELSE 'XXX-XX-' || SUBSTR(ssn, -4) END"
}

# ==========================================
# DATA PRODUCTS
# ==========================================

resource "galaxy_data_product" "customer_360" {
  name        = "customer360${local.test_suffix}"
  description = "Complete customer view across all systems"
  summary     = "360-degree view of customer data"
  schema_name = "customeranalytics"
  catalog_id  = galaxy_mysql_catalog.transactional.id

  contacts = []
  # contacts = [
  #   {
  #     email   = local.admin_email
  #     user_id = local.admin_user_id
  #   }
  # ]
}

# ==========================================
# POLICIES
# ==========================================

resource "galaxy_policy" "data_retention" {
  name        = "retention${local.test_suffix}"
  description = "30-day retention for temporary tables"
  role_id     = galaxy_role.admin.id
  predicate   = "true"

  scopes = [
    {
      entity_id       = galaxy_mysql_catalog.transactional.id
      entity_kind     = "Catalog"
      row_filter_ids  = [galaxy_row_filter.region_filter.id]
      column_mask_ids = []
    }
  ]
}

# ==========================================
# DATA SOURCES
# ==========================================


# Role data source 
data "galaxy_role" "existing_admin" {
  id = galaxy_role.admin.id
}

# Catalogs data source
data "galaxy_catalogs" "all_catalogs" {
  depends_on = [
    galaxy_mysql_catalog.transactional,
    galaxy_postgresql_catalog.analytics,
    galaxy_redshift_catalog.warehouse
  ]
}

# Cluster data source
data "galaxy_cluster" "main_cluster" {
  id = galaxy_cluster.main.id
}

# Policies data source
data "galaxy_policies" "all_policies" {
  # This will list all policies in the domain
}

# Cross account IAM roles data source
data "galaxy_cross_account_iam_roles" "iam_roles" {
  # This will list all cross-account IAM roles
}

# Catalog metadata data source
data "galaxy_catalog_metadata" "transactional_metadata" {
  id = "name=${galaxy_mysql_catalog.transactional.name}"

  depends_on = [
    galaxy_mysql_catalog.transactional
  ]
}

# User data source example using a valid email address
data "galaxy_user" "current_user" {
  # Use email lookup for the current user
  id = "email=brady.burke@starburstdata.com"
}

# ==========================================
# OUTPUTS
# ==========================================

output "cluster_id" {
  value       = galaxy_cluster.main.id
  description = "The ID of the main cluster"
}

output "cluster_state" {
  value       = galaxy_cluster.main.cluster_state
  description = "Current state of the cluster"
}

output "trino_uri" {
  value       = galaxy_cluster.main.trino_uri
  description = "Trino connection URI"
  sensitive   = true
}

output "service_account_etl" {
  value       = galaxy_service_account.etl.username
  description = "ETL service account username"
}

output "admin_role_id" {
  value       = galaxy_role.admin.id
  description = "Admin role ID"
}

# Data source outputs
output "existing_admin_role" {
  value       = data.galaxy_role.existing_admin
  description = "Admin role details from data source"
}

output "current_user_info" {
  value       = data.galaxy_user.current_user
  description = "Current user information from data source"
  sensitive   = true
}

output "all_catalogs_count" {
  value       = data.galaxy_catalogs.all_catalogs
  description = "List of all catalogs"
}

output "cluster_details" {
  value       = data.galaxy_cluster.main_cluster
  description = "Main cluster details from data source"
  sensitive   = true
}

output "policies_count" {
  value       = data.galaxy_policies.all_policies
  description = "List of all policies"
}

output "iam_roles_count" {
  value       = data.galaxy_cross_account_iam_roles.iam_roles
  description = "List of cross-account IAM roles"
}

output "catalog_metadata" {
  value       = data.galaxy_catalog_metadata.transactional_metadata
  description = "Catalog metadata information"
}

# ==========================================
# VARIABLES
# ==========================================

variable "TESTING_MYSQL_AWS_HOST" {
  type        = string
  sensitive   = true
  description = "Testing MySQL AWS host from integration secrets"
}

variable "TESTING_MYSQL_AWS_USERNAME" {
  type        = string
  sensitive   = true
  description = "Testing MySQL AWS username from integration secrets"
}

variable "TESTING_MYSQL_AWS_PASSWORD" {
  type        = string
  sensitive   = true
  description = "Testing MySQL AWS password from integration secrets"
}

variable "TESTING_POSTGRESQL_AWS_HOST" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL AWS host from integration secrets"
}

variable "TESTING_POSTGRESQL_AWS_DATABASE" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL AWS database from integration secrets"
}

variable "TESTING_POSTGRESQL_AWS_USERNAME" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL AWS username from integration secrets"
}

variable "TESTING_POSTGRESQL_AWS_PASSWORD" {
  type        = string
  sensitive   = true
  description = "Testing PostgreSQL AWS password from integration secrets"
}

variable "TESTING_REDSHIFT_PASSWORD" {
  type        = string
  sensitive   = true
  description = "Redshift database password"
}

# S3 variables removed - S3 catalog removed due to cross-region issues

variable "TESTING_REDSHIFT_ENDPOINT" {
  type        = string
  sensitive   = true
  description = "Testing Redshift endpoint from integration secrets"
}


variable "TESTING_REDSHIFT_USER" {
  type        = string
  sensitive   = true
  description = "Testing Redshift user from integration secrets"
}

