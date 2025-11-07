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

func TestAccDataSourceServiceAccount_Basic(t *testing.T) {
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
			// Create service account and read via data source
			{
				Config: testAccDataSourceServiceAccountConfig(uniqueId),
				ConfigStateChecks: []statecheck.StateCheck{
					// Check resource
					statecheck.ExpectKnownValue(
						"galaxy_service_account.test",
						tfjsonpath.New("service_account_id"),
						knownvalue.NotNull(),
					),
					// Check data source reads same values
					statecheck.ExpectKnownValue(
						"data.galaxy_service_account.test",
						tfjsonpath.New("service_account_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_service_account.test",
						tfjsonpath.New("additional_role_ids"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccDataSourceServiceAccounts_List(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// List all service accounts
			{
				Config: testAccDataSourceServiceAccountsConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Check that result is not null
					statecheck.ExpectKnownValue(
						"data.galaxy_service_accounts.all",
						tfjsonpath.New("result"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// testAccDataSourceServiceAccountConfig returns a configuration for testing service account data source
func testAccDataSourceServiceAccountConfig(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_service_account" "test" {
  username              = "sa_%[1]s"
  with_initial_password = false
  additional_role_ids   = []
}

data "galaxy_service_account" "test" {
  service_account_id = galaxy_service_account.test.service_account_id
}
`, suffix)
}

// testAccDataSourceServiceAccountsConfig returns a configuration for testing service accounts data source
func testAccDataSourceServiceAccountsConfig() string {
	return `
data "galaxy_service_accounts" "all" {}
`
}
