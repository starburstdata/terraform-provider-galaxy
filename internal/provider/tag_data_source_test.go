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

func TestAccDataSourceTag_Basic(t *testing.T) {
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
			// Create tag and read via data source
			{
				Config: testAccDataSourceTagConfig(uniqueId),
				ConfigStateChecks: []statecheck.StateCheck{
					// Check resource
					statecheck.ExpectKnownValue(
						"galaxy_tag.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("tag_%s", uniqueId)),
					),
					// Check data source reads same values
					statecheck.ExpectKnownValue(
						"data.galaxy_tag.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("tag_%s", uniqueId)),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_tag.test",
						tfjsonpath.New("tag_id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccDataSourceTags_List(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// List all tags
			{
				Config: testAccDataSourceTagsConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Check that result is not null
					statecheck.ExpectKnownValue(
						"data.galaxy_tags.all",
						tfjsonpath.New("result"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// testAccDataSourceTagConfig returns a configuration for testing tag data source
func testAccDataSourceTagConfig(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_tag" "test" {
  name        = "tag_%[1]s"
  description = "Tag for data source acceptance testing"
  color       = "#FF0000"
}

data "galaxy_tag" "test" {
  tag_id = galaxy_tag.test.tag_id
}
`, suffix)
}

// testAccDataSourceTagsConfig returns a configuration for testing tags data source
func testAccDataSourceTagsConfig() string {
	return `
data "galaxy_tags" "all" {}
`
}
