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
  description = "Testing S3 bucket"
  type        = string
  default     = "integration-tests-bucket"
}

variable "TESTING_S3_IAM_ROLE_ARN" {
  description = "Testing S3 external IAM role ARN"
  type        = string
  default     = "arn:aws:iam::000000000000:role/placeholder"
}

resource "galaxy_file_ingest_source" "example" {
  name                   = "file${local.test_suffix}"
  description            = "E2E file ingest source"
  bucket                 = var.TESTING_S3_BUCKET
  bucket_cloud_region_id = "aws-us-east1"
  source_cloud_region_id = "aws-us-east1"
  path_prefix            = "ingest/${local.test_suffix}"
  cloud_storage_credentials = {
    credentials_type  = "externalIamRole"
    external_role_arn = var.TESTING_S3_IAM_ROLE_ARN
  }
}

data "galaxy_file_ingest_source" "example" {
  file_ingest_source_id = galaxy_file_ingest_source.example.file_ingest_source_id
}

output "file_ingest_source_id" {
  value = galaxy_file_ingest_source.example.file_ingest_source_id
}

output "file_ingest_source_name" {
  value = data.galaxy_file_ingest_source.example.name
}
