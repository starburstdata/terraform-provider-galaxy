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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccResourceRoleGrant_Basic(t *testing.T) {
	uniqueId := id.UniqueId()
	if len(uniqueId) > 8 {
		uniqueId = uniqueId[len(uniqueId)-8:]
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleGrantConfigBasic(uniqueId),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_role_grant.test",
						tfjsonpath.New("role_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_role_grant.test",
						tfjsonpath.New("granted_role_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_role_grant.test",
						tfjsonpath.New("admin_option"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"galaxy_role_grant.test",
						tfjsonpath.New("granted_role_name"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccResourceRoleGrant_WithAdminOption(t *testing.T) {
	uniqueId := id.UniqueId()
	if len(uniqueId) > 8 {
		uniqueId = uniqueId[len(uniqueId)-8:]
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleGrantConfigAdminOption(uniqueId),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_role_grant.test",
						tfjsonpath.New("role_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_role_grant.test",
						tfjsonpath.New("granted_role_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_role_grant.test",
						tfjsonpath.New("admin_option"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"galaxy_role_grant.test",
						tfjsonpath.New("granted_role_name"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccResourceRoleGrant_ImportState(t *testing.T) {
	uniqueId := id.UniqueId()
	if len(uniqueId) > 8 {
		uniqueId = uniqueId[len(uniqueId)-8:]
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoleGrantConfigBasic(uniqueId),
			},
			{
				ResourceName:                         "galaxy_role_grant.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "role_id",
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["galaxy_role_grant.test"]
					if !ok {
						return "", fmt.Errorf("Resource not found in state")
					}
					roleID := rs.Primary.Attributes["role_id"]
					grantedRoleID := rs.Primary.Attributes["granted_role_id"]
					if roleID == "" || grantedRoleID == "" {
						return "", fmt.Errorf("role_id or granted_role_id not found in state")
					}
					return fmt.Sprintf("%s/%s", roleID, grantedRoleID), nil
				},
			},
		},
	})
}

func testAccRoleGrantConfigBasic(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_role" "parent" {
  role_name              = "grantparent_%[1]s"
  grant_to_creating_role = true
}

resource "galaxy_role" "child" {
  role_name              = "grantchild_%[1]s"
  grant_to_creating_role = true
}

resource "galaxy_role_grant" "test" {
  role_id         = galaxy_role.parent.role_id
  granted_role_id = galaxy_role.child.role_id
  admin_option    = false
}
`, suffix)
}

func testAccRoleGrantConfigAdminOption(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_role" "parent" {
  role_name              = "grantparent_%[1]s"
  grant_to_creating_role = true
}

resource "galaxy_role" "child" {
  role_name              = "grantchild_%[1]s"
  grant_to_creating_role = true
}

resource "galaxy_role_grant" "test" {
  role_id         = galaxy_role.parent.role_id
  granted_role_id = galaxy_role.child.role_id
  admin_option    = true
}
`, suffix)
}
