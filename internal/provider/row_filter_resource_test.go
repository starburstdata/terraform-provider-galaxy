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

func TestAccResourceRowFilter_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRowFilterConfigBasic("tfacc"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_row_filter.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("regionfilter_tfacc"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_row_filter.test",
						tfjsonpath.New("row_filter_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_row_filter.test",
						tfjsonpath.New("expression"),
						knownvalue.StringExact("region = 'US-EAST'"),
					),
				},
			},
			// Update and Read testing
			{
				Config: testAccRowFilterConfigUpdate("tfacc"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_row_filter.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Updated: Restrict access to US regions only"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_row_filter.test",
						tfjsonpath.New("expression"),
						knownvalue.StringExact("region IN ('US-EAST', 'US-WEST')"),
					),
				},
			},
		},
	})
}

func TestAccResourceRowFilter_TimeFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with time-based filter
			{
				Config: testAccRowFilterConfigTime("tfacc_time"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_row_filter.time",
						tfjsonpath.New("name"),
						knownvalue.StringExact("timefilter_tfacc_time"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_row_filter.time",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Only show data from last 30 days"),
					),
				},
			},
		},
	})
}

// testAccRowFilterConfigBasic returns a basic row filter configuration
func testAccRowFilterConfigBasic(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_row_filter" "test" {
  name        = "regionfilter_%[1]s"
  description = "Restrict access to US East region data only"
  expression  = "region = 'US-EAST'"
}
`, suffix)
}

// testAccRowFilterConfigUpdate returns an updated row filter configuration
func testAccRowFilterConfigUpdate(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_row_filter" "test" {
  name        = "regionfilter_%[1]s"
  description = "Updated: Restrict access to US regions only"
  expression  = "region IN ('US-EAST', 'US-WEST')"
}
`, suffix)
}

// testAccRowFilterConfigTime returns a row filter configuration with time-based filtering
func testAccRowFilterConfigTime(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_row_filter" "time" {
  name        = "timefilter_%[1]s"
  description = "Only show data from last 30 days"
  expression  = "event_date >= CURRENT_DATE - INTERVAL '30' DAY"
}
`, suffix)
}
