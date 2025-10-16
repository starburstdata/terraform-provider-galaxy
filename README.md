# Terraform Provider for Starburst Galaxy

The Terraform Provider for Starburst Galaxy enables Infrastructure-as-Code management of Starburst Galaxy resources including clusters, catalogs, users, roles, and more.

## Getting Started

### Prerequisites
- Terraform 1.0+
- Go 1.20+ (for local development)
- Starburst Galaxy account with API access
- OAuth2 client credentials (client ID and secret)

### Installation

The provider is available in the Terraform Registry. Add the following to your Terraform configuration:

```hcl
terraform {
  required_providers {
    galaxy = {
      source  = "hashicorp.com/starburstdata/galaxy"
      version = "~> 1.0"
    }
  }
}
```

### Configuration

Configure the provider using environment variables:

```bash
export GALAXY_CLIENT_ID="your-client-id"
export GALAXY_CLIENT_SECRET="your-client-secret"
export GALAXY_DOMAIN="https://your-galaxy-domain.starburstdata.com"
```

In your Terraform configuration:

```hcl
provider "galaxy" {
  # Configuration is read from environment variables
}
```

## Features

- **Cluster Management**: Create and manage Starburst Galaxy clusters with auto-scaling and WarpSpeed capabilities
- **Catalog Configuration**: Connect to various data sources including S3, BigQuery, Snowflake, PostgreSQL, MySQL, and more
- **Access Control**: Manage users, service accounts, roles, and fine-grained permissions
- **Data Governance**: Implement row filters, column masks, and data products for compliance and security
- **Policy Management**: Define and enforce data retention and access policies

### Basic Usage

Create a cluster:

```hcl
resource "galaxy_cluster" "example" {
  name                 = "my-cluster"
  cloud_region_id      = "your-region-id"
  min_workers          = 1
  max_workers          = 10
  private_link_cluster = false
  catalog_refs         = []
  
  processing_mode          = "WarpSpeed"
  warp_resiliency_enabled  = true
  result_cache_enabled     = true
  
  idle_stop_minutes = 30
}
```

Create an S3 catalog:

```hcl
resource "galaxy_s3_catalog" "data_lake" {
  name            = "data-lake"
  cloud_region_id = "your-region-id"
  
  s3_catalog_properties = {
    bucket_name           = "my-data-bucket"
    default_location      = "s3://my-data-bucket/data"
    region               = "us-east-1"
    aws_access_key_id    = var.aws_access_key
    aws_secret_access_key = var.aws_secret_key
  }
  
  cluster_id = galaxy_cluster.example.id
}
```

## Local Development Guide

### Building the Provider

1. Clone the repository:
```bash
git clone https://github.com/starburstdata/terraform-provider-galaxy.git
cd terraform-provider-galaxy
```

2. Build and install the provider:
```bash
go install .
```

This installs the binary to `$GOPATH/bin` (typically `~/go/bin`).

### Configuring Terraform for Local Development

To use your locally-built provider instead of downloading from the registry, configure Terraform with development overrides:

1. Create or edit `~/.terraformrc` (macOS/Linux) or `%APPDATA%/terraform.rc` (Windows):

```hcl
provider_installation {
  dev_overrides {
    "hashicorp.com/starburstdata/galaxy" = "/Users/yourusername/go/bin"  # Replace with your GOBIN path
  }
  
  # For all other providers, install normally from the registry
  direct {}
}
```

2. Verify your GOBIN path:
```bash
go env GOBIN
# If empty, it defaults to:
echo $(go env GOPATH)/bin
```

3. With dev_overrides configured, Terraform will use your local build automatically. You do **not** need to run `terraform init` when using dev_overrides.

### Testing Your Changes

1. Make your code changes
2. Rebuild the provider and install to /go/bin:
```bash
go install .
```

3. Test with Terraform:
```bash
cd examples/test-cluster
terraform plan
terraform apply
```

> **Note**: When using dev_overrides, Terraform will show a warning that development overrides are in effect. This is expected.

### Running the Test Suite

Run unit tests:
```bash
go test ./...
```

Run acceptance tests (requires valid Galaxy credentials):
```bash
export GALAXY_CLIENT_ID="your-client-id"
export GALAXY_CLIENT_SECRET="your-client-secret"
export GALAXY_DOMAIN="https://your-domain.galaxy.starburst.io"

TF_ACC=1 go test ./internal/provider -v -timeout 30m
```

## Resources

- `galaxy_bigquery_catalog` - Google BigQuery catalog
- `galaxy_cassandra_catalog` - Apache Cassandra catalog
- `galaxy_cluster` - Starburst Galaxy clusters
- `galaxy_column_mask` - Column-level data masking
- `galaxy_cross_account_iam_role` - AWS cross-account IAM roles
- `galaxy_data_product` - Data product definitions
- `galaxy_gcs_catalog` - Google Cloud Storage catalog
- `galaxy_mongodb_catalog` - MongoDB catalog
- `galaxy_mysql_catalog` - MySQL database catalog
- `galaxy_opensearch_catalog` - OpenSearch catalog
- `galaxy_policy` - Data governance policies
- `galaxy_postgresql_catalog` - PostgreSQL database catalog
- `galaxy_redshift_catalog` - Amazon Redshift catalog
- `galaxy_role` - Role definitions
- `galaxy_role_privilege_grant` - Role privilege assignments
- `galaxy_row_filter` - Row-level security filters
- `galaxy_s3_catalog` - S3 data lake catalog
- `galaxy_service_account` - Service accounts for automation
- `galaxy_service_account_password` - Service account credentials
- `galaxy_snowflake_catalog` - Snowflake data warehouse catalog
- `galaxy_sql_job` - SQL job definitions
- `galaxy_sqlserver_catalog` - Microsoft SQL Server catalog
- `galaxy_tag` - Data classification tags

