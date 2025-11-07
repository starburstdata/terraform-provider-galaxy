// Copyright (c) Starburst Data, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccResourceMongoDBCatalog_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMongoDBCatalogConfigBasic("tfacc"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_mongodb_catalog.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("mongocattfacc"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_mongodb_catalog.test",
						tfjsonpath.New("catalog_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_mongodb_catalog.test",
						tfjsonpath.New("connection_type"),
						knownvalue.StringExact("direct"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_mongodb_catalog.test",
						tfjsonpath.New("read_only"),
						knownvalue.Bool(false),
					),
				},
			},
			// Update and Read testing
			{
				Config: testAccMongoDBCatalogConfigUpdated("tfacc"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_mongodb_catalog.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Updated MongoDB catalog for acceptance testing"),
					),
				},
			},
		},
	})
}

// testAccMongoDBCatalogConfigBasic returns a basic MongoDB catalog configuration
func testAccMongoDBCatalogConfigBasic(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_mongodb_catalog" "test" {
  name            = "mongocat%[1]s"
  host            = "184.72.111.164:57017/admin"
  database        = "admin"
  username        = "galaxy-test"
  password        = "9bfn9v39chkmysgq"
  read_only       = false
  connection_type = "direct"
  regions         = []
  description     = "MongoDB catalog for acceptance testing"
}
`, suffix)
}

// testAccMongoDBCatalogConfigUpdated returns an updated MongoDB catalog configuration
func testAccMongoDBCatalogConfigUpdated(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_mongodb_catalog" "test" {
  name            = "mongocat%[1]s"
  host            = "184.72.111.164:57017/admin"
  database        = "admin"
  username        = "galaxy-test"
  password        = "9bfn9v39chkmysgq"
  read_only       = false
  connection_type = "direct"
  regions         = []
  description     = "Updated MongoDB catalog for acceptance testing"
}
`, suffix)
}
