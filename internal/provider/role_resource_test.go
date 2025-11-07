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

func TestAccResourceRole_Basic(t *testing.T) {
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
			// Create and Read testing
			{
				Config: testAccRoleConfigBasic(uniqueId),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_role.test",
						tfjsonpath.New("role_name"),
						knownvalue.StringExact(fmt.Sprintf("testrole_%s", uniqueId)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_role.test",
						tfjsonpath.New("role_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_role.test",
						tfjsonpath.New("grant_to_creating_role"),
						knownvalue.Bool(true),
					),
				},
			},
			// Update and Read testing
			{
				Config: testAccRoleConfigUpdate(uniqueId),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_role.test",
						tfjsonpath.New("role_description"),
						knownvalue.StringExact("Updated role description for acceptance testing"),
					),
				},
			},
		},
	})
}

func TestAccResourceRole_WithDescription(t *testing.T) {
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
			// Create and Read testing with description
			{
				Config: testAccRoleConfigWithDescription(uniqueId),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_role.test_desc",
						tfjsonpath.New("role_name"),
						knownvalue.StringExact(fmt.Sprintf("testrole_%s", uniqueId)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_role.test_desc",
						tfjsonpath.New("role_description"),
						knownvalue.StringExact("Role for acceptance testing with description"),
					),
				},
			},
		},
	})
}

func TestAccResourceRole_MultipleRoles(t *testing.T) {
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
			// Create multiple roles
			{
				Config: testAccRoleConfigMultiple(uniqueId),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_role.admin",
						tfjsonpath.New("role_name"),
						knownvalue.StringExact(fmt.Sprintf("admin_%s", uniqueId)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_role.readonly",
						tfjsonpath.New("role_name"),
						knownvalue.StringExact(fmt.Sprintf("readonly_%s", uniqueId)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_role.dataeng",
						tfjsonpath.New("role_name"),
						knownvalue.StringExact(fmt.Sprintf("dataeng_%s", uniqueId)),
					),
				},
			},
		},
	})
}

// testAccRoleConfigBasic returns a basic role configuration
func testAccRoleConfigBasic(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_role" "test" {
  role_name              = "testrole_%[1]s"
  grant_to_creating_role = true
}
`, suffix)
}

// testAccRoleConfigUpdate returns an updated role configuration
func testAccRoleConfigUpdate(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_role" "test" {
  role_name              = "testrole_%[1]s"
  grant_to_creating_role = true
  role_description       = "Updated role description for acceptance testing"
}
`, suffix)
}

// testAccRoleConfigWithDescription returns a role configuration with description
func testAccRoleConfigWithDescription(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_role" "test_desc" {
  role_name              = "testrole_%[1]s"
  grant_to_creating_role = true
  role_description       = "Role for acceptance testing with description"
}
`, suffix)
}

// testAccRoleConfigMultiple returns a configuration with multiple roles
func testAccRoleConfigMultiple(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_role" "admin" {
  role_name              = "admin_%[1]s"
  grant_to_creating_role = true
  role_description       = "Administrative role with full permissions"
}

resource "galaxy_role" "readonly" {
  role_name              = "readonly_%[1]s"
  grant_to_creating_role = true
  role_description       = "Read-only role for data access"
}

resource "galaxy_role" "dataeng" {
  role_name              = "dataeng_%[1]s"
  grant_to_creating_role = true
  role_description       = "Data engineering role with write access"
}
`, suffix)
}
