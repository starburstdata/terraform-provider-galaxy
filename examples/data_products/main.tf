terraform {
  required_providers {
    galaxy = {
      source = "starburstdata/galaxy"
    }
  }
}

provider "galaxy" {
  # Credentials from environment variables
}

# List all data products
data "galaxy_data_products" "all" {}

output "data_products_count" {
  value = length(try(data.galaxy_data_products.all.result, []))
}

output "data_products_result" {
  value = try(data.galaxy_data_products.all.result, [])
}
