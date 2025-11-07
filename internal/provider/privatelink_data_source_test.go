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

func TestAccDataSourcePrivateLink_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePrivateLinkConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.galaxy_privatelink.test",
						tfjsonpath.New("privatelink_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_privatelink.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("brady-pl-test"),
					),
				},
			},
		},
	})
}

func TestAccDataSourcePrivateLinks_List(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePrivateLinksConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.galaxy_privatelinks.all",
						tfjsonpath.New("result"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccDataSourcePrivateLinkConfig() string {
	return `
data "galaxy_privatelink" "test" {
  privatelink_id = "pvtlnk-5554740448"
}
`
}

func testAccDataSourcePrivateLinksConfig() string {
	return `
data "galaxy_privatelinks" "all" {}
`
}
