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

func TestAccResourceTag_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTagConfigBasic("tfacc"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_tag.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("pii_tfacc"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_tag.test",
						tfjsonpath.New("tag_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_tag.test",
						tfjsonpath.New("color"),
						knownvalue.StringExact("#FF0000"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_tag.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Personally Identifiable Information"),
					),
				},
			},
			// Update and Read testing
			{
				Config: testAccTagConfigUpdate("tfacc"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_tag.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Updated: Personally Identifiable Information - requires special handling"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_tag.test",
						tfjsonpath.New("color"),
						knownvalue.StringExact("#CC0000"),
					),
				},
			},
		},
	})
}

func TestAccResourceTag_MultipleTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create multiple tags
			{
				Config: testAccTagConfigMultiple("tfacc_multi"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_tag.pii",
						tfjsonpath.New("name"),
						knownvalue.StringExact("pii_tfacc_multi"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_tag.public",
						tfjsonpath.New("name"),
						knownvalue.StringExact("public_tfacc_multi"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_tag.confidential",
						tfjsonpath.New("name"),
						knownvalue.StringExact("confidential_tfacc_multi"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_tag.pii",
						tfjsonpath.New("color"),
						knownvalue.StringExact("#FF0000"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_tag.public",
						tfjsonpath.New("color"),
						knownvalue.StringExact("#00FF00"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_tag.confidential",
						tfjsonpath.New("color"),
						knownvalue.StringExact("#FFA500"),
					),
				},
			},
		},
	})
}

// testAccTagConfigBasic returns a basic tag configuration
func testAccTagConfigBasic(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_tag" "test" {
  name        = "pii_%[1]s"
  description = "Personally Identifiable Information"
  color       = "#FF0000"
}
`, suffix)
}

// testAccTagConfigUpdate returns an updated tag configuration
func testAccTagConfigUpdate(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_tag" "test" {
  name        = "pii_%[1]s"
  description = "Updated: Personally Identifiable Information - requires special handling"
  color       = "#CC0000"
}
`, suffix)
}

// testAccTagConfigMultiple returns a configuration with multiple tags
func testAccTagConfigMultiple(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_tag" "pii" {
  name        = "pii_%[1]s"
  description = "Personally Identifiable Information"
  color       = "#FF0000"
}

resource "galaxy_tag" "public" {
  name        = "public_%[1]s"
  description = "Public data that can be shared freely"
  color       = "#00FF00"
}

resource "galaxy_tag" "confidential" {
  name        = "confidential_%[1]s"
  description = "Confidential business data"
  color       = "#FFA500"
}
`, suffix)
}
