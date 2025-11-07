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

func TestAccDataSourceUser_Basic(t *testing.T) {
	// Get a user from the list and then fetch that specific user
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceUserConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify we can read a user's details
					statecheck.ExpectKnownValue(
						"data.galaxy_user.test",
						tfjsonpath.New("user_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_user.test",
						tfjsonpath.New("email"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccDataSourceUsers_List(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// List all users
			{
				Config: testAccDataSourceUsersConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Check that result is not null
					statecheck.ExpectKnownValue(
						"data.galaxy_users.all",
						tfjsonpath.New("result"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// testAccDataSourceUserConfig returns a configuration for testing user data source
// It uses the users list to get the first user's ID
func testAccDataSourceUserConfig() string {
	return `
data "galaxy_users" "all" {}

data "galaxy_user" "test" {
  user_id = data.galaxy_users.all.result[0].user_id
}
`
}

// testAccDataSourceUsersConfig returns a configuration for testing users data source
func testAccDataSourceUsersConfig() string {
	return `
data "galaxy_users" "all" {}
`
}
