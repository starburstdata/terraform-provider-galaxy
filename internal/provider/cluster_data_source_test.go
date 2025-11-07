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

func TestAccDataSourceCluster_Basic(t *testing.T) {
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
			// Create cluster and read via data source
			{
				Config: testAccDataSourceClusterConfig(uniqueId),
				ConfigStateChecks: []statecheck.StateCheck{
					// Check resource
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("cluster-%s", uniqueId)),
					),
					// Check data source reads same values
					statecheck.ExpectKnownValue(
						"data.galaxy_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("cluster-%s", uniqueId)),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_cluster.test",
						tfjsonpath.New("cluster_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_cluster.test",
						tfjsonpath.New("min_workers"),
						knownvalue.Int64Exact(1),
					),
					statecheck.ExpectKnownValue(
						"data.galaxy_cluster.test",
						tfjsonpath.New("max_workers"),
						knownvalue.Int64Exact(2),
					),
				},
			},
		},
	})
}

func TestAccDataSourceClusters_List(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// List all clusters
			{
				Config: testAccDataSourceClustersConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Check that result is not null
					statecheck.ExpectKnownValue(
						"data.galaxy_clusters.all",
						tfjsonpath.New("result"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// testAccDataSourceClusterConfig returns a configuration for testing cluster data source
func testAccDataSourceClusterConfig(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_cluster" "test" {
  name                    = "cluster-%[1]s"
  cloud_region_id         = "aws-us-east1"
  min_workers             = 1
  max_workers             = 2
  idle_stop_minutes       = 15
  private_link_cluster    = false
  result_cache_enabled    = false
  warp_resiliency_enabled = false
  catalog_refs            = []
}

data "galaxy_cluster" "test" {
  cluster_id = galaxy_cluster.test.cluster_id
}
`, suffix)
}

// testAccDataSourceClustersConfig returns a configuration for testing clusters data source
func testAccDataSourceClustersConfig() string {
	return `
data "galaxy_clusters" "all" {}
`
}
