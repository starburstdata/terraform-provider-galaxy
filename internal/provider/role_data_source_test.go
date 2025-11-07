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

func TestAccDataSourceRole_Basic(t *testing.T) {
	// Generate a short random suffix to avoid conflicts with leftover resources
	uniqueId := id.UniqueId()
	// Take only the last 8 characters to keep the name short
	if len(uniqueId) > 8 {
		uniqueId = uniqueId[len(uniqueId)-8:]
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create role and read via data source
			{
				Config: testAccDataSourceRoleConfig(uniqueId),
				ConfigStateChecks: []statecheck.StateCheck{
					// Check resource
					statecheck.ExpectKnownValue(
						"galaxy_role.test",
						tfjsonpath.New("role_name"),
						knownvalue.StringExact(fmt.Sprintf("testrole_%s", uniqueId)),
					),
					// Check data source reads same values
					statecheck.ExpectKnownValue(
						"data.galaxy_role.test",
						tfjsonpath.New("role_name"),
						knownvalue.StringExact(fmt.Sprintf("testrole_%s", uniqueId)),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_role.test",
						tfjsonpath.New("role_id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccDataSourceRoles_List(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// List all roles
			{
				Config: testAccDataSourceRolesConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Check that result is not null
					statecheck.ExpectKnownValue(
						"data.galaxy_roles.all",
						tfjsonpath.New("result"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// testAccDataSourceRoleConfig returns a configuration for testing role data source
func testAccDataSourceRoleConfig(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_role" "test" {
  role_name              = "testrole_%[1]s"
  grant_to_creating_role = true
  role_description       = "Role for data source acceptance testing"
}

data "galaxy_role" "test" {
  role_id = galaxy_role.test.role_id
}
`, suffix)
}

// testAccDataSourceRolesConfig returns a configuration for testing roles data source
func testAccDataSourceRolesConfig() string {
	return `
data "galaxy_roles" "all" {}
`
}
