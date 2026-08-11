// Copyright (c) Starburst Data, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDataSourceDataQualitySchedules_Basic(t *testing.T) {
	if os.Getenv("TF_VAR_TESTING_ACTIVE_CLUSTER_ID") == "" {
		t.Skip("TF_VAR_TESTING_ACTIVE_CLUSTER_ID must be set for data_quality_schedules data source tests")
	}

	suffix := testSuffix

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDataQualitySchedulesConfig(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.galaxy_data_quality_schedules.test",
						tfjsonpath.New("catalog_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_data_quality_schedules.test",
						tfjsonpath.New("data_quality_checks"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccDataSourceDataQualitySchedulesConfig(suffix string) string {
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
  name          = "dqsched%[1]s"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = false
  description   = "PostgreSQL catalog for data quality schedules data source test"
}

resource "galaxy_cluster" "test" {
  name                    = "dqsched%[1]s"
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
  role_name              = "dqschedgrant%[1]s"
  role_description       = "Role for data quality schedules data source test"
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
  name        = "dqschedcheck_%[1]s"
  description = "Data quality check for schedules data source test"
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

resource "galaxy_data_quality_schedule" "test" {
  catalog_id      = galaxy_postgresql_catalog.test.catalog_id
  schema_id       = "anu_test"
  table_id        = "employees"
  cluster_id      = galaxy_cluster.test.cluster_id
  cron_expression = "0 0 * * *"
  role_id         = galaxy_role.dq_check_grant.role_id
  depends_on      = [galaxy_data_quality_check.test]
}

data "galaxy_data_quality_schedules" "test" {
  catalog_id = galaxy_postgresql_catalog.test.catalog_id
  schema_id  = "anu_test"
  table_id   = "employees"
  depends_on = [galaxy_data_quality_schedule.test]
}
`, suffix)
}
