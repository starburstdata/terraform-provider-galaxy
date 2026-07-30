// Copyright (c) Starburst Data, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccResourceUsageExample_Basic(t *testing.T) {
	suffix := testSuffix

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccUsageExampleResourceConfig(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_usage_example.test",
						tfjsonpath.New("usage_example_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_usage_example.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("ue_%s", suffix)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_usage_example.test",
						tfjsonpath.New("code"),
						knownvalue.StringExact("SELECT 1"),
					),
				},
			},
			// Update and Read testing
			{
				Config: testAccUsageExampleResourceConfigUpdate(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_usage_example.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("ue_%s", suffix)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_usage_example.test",
						tfjsonpath.New("code"),
						knownvalue.StringExact("SELECT 2"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_usage_example.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Updated usage example description"),
					),
				},
			},
		},
	})
}

func testAccUsageExampleResourceConfig(suffix string) string {
	return fmt.Sprintf(`
variable "TESTING_POSTGRESQL_AWS_HOST" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_DATABASE" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_USERNAME" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_PASSWORD" {
  type      = string
  sensitive = true
}

resource "galaxy_postgresql_catalog" "test" {
  name          = "uecat%[1]s"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = true
}

data "galaxy_users" "all" {}

locals {
  first_user = data.galaxy_users.all.result[0]
}

resource "galaxy_data_product" "test" {
  name        = "uedp%[1]s"
  catalog_id  = galaxy_postgresql_catalog.test.catalog_id
  schema_name = "public"
  summary     = "Data product for usage example testing"

  contacts = [
    {
      email   = local.first_user.email
      user_id = local.first_user.user_id
    }
  ]
}

resource "galaxy_usage_example" "test" {
  data_product_id = galaxy_data_product.test.data_product_id
  name            = "ue_%[1]s"
  code            = "SELECT 1"
}
`, suffix)
}

func testAccUsageExampleResourceConfigUpdate(suffix string) string {
	return fmt.Sprintf(`
variable "TESTING_POSTGRESQL_AWS_HOST" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_DATABASE" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_USERNAME" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_PASSWORD" {
  type      = string
  sensitive = true
}

resource "galaxy_postgresql_catalog" "test" {
  name          = "uecat%[1]s"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = true
}

data "galaxy_users" "all" {}

locals {
  first_user = data.galaxy_users.all.result[0]
}

resource "galaxy_data_product" "test" {
  name        = "uedp%[1]s"
  catalog_id  = galaxy_postgresql_catalog.test.catalog_id
  schema_name = "public"
  summary     = "Data product for usage example testing"

  contacts = [
    {
      email   = local.first_user.email
      user_id = local.first_user.user_id
    }
  ]
}

resource "galaxy_usage_example" "test" {
  data_product_id = galaxy_data_product.test.data_product_id
  name            = "ue_%[1]s"
  code            = "SELECT 2"
  description     = "Updated usage example description"
}
`, suffix)
}
