terraform {
  required_providers {
    galaxy = {
      source = "starburstdata/galaxy"
    }
  }
}

provider "galaxy" {}

data "galaxy_live_tables" "all" {}

output "live_tables_count" {
  value = length(data.galaxy_live_tables.all.result)
}
