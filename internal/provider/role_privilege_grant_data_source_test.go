// Copyright (c) Starburst Data, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDataSourceRolePrivilegeGrants_Basic(t *testing.T) {
	suffix := testAccSuffix()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: create the grant; data source may return empty due to API eventual consistency
				Config: testAccDataSourceRolePrivilegeGrantsConfig(suffix),
			},
			{
				// Step 2: refresh - grant is now visible in the API cache
				Config: testAccDataSourceRolePrivilegeGrantsConfig(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.galaxy_role_privilege_grants.test",
						tfjsonpath.New("role_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_role_privilege_grants.test",
						tfjsonpath.New("result"),
						knownvalue.ListPartial(map[int]knownvalue.Check{
							0: knownvalue.ObjectExact(map[string]knownvalue.Check{
								"entity_kind":  knownvalue.StringExact("Catalog"),
								"privilege":    knownvalue.StringExact("CreateSchema"),
								"grant_kind":   knownvalue.StringExact("Allow"),
								"grant_option": knownvalue.Bool(false),
								"entity_id":    knownvalue.NotNull(),
								"schema_name":  knownvalue.Null(),
								"table_name":   knownvalue.Null(),
								"column_name":  knownvalue.Null(),
							}),
						}),
					),
				},
			},
		},
	})
}

func TestAccDataSourceRolePrivilegeGrants_Empty(t *testing.T) {
	suffix := testAccSuffix()
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRolePrivilegeGrantsEmptyConfig(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.galaxy_role_privilege_grants.empty",
						tfjsonpath.New("result"),
						knownvalue.ListSizeExact(0),
					),
				},
			},
		},
	})
}

func testAccSuffix() string {
	s := id.UniqueId()
	if len(s) > 8 {
		return s[len(s)-8:]
	}
	return s
}

func testAccDataSourceRolePrivilegeGrantsConfig(suffix string) string {
	return fmt.Sprintf(`
%s

resource "galaxy_postgresql_catalog" "test" {
  name          = "rpgdscat_%[2]s"
  endpoint      = var.TESTING_POSTGRESQL_AWS_HOST
  port          = 5432
  database_name = var.TESTING_POSTGRESQL_AWS_DATABASE
  username      = var.TESTING_POSTGRESQL_AWS_USERNAME
  password      = var.TESTING_POSTGRESQL_AWS_PASSWORD
  read_only     = false
}

resource "galaxy_role" "test" {
  role_name              = "rpgdsrole_%[2]s"
  grant_to_creating_role = false
}

resource "galaxy_role_privilege_grant" "test" {
  role_id      = galaxy_role.test.role_id
  entity_id    = galaxy_postgresql_catalog.test.catalog_id
  entity_kind  = "Catalog"
  privilege    = "CreateSchema"
  grant_kind   = "Allow"
  grant_option = false
}

data "galaxy_role_privilege_grants" "test" {
  role_id    = galaxy_role.test.role_id
  depends_on = [galaxy_role_privilege_grant.test]
}
`, testAccPostgresVariables(), suffix)
}

func testAccDataSourceRolePrivilegeGrantsEmptyConfig(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_role" "empty" {
  role_name              = "rpgdsempty_%[1]s"
  grant_to_creating_role = false
}

data "galaxy_role_privilege_grants" "empty" {
  role_id    = galaxy_role.empty.role_id
  depends_on = [galaxy_role.empty]
}
`, suffix)
}

func testAccPostgresVariables() string {
	return `
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
}`
}
