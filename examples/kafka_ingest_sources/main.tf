terraform {
  required_providers {
    galaxy = {
      source = "starburstdata/galaxy"
    }
  }
}

provider "galaxy" {}

locals {
  test_suffix = var.test_suffix != "" ? var.test_suffix : substr(replace(uuid(), "[^0-9]", ""), 0, 6)
}

variable "test_suffix" {
  description = "Suffix to append to resource names for testing"
  type        = string
  default     = ""
}

resource "galaxy_kafka_ingest_source" "example" {
  name            = "kfkls${local.test_suffix}"
  cloud_region_id = "aws-us-east1"
  kafka_brokers   = ["broker1.example.com:9092"]
  kafka_auth = {
    auth_type  = "saslScramAuth"
    scram_type = "Sha256"
    username   = "test"
    password   = "test-password-12345"
  }
}

data "galaxy_kafka_ingest_sources" "all" {
  depends_on = [galaxy_kafka_ingest_source.example]
}

output "kafka_ingest_sources_count" {
  value = length(data.galaxy_kafka_ingest_sources.all.result)
}
