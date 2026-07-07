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

func TestAccResourceDataQualityCheck_Basic(t *testing.T) {
	suffix := testSuffix

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDataQualityCheckResourceConfig(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_data_quality_check.test",
						tfjsonpath.New("data_quality_check_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_data_quality_check.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("dqcheck_%s", suffix)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_data_quality_check.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Test data quality check"),
					),
				},
			},
			// Import testing
			{
				ResourceName:                         "galaxy_data_quality_check.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateIdFunc("galaxy_data_quality_check.test", "data_quality_check_id"),
				ImportStateVerifyIdentifierAttribute: "data_quality_check_id",
				ImportStateVerifyIgnore: []string{
					"cluster_id", // write-only for check execution
				},
			},
			// Update and Read testing
			{
				Config: testAccDataQualityCheckResourceConfigUpdate(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_data_quality_check.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Updated data quality check description"),
					),
				},
			},
		},
	})
}

func testAccDataQualityCheckResourceConfig(suffix string) string {
	return fmt.Sprintf(`
variable "TESTING_POSTGRESQL_AWS_HOST" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_DATABASE" {
  type = string
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
  name          = "dqchk%[1]s"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = false
  description   = "PostgreSQL catalog for data quality check testing"
}

resource "galaxy_cluster" "test" {
  name                    = "dqchk%[1]s"
  cloud_region_id         = "aws-us-east1"
  min_workers             = 1
  max_workers             = 1
  idle_stop_minutes       = 15
  private_link_cluster    = false
  result_cache_enabled    = false
  warp_resiliency_enabled = false
  catalog_refs            = [galaxy_postgresql_catalog.test.catalog_id]
}

resource "galaxy_role" "dq_check_grant" {
  role_name              = "dqchkgrant%[1]s"
  role_description       = "Role granting query access to the data quality check test catalog"
  grant_to_creating_role = true
}

resource "galaxy_role_privilege_grant" "dq_check_grant" {
  role_id      = galaxy_role.dq_check_grant.role_id
  entity_id    = galaxy_postgresql_catalog.test.catalog_id
  entity_kind  = "Column"
  privilege    = "Select"
  grant_kind   = "Allow"
  grant_option = false
  schema_name  = "*"
  table_name   = "*"
  column_name  = "*"
}

resource "galaxy_data_quality_check" "test" {
  name        = "dqcheck_%[1]s"
  description = "Test data quality check"
  catalog_id  = galaxy_postgresql_catalog.test.catalog_id
  schema_id   = "anu_test"
  table_id    = "employees"
  severity    = "Low"
  category    = "Completeness"
  kind        = "SqlQuery"
  cluster_id  = galaxy_cluster.test.cluster_id
  query       = "select exists(select * from ${galaxy_postgresql_catalog.test.name}.anu_test.employees)"
  depends_on  = [galaxy_role_privilege_grant.dq_check_grant]
}
`, suffix)
}

func testAccDataQualityCheckResourceConfigUpdate(suffix string) string {
	return fmt.Sprintf(`
variable "TESTING_POSTGRESQL_AWS_HOST" {
  type      = string
  sensitive = true
}

variable "TESTING_POSTGRESQL_AWS_DATABASE" {
  type = string
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
  name          = "dqchk%[1]s"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = false
  description   = "PostgreSQL catalog for data quality check testing"
}

resource "galaxy_cluster" "test" {
  name                    = "dqchk%[1]s"
  cloud_region_id         = "aws-us-east1"
  min_workers             = 1
  max_workers             = 1
  idle_stop_minutes       = 15
  private_link_cluster    = false
  result_cache_enabled    = false
  warp_resiliency_enabled = false
  catalog_refs            = [galaxy_postgresql_catalog.test.catalog_id]
}

resource "galaxy_role" "dq_check_grant" {
  role_name              = "dqchkgrant%[1]s"
  role_description       = "Role granting query access to the data quality check test catalog"
  grant_to_creating_role = true
}

resource "galaxy_role_privilege_grant" "dq_check_grant" {
  role_id      = galaxy_role.dq_check_grant.role_id
  entity_id    = galaxy_postgresql_catalog.test.catalog_id
  entity_kind  = "Column"
  privilege    = "Select"
  grant_kind   = "Allow"
  grant_option = false
  schema_name  = "*"
  table_name   = "*"
  column_name  = "*"
}

resource "galaxy_data_quality_check" "test" {
  name        = "dqcheck_%[1]s"
  description = "Updated data quality check description"
  catalog_id  = galaxy_postgresql_catalog.test.catalog_id
  schema_id   = "anu_test"
  table_id    = "employees"
  severity    = "Low"
  category    = "Completeness"
  kind        = "SqlQuery"
  cluster_id  = galaxy_cluster.test.cluster_id
  query       = "select exists(select * from ${galaxy_postgresql_catalog.test.name}.anu_test.employees)"
  depends_on  = [galaxy_role_privilege_grant.dq_check_grant]
}
`, suffix)
}
