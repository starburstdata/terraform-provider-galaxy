terraform {
  required_providers {
    galaxy = {
      source = "starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials from environment variables:
  # GALAXY_CLIENT_ID, GALAXY_CLIENT_SECRET, GALAXY_DOMAIN
}

# ==========================================
# PLURAL DATA SOURCES EXAMPLES
# ==========================================

# Get all clusters
data "galaxy_clusters" "all" {}

# Get all users
data "galaxy_users" "all" {}

# Get all roles
data "galaxy_roles" "all" {}

# Get all service accounts
data "galaxy_service_accounts" "all" {}

# Get all tags
data "galaxy_tags" "all" {}

# Get all policies
data "galaxy_policies" "all" {}

# Get all data products
data "galaxy_data_products" "all" {}

# Get all catalogs
data "galaxy_catalogs" "all" {}

# Get all S3 catalogs
data "galaxy_s3_catalogs" "all" {}

# Get all BigQuery catalogs
data "galaxy_bigquery_catalogs" "all" {}

# Get all Redshift catalogs
data "galaxy_redshift_catalogs" "all" {}

# Get all PostgreSQL catalogs
data "galaxy_postgresql_catalogs" "all" {}

# Get all MySQL catalogs
data "galaxy_mysql_catalogs" "all" {}

# Get all GCS catalogs
data "galaxy_gcs_catalogs" "all" {}

# Get all Snowflake catalogs
data "galaxy_snowflake_catalogs" "all" {}

# Get all SQL Server catalogs
data "galaxy_sqlserver_catalogs" "all" {}

# Get all MongoDB catalogs
data "galaxy_mongodb_catalogs" "all" {}

# Get all Cassandra catalogs
data "galaxy_cassandra_catalogs" "all" {}

# Get all OpenSearch catalogs
data "galaxy_opensearch_catalogs" "all" {}

# Get all column masks
data "galaxy_column_masks" "all" {}

# Get all row filters
data "galaxy_row_filters" "all" {}

# Get all cross account IAM roles
data "galaxy_cross_account_iam_roles" "all" {}

# Get all Kafka ingest sources
data "galaxy_kafka_ingest_sources" "all" {}

# Get all file ingest sources
data "galaxy_file_ingest_sources" "all" {}

# ==========================================
# OUTPUTS
# ==========================================

output "clusters_count" {
  value       = length(data.galaxy_clusters.all.result)
  description = "Total number of clusters"
}

output "users_count" {
  value       = length(data.galaxy_users.all.result)
  description = "Total number of users"
}

output "roles_count" {
  value       = length(data.galaxy_roles.all.result)
  description = "Total number of roles"
}

output "service_accounts_count" {
  value       = length(data.galaxy_service_accounts.all.result)
  description = "Total number of service accounts"
}

output "tags_count" {
  value       = length(data.galaxy_tags.all.result)
  description = "Total number of tags"
}

output "policies_count" {
  value       = length(data.galaxy_policies.all.result)
  description = "Total number of policies"
}

output "data_products_count" {
  value       = length(data.galaxy_data_products.all.result)
  description = "Total number of data products"
}

output "catalogs_count" {
  value       = length(data.galaxy_catalogs.all.result)
  description = "Total number of catalogs"
}

output "s3_catalogs_count" {
  value       = length(data.galaxy_s3_catalogs.all.result)
  description = "Total number of S3 catalogs"
}

output "bigquery_catalogs_count" {
  value       = length(data.galaxy_bigquery_catalogs.all.result)
  description = "Total number of BigQuery catalogs"
}

output "redshift_catalogs_count" {
  value       = length(data.galaxy_redshift_catalogs.all.result)
  description = "Total number of Redshift catalogs"
}

output "postgresql_catalogs_count" {
  value       = length(data.galaxy_postgresql_catalogs.all.result)
  description = "Total number of PostgreSQL catalogs"
}

output "mysql_catalogs_count" {
  value       = length(data.galaxy_mysql_catalogs.all.result)
  description = "Total number of MySQL catalogs"
}

output "gcs_catalogs_count" {
  value       = length(data.galaxy_gcs_catalogs.all.result)
  description = "Total number of GCS catalogs"
}

output "snowflake_catalogs_count" {
  value       = length(data.galaxy_snowflake_catalogs.all.result)
  description = "Total number of Snowflake catalogs"
}

output "sqlserver_catalogs_count" {
  value       = length(data.galaxy_sqlserver_catalogs.all.result)
  description = "Total number of SQL Server catalogs"
}

output "mongodb_catalogs_count" {
  value       = length(data.galaxy_mongodb_catalogs.all.result)
  description = "Total number of MongoDB catalogs"
}

output "cassandra_catalogs_count" {
  value       = length(data.galaxy_cassandra_catalogs.all.result)
  description = "Total number of Cassandra catalogs"
}

output "opensearch_catalogs_count" {
  value       = length(data.galaxy_opensearch_catalogs.all.result)
  description = "Total number of OpenSearch catalogs"
}

output "column_masks_count" {
  value       = length(data.galaxy_column_masks.all.result)
  description = "Total number of column masks"
}

output "row_filters_count" {
  value       = length(data.galaxy_row_filters.all.result)
  description = "Total number of row filters"
}

output "cross_account_iam_roles_count" {
  value       = length(data.galaxy_cross_account_iam_roles.all.result)
  description = "Total number of cross account IAM roles"
}

output "kafka_ingest_sources_count" {
  value       = length(data.galaxy_kafka_ingest_sources.all.result)
  description = "Total number of Kafka ingest sources"
}

output "file_ingest_sources_count" {
  value       = length(data.galaxy_file_ingest_sources.all.result)
  description = "Total number of file ingest sources"
}

# Example of accessing specific attributes from the first cluster (if any)
output "first_cluster_info" {
  value = length(data.galaxy_clusters.all.result) > 0 ? {
    cluster_id = data.galaxy_clusters.all.result[0].cluster_id
    name       = data.galaxy_clusters.all.result[0].name
    state      = data.galaxy_clusters.all.result[0].cluster_state
  } : null
  description = "Information about the first cluster (if any exists)"
}

# Example of accessing specific attributes from the first user (if any)
output "first_user_info" {
  value = length(data.galaxy_users.all.result) > 0 ? {
    user_id = data.galaxy_users.all.result[0].user_id
    email   = data.galaxy_users.all.result[0].email
  } : null
  description = "Information about the first user (if any exists)"
}

# Example of accessing specific attributes from the first Kafka ingest source (if any)
output "first_kafka_ingest_source_info" {
  value = length(data.galaxy_kafka_ingest_sources.all.result) > 0 ? {
    kafka_ingest_source_id = data.galaxy_kafka_ingest_sources.all.result[0].kafka_ingest_source_id
    name                   = data.galaxy_kafka_ingest_sources.all.result[0].name
    description            = data.galaxy_kafka_ingest_sources.all.result[0].description
    kafka_brokers          = data.galaxy_kafka_ingest_sources.all.result[0].kafka_brokers
  } : null
  description = "Information about the first Kafka ingest source (if any exists)"
}

# Example of accessing specific attributes from the first file ingest source (if any)
output "first_file_ingest_source_info" {
  value = length(data.galaxy_file_ingest_sources.all.result) > 0 ? {
    file_ingest_source_id = data.galaxy_file_ingest_sources.all.result[0].file_ingest_source_id
    name                  = data.galaxy_file_ingest_sources.all.result[0].name
    description           = data.galaxy_file_ingest_sources.all.result[0].description
    bucket                = data.galaxy_file_ingest_sources.all.result[0].bucket
    prefix                = data.galaxy_file_ingest_sources.all.result[0].prefix
  } : null
  description = "Information about the first file ingest source (if any exists)"
}