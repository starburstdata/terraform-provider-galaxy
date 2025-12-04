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

func TestAccDataSourceRoleGrant_Basic(t *testing.T) {
	// Generate a short random suffix to avoid conflicts with leftover resources
	uniqueId := id.UniqueId()
	if len(uniqueId) > 8 {
		uniqueId = uniqueId[len(uniqueId)-8:]
	}
	suffix := uniqueId
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRoleGrantConfig(suffix),
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

func testAccDataSourceRoleGrantConfig(suffix string) string {
	return fmt.Sprintf(`
# Create a test role
resource "galaxy_role" "test" {
  role_name              = "rolegrantds%[1]s"
  grant_to_creating_role = true
}

# Read rolegrants for the test role
data "galaxy_rolegrant" "test" {
  role_id = galaxy_role.test.role_id
}
`, suffix)
}
