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

func TestAccResourceServiceAccount_Basic(t *testing.T) {
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
			// Create and Read testing
			{
				Config: testAccServiceAccountConfigBasic(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_service_account.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(fmt.Sprintf("tfaccsa_%s", suffix)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_service_account.test",
						tfjsonpath.New("service_account_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_service_account.test",
						tfjsonpath.New("with_initial_password"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

func TestAccResourceServiceAccount_WithRoles(t *testing.T) {
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
			// Create service account with role
			{
				Config: testAccServiceAccountConfigWithRoles(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_service_account.test_with_role",
						tfjsonpath.New("username"),
						knownvalue.StringExact(fmt.Sprintf("tfaccsa_%s", suffix)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_service_account.test_with_role",
						tfjsonpath.New("additional_role_ids"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// testAccServiceAccountConfigBasic returns a basic service account configuration
func testAccServiceAccountConfigBasic(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_service_account" "test" {
  username              = "tfaccsa_%[1]s"
  with_initial_password = true
  additional_role_ids   = []
}
`, suffix)
}

// testAccServiceAccountConfigWithRoles returns a service account configuration with roles
func testAccServiceAccountConfigWithRoles(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_role" "automation" {
  role_name              = "tfaccrole_%[1]s"
  grant_to_creating_role = true
}

resource "galaxy_service_account" "test_with_role" {
  username              = "tfaccsa_%[1]s"
  with_initial_password = true

  additional_role_ids = [
    galaxy_role.automation.role_id
  ]
}
`, suffix)
}
