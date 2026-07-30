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

func TestAccResourceDataQualitySchedule_Basic(t *testing.T) {
	if os.Getenv("TF_VAR_TESTING_ACTIVE_CLUSTER_ID") == "" {
		t.Skip("TF_VAR_TESTING_ACTIVE_CLUSTER_ID must be set for data_quality_schedule resource tests")
	}

	suffix := testSuffix

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDataQualityScheduleResourceConfig(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_data_quality_schedule.test",
						tfjsonpath.New("data_quality_schedule_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_data_quality_schedule.test",
						tfjsonpath.New("cron_expression"),
						knownvalue.StringExact("0 0 * * *"),
					),
				},
			},
			// Update and Read testing
			{
				Config: testAccDataQualityScheduleResourceConfigUpdate(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_data_quality_schedule.test",
						tfjsonpath.New("cron_expression"),
						knownvalue.StringExact("0 6 * * *"),
					),
				},
			},
		},
	})
}

func testAccDataQualityScheduleResourceConfig(suffix string) string {
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
  name          = "dqschres%[1]s"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = false
  description   = "PostgreSQL catalog for data quality schedule resource test"
}

resource "galaxy_cluster" "test" {
  name                    = "dqschres%[1]s"
  cloud_region_id         = "aws-us-east1"
  min_workers             = 1
  max_workers             = 1
  idle_stop_minutes       = 15
  private_link_cluster    = false
  result_cache_enabled    = false
  warp_resiliency_enabled = false
  catalog_refs            = [galaxy_postgresql_catalog.test.catalog_id]
}

resource "galaxy_role" "dq_schedule_grant" {
  role_name              = "dqschrgrant%[1]s"
  role_description       = "Role for data quality schedule resource test"
  grant_to_creating_role = true
}

resource "galaxy_role_privilege_grant" "dq_schedule_grant" {
  role_id      = galaxy_role.dq_schedule_grant.role_id
  entity_id    = galaxy_postgresql_catalog.test.catalog_id
  entity_kind  = "Column"
  privilege    = "Select"
  grant_kind   = "Allow"
  grant_option = false
  schema_name  = "*"
  table_name   = "*"
  column_name  = "*"
}

# Creating a cluster only grants the creator CREATE_CLUSTER on it - it does NOT grant
# USE_CLUSTER to the role that executes checks/schedules on the cluster. Both
# galaxy_data_quality_check and galaxy_data_quality_schedule validate USE_CLUSTER on the
# cluster for role_id independently, so without this grant creation fails with a 403
# PERMISSION_DENIED (GET_CLUSTER) error.
resource "galaxy_role_privilege_grant" "dq_schedule_use_cluster" {
  role_id      = galaxy_role.dq_schedule_grant.role_id
  entity_id    = galaxy_cluster.test.cluster_id
  entity_kind  = "Cluster"
  privilege    = "UseCluster"
  grant_kind   = "Allow"
  grant_option = false
}

resource "galaxy_data_quality_check" "test" {
  name        = "dqschrcheck_%[1]s"
  description = "Data quality check for schedule resource test"
  catalog_id  = galaxy_postgresql_catalog.test.catalog_id
  schema_id   = "anu_test"
  table_id    = "employees"
  severity    = "Low"
  category    = "Completeness"
  kind        = "SqlQuery"
  cluster_id  = galaxy_cluster.test.cluster_id
  query       = "select exists(select * from ${galaxy_postgresql_catalog.test.name}.anu_test.employees)"
  depends_on  = [galaxy_role_privilege_grant.dq_schedule_grant, galaxy_role_privilege_grant.dq_schedule_use_cluster]
}

resource "galaxy_data_quality_schedule" "test" {
  catalog_id      = galaxy_postgresql_catalog.test.catalog_id
  schema_id       = "anu_test"
  table_id        = "employees"
  cluster_id      = galaxy_cluster.test.cluster_id
  cron_expression = "0 0 * * *"
  role_id         = galaxy_role.dq_schedule_grant.role_id
  depends_on      = [galaxy_data_quality_check.test, galaxy_role_privilege_grant.dq_schedule_use_cluster]
}
`, suffix)
}

func testAccDataQualityScheduleResourceConfigUpdate(suffix string) string {
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
  name          = "dqschres%[1]s"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = false
  description   = "PostgreSQL catalog for data quality schedule resource test"
}

resource "galaxy_cluster" "test" {
  name                    = "dqschres%[1]s"
  cloud_region_id         = "aws-us-east1"
  min_workers             = 1
  max_workers             = 1
  idle_stop_minutes       = 15
  private_link_cluster    = false
  result_cache_enabled    = false
  warp_resiliency_enabled = false
  catalog_refs            = [galaxy_postgresql_catalog.test.catalog_id]
}

resource "galaxy_role" "dq_schedule_grant" {
  role_name              = "dqschrgrant%[1]s"
  role_description       = "Role for data quality schedule resource test"
  grant_to_creating_role = true
}

resource "galaxy_role_privilege_grant" "dq_schedule_grant" {
  role_id      = galaxy_role.dq_schedule_grant.role_id
  entity_id    = galaxy_postgresql_catalog.test.catalog_id
  entity_kind  = "Column"
  privilege    = "Select"
  grant_kind   = "Allow"
  grant_option = false
  schema_name  = "*"
  table_name   = "*"
  column_name  = "*"
}

# Creating a cluster only grants the creator CREATE_CLUSTER on it - it does NOT grant
# USE_CLUSTER to the role that executes checks/schedules on the cluster. Both
# galaxy_data_quality_check and galaxy_data_quality_schedule validate USE_CLUSTER on the
# cluster for role_id independently, so without this grant creation fails with a 403
# PERMISSION_DENIED (GET_CLUSTER) error.
resource "galaxy_role_privilege_grant" "dq_schedule_use_cluster" {
  role_id      = galaxy_role.dq_schedule_grant.role_id
  entity_id    = galaxy_cluster.test.cluster_id
  entity_kind  = "Cluster"
  privilege    = "UseCluster"
  grant_kind   = "Allow"
  grant_option = false
}

resource "galaxy_data_quality_check" "test" {
  name        = "dqschrcheck_%[1]s"
  description = "Data quality check for schedule resource test"
  catalog_id  = galaxy_postgresql_catalog.test.catalog_id
  schema_id   = "anu_test"
  table_id    = "employees"
  severity    = "Low"
  category    = "Completeness"
  kind        = "SqlQuery"
  cluster_id  = galaxy_cluster.test.cluster_id
  query       = "select exists(select * from ${galaxy_postgresql_catalog.test.name}.anu_test.employees)"
  depends_on  = [galaxy_role_privilege_grant.dq_schedule_grant, galaxy_role_privilege_grant.dq_schedule_use_cluster]
}

resource "galaxy_data_quality_schedule" "test" {
  catalog_id      = galaxy_postgresql_catalog.test.catalog_id
  schema_id       = "anu_test"
  table_id        = "employees"
  cluster_id      = galaxy_cluster.test.cluster_id
  cron_expression = "0 6 * * *"
  role_id         = galaxy_role.dq_schedule_grant.role_id
  depends_on      = [galaxy_data_quality_check.test, galaxy_role_privilege_grant.dq_schedule_use_cluster]
}
`, suffix)
}
