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

func TestAccResourceServiceAccountPassword_Basic(t *testing.T) {
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
			// Note: Passwords cannot be updated after creation per API limitations
			{
				Config: testAccServiceAccountPasswordConfigBasic(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_service_account_password.test",
						tfjsonpath.New("service_account_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_service_account_password.test",
						tfjsonpath.New("service_account_password_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_service_account_password.test",
						tfjsonpath.New("password"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_service_account_password.test",
						tfjsonpath.New("password_prefix"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_service_account_password.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Production API access password"),
					),
				},
			},
		},
	})
}

func TestAccResourceServiceAccountPassword_MultiplePasswords(t *testing.T) {
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
			// Create service account with multiple passwords
			{
				Config: testAccServiceAccountPasswordConfigMultiple(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_service_account_password.primary",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Primary password"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_service_account_password.rotation",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Rotation password for zero-downtime updates"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_service_account_password.primary",
						tfjsonpath.New("service_account_password_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_service_account_password.rotation",
						tfjsonpath.New("service_account_password_id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// testAccServiceAccountPasswordConfigBasic returns a basic service account password configuration
func testAccServiceAccountPasswordConfigBasic(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_service_account" "test" {
  username              = "tfaccsa_%[1]s"
  with_initial_password = false
  additional_role_ids   = []
}

resource "galaxy_service_account_password" "test" {
  service_account_id = galaxy_service_account.test.service_account_id
  description        = "Production API access password"
}
`, suffix)
}

// testAccServiceAccountPasswordConfigMultiple returns a configuration with multiple passwords
func testAccServiceAccountPasswordConfigMultiple(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_service_account" "test" {
  username              = "tfaccsa_%[1]s"
  with_initial_password = false
  additional_role_ids   = []
}

resource "galaxy_service_account_password" "primary" {
  service_account_id = galaxy_service_account.test.service_account_id
  description        = "Primary password"
}

resource "galaxy_service_account_password" "rotation" {
  service_account_id = galaxy_service_account.test.service_account_id
  description        = "Rotation password for zero-downtime updates"
}
`, suffix)
}

// TestAccResourceServiceAccountPassword_MinimalConfig tests creating a service account password with only required fields,
// omitting all optional parameters like description.
func TestAccResourceServiceAccountPassword_MinimalConfig(t *testing.T) {
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
				Config: testAccServiceAccountPasswordConfigMinimal(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_service_account_password.test",
						tfjsonpath.New("service_account_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_service_account_password.test",
						tfjsonpath.New("service_account_password_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_service_account_password.test",
						tfjsonpath.New("password"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// testAccServiceAccountPasswordConfigMinimal returns a minimal service account password configuration
func testAccServiceAccountPasswordConfigMinimal(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_service_account" "test" {
  username              = "tfaccpwmin_%[1]s"
  with_initial_password = false
  additional_role_ids   = []
}

resource "galaxy_service_account_password" "test" {
  service_account_id = galaxy_service_account.test.service_account_id
  # No optional fields - testing that omitting them doesn't cause API errors
  # Specifically: description
}
`, suffix)
}