## Data Sources

### Single-item Data Sources
- `galaxy_bigquery_catalog` - Read a BigQuery catalog
- `galaxy_cassandra_catalog` - Read a Cassandra catalog
- `galaxy_catalog_metadata` - Read catalog metadata
- `galaxy_cluster` - Read a cluster
- `galaxy_column` - Read a column
- `galaxy_column_mask` - Read a column mask
- `galaxy_data_product` - Read a data product
- `galaxy_data_quality_summary` - Read data quality summary
- `galaxy_gcs_catalog` - Read a GCS catalog
- `galaxy_mongodb_catalog` - Read a MongoDB catalog
- `galaxy_mysql_catalog` - Read a MySQL catalog
- `galaxy_opensearch_catalog` - Read an OpenSearch catalog
- `galaxy_policy` - Read a policy
- `galaxy_postgresql_catalog` - Read a PostgreSQL catalog
- `galaxy_privatelink` - Read a private link
- `galaxy_redshift_catalog` - Read a Redshift catalog
- `galaxy_role` - Read a role
- `galaxy_rolegrant` - Read a role grant
- `galaxy_row_filter` - Read a row filter
- `galaxy_s3_catalog` - Read an S3 catalog
- `galaxy_schema` - Read a schema
- `galaxy_service_account` - Read a service account
- `galaxy_snowflake_catalog` - Read a Snowflake catalog
- `galaxy_sql_job` - Read a SQL job
- `galaxy_sql_job_history` - Read SQL job history
- `galaxy_sql_job_status` - Read SQL job status
- `galaxy_sqlserver_catalog` - Read a SQL Server catalog
- `galaxy_table` - Read a table
- `galaxy_tag` - Read a tag
- `galaxy_user` - Read a user

### List Data Sources
- `galaxy_bigquery_catalogs` - List all BigQuery catalogs
- `galaxy_cassandra_catalogs` - List all Cassandra catalogs
- `galaxy_catalogs` - List all catalogs
- `galaxy_clusters` - List all clusters
- `galaxy_column_masks` - List all column masks
- `galaxy_cross_account_iam_role_metadatas` - List all cross-account IAM role metadata
- `galaxy_cross_account_iam_roles` - List all cross-account IAM roles
- `galaxy_data_products` - List all data products
- `galaxy_data_quality_summaries` - List all data quality summaries
- `galaxy_gcs_catalogs` - List all GCS catalogs
- `galaxy_mongodb_catalogs` - List all MongoDB catalogs
- `galaxy_mysql_catalogs` - List all MySQL catalogs
- `galaxy_opensearch_catalogs` - List all OpenSearch catalogs
- `galaxy_policies` - List all policies
- `galaxy_postgresql_catalogs` - List all PostgreSQL catalogs
- `galaxy_privatelinks` - List all private links
- `galaxy_redshift_catalogs` - List all Redshift catalogs
- `galaxy_roles` - List all roles
- `galaxy_row_filters` - List all row filters
- `galaxy_s3_catalogs` - List all S3 catalogs
- `galaxy_service_accounts` - List all service accounts
- `galaxy_snowflake_catalogs` - List all Snowflake catalogs
- `galaxy_sql_jobs` - List all SQL jobs
- `galaxy_sqlserver_catalogs` - List all SQL Server catalogs
- `galaxy_tags` - List all tags
- `galaxy_users` - List all users

### Validation Data Sources
- `galaxy_bigquery_catalog_validation` - Validate BigQuery catalog configuration
- `galaxy_cassandra_catalog_validation` - Validate Cassandra catalog configuration
- `galaxy_gcs_catalog_validation` - Validate GCS catalog configuration
- `galaxy_mongodb_catalog_validation` - Validate MongoDB catalog configuration
- `galaxy_mysql_catalog_validation` - Validate MySQL catalog configuration
- `galaxy_opensearch_catalog_validation` - Validate OpenSearch catalog configuration
- `galaxy_postgresql_catalog_validation` - Validate PostgreSQL catalog configuration
- `galaxy_redshift_catalog_validation` - Validate Redshift catalog configuration
- `galaxy_s3_catalog_validation` - Validate S3 catalog configuration
- `galaxy_snowflake_catalog_validation` - Validate Snowflake catalog configuration
- `galaxy_sqlserver_catalog_validation` - Validate SQL Server catalog configuration

## Examples

See the [examples](./examples) directory for complete configuration examples:
- [Cluster management](./examples/cluster.tf)
- [S3 catalog setup](./examples/s3_catalog.tf)
- [User and role management](./examples/roles_and_permissions.tf)
- [Data governance](./examples/data_governance.tf)
- [Multiple catalog types](./examples/other_catalogs.tf)

## Support

- **Documentation**: [Starburst Galaxy Documentation](https://docs.starburst.io/starburst-galaxy/)
- **Issues**: [GitHub Issues](https://github.com/starburstdata/terraform-provider-galaxy/issues)
- **Community**: [Starburst Community](https://www.starburst.io/community/)

## License

TBD

## Acknowledgments

This provider is built using:
- [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework)
- [Starburst Galaxy API](https://docs.starburst.io/starburst-galaxy/developer-tools/api/)

---

Made with ❤️ by the Starburst team
