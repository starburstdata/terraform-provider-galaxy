// Copyright (c) Starburst Data, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
)

func TestAccDataSourcePrivateLink_Basic(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("%s not set, skipping acceptance test", resource.EnvTfAcc)
	}
	testAccPreCheck(t)
	privatelinkID, privatelinkName := testAccFirstPrivateLink(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourcePrivateLinkConfig(privatelinkID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.galaxy_privatelink.test",
						tfjsonpath.New("privatelink_id"),
						knownvalue.StringExact(privatelinkID),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_privatelink.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(privatelinkName),
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

// testAccFirstPrivateLink resolves an existing private link at test time so the test does not depend on a fixture ID
func testAccFirstPrivateLink(t *testing.T) (string, string) {
	t.Helper()
	apiClient := client.NewGalaxyClient(
		os.Getenv("GALAXY_DOMAIN"),
		os.Getenv("GALAXY_CLIENT_ID"),
		os.Getenv("GALAXY_CLIENT_SECRET"),
		"test",
	)
	response, err := apiClient.ListPrivatelinks(context.Background())
	if err != nil {
		t.Fatalf("failed to list private links: %s", err)
	}
	results, _ := response["result"].([]interface{})
	for _, item := range results {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := itemMap["privatelinkId"].(string)
		name, _ := itemMap["name"].(string)
		if id != "" && name != "" {
			return id, name
		}
	}
	t.Skip("no private links with an id and name exist in the test account, skipping")
	return "", ""
}

func testAccDataSourcePrivateLinkConfig(privatelinkID string) string {
	return fmt.Sprintf(`
data "galaxy_privatelink" "test" {
  privatelink_id = %q
}
`, privatelinkID)
}

func testAccDataSourcePrivateLinksConfig() string {
	return `
data "galaxy_privatelinks" "all" {}
`
}
