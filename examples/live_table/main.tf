terraform {
  required_providers {
    galaxy = {
      source = "starburstdata/galaxy"
    }
  }
}

provider "galaxy" {}

# galaxy_live_table resource creation requires a pre-existing metastore database and bucket
# whose paths do not overlap with the source. The integration test environment does not
# provision those prerequisites, so this example demonstrates the plural data source only.
# See acceptance test TestAccResourceLiveTable_Basic for the resource-level config, which is
# skipped unless GALAXY_LIVE_TABLE_CATALOG_ID and GALAXY_LIVE_TABLE_INGEST_SOURCE_ID are set.

data "galaxy_live_tables" "all" {}

output "live_tables_count" {
  value = length(data.galaxy_live_tables.all.result)
}
