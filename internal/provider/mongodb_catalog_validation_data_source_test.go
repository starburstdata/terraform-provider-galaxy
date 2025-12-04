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

func TestAccDataSourceMongoDBCatalogValidation_Basic(t *testing.T) {
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
				Config: testAccDataSourceMongoDBCatalogValidationConfig(suffix),
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
	return fmt.Sprintf(`
resource "galaxy_mongodb_catalog" "test" {
  name            = "mongoval%[1]s"
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
`, suffix)
}
