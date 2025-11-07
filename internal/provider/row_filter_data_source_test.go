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

func TestAccDataSourceRowFilter_Basic(t *testing.T) {
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
				Config: testAccDataSourceRowFilterConfig(uniqueId),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_row_filter.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("row_filter_%s", uniqueId)),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_row_filter.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("row_filter_%s", uniqueId)),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_row_filter.test",
						tfjsonpath.New("row_filter_id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccDataSourceRowFilters_List(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceRowFiltersConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.galaxy_row_filters.all",
						tfjsonpath.New("result"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccDataSourceRowFilterConfig(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_row_filter" "test" {
  name        = "row_filter_%[1]s"
  description = "Row filter for data source acceptance testing"
  expression  = "true"
}

data "galaxy_row_filter" "test" {
  row_filter_id = galaxy_row_filter.test.row_filter_id
}
`, suffix)
}

func testAccDataSourceRowFiltersConfig() string {
	return `
data "galaxy_row_filters" "all" {}
`
}
