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

func TestAccDataSourceColumnMask_Basic(t *testing.T) {
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
			{
				Config: testAccDataSourceColumnMaskConfig(uniqueId),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_column_mask.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("column_mask_%s", uniqueId)),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_column_mask.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("column_mask_%s", uniqueId)),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_column_mask.test",
						tfjsonpath.New("column_mask_id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccDataSourceColumnMasks_List(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceColumnMasksConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.galaxy_column_masks.all",
						tfjsonpath.New("result"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccDataSourceColumnMaskConfig(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_column_mask" "test" {
  name             = "column_mask_%[1]s"
  description      = "Column mask for data source acceptance testing"
  expression       = "NULL"
  column_mask_type = "Varchar"
}

data "galaxy_column_mask" "test" {
  column_mask_id = galaxy_column_mask.test.column_mask_id
}
`, suffix)
}

func testAccDataSourceColumnMasksConfig() string {
	return `
data "galaxy_column_masks" "all" {}
`
}
