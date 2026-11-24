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

variable "TESTING_S3_BUCKET" {
  type    = string
  default = "integration-tests-bucket"
}

variable "TESTING_S3_IAM_ROLE_ARN" {
  type    = string
  default = "arn:aws:iam::000000000000:role/placeholder"
}

resource "galaxy_file_ingest_source" "example" {
  name                   = "fls${local.test_suffix}"
  bucket                 = var.TESTING_S3_BUCKET
  bucket_cloud_region_id = "aws-us-east1"
  source_cloud_region_id = "aws-us-east1"
  cloud_storage_credentials = {
    credentials_type  = "externalIamRole"
    external_role_arn = var.TESTING_S3_IAM_ROLE_ARN
  }
}

data "galaxy_file_ingest_sources" "all" {
  depends_on = [galaxy_file_ingest_source.example]
}

output "file_ingest_sources_count" {
  value = length(data.galaxy_file_ingest_sources.all.result)
}
