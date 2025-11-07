// Copyright (c) Starburst Data, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccDataSourceRoleGrant_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRoleGrantConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.galaxy_rolegrant.test",
						tfjsonpath.New("role_id"),
						knownvalue.NotNull(),
					),
					// Result can be empty list if no grants
				},
			},
		},
	})
}

func testAccDataSourceRoleGrantConfig() string {
	return `
# Create a test role
resource "galaxy_role" "test" {
  role_name              = "test_rolegrant_ds"
  grant_to_creating_role = true
}

# Read rolegrants for the test role
data "galaxy_rolegrant" "test" {
  role_id = galaxy_role.test.role_id
}
`
}
