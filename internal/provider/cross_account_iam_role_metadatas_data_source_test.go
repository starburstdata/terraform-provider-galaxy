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

func TestAccDataSourceCrossAccountIAMRoleMetadatas_List(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceCrossAccountIAMRoleMetadatasConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.galaxy_cross_account_iam_role_metadatas.all",
						tfjsonpath.New("external_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_cross_account_iam_role_metadatas.all",
						tfjsonpath.New("starburst_aws_account_id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func testAccDataSourceCrossAccountIAMRoleMetadatasConfig() string {
	return `
data "galaxy_cross_account_iam_role_metadatas" "all" {}
`
}
