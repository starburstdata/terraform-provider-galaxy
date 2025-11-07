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

func TestAccDataSourceMongoDBCatalogValidation_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceMongoDBCatalogValidationConfig("val"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.galaxy_mongodb_catalog_validation.test",
						tfjsonpath.New("validation_successful"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccDataSourceMongoDBCatalogValidationConfig(suffix string) string {
	return `
resource "galaxy_mongodb_catalog" "test" {
  name            = "mongoval"
  host            = "184.72.111.164:57017/admin"
  database        = "admin"
  username        = "galaxy-test"
  password        = "9bfn9v39chkmysgq"
  read_only       = false
  connection_type = "direct"
  regions         = []
  description     = "MongoDB catalog for validation testing"
}

data "galaxy_mongodb_catalog_validation" "test" {
  catalog_id = galaxy_mongodb_catalog.test.catalog_id
}
`
}
