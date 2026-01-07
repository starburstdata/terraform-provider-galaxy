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

func TestAccResourceCluster_Basic(t *testing.T) {
	// Generate a short random suffix to avoid conflicts with leftover resources
	uniqueId := id.UniqueId()
	// Take only the last 8 characters to keep the name short
	if len(uniqueId) > 8 {
		uniqueId = uniqueId[len(uniqueId)-8:]
	}
	suffix := uniqueId
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccClusterConfigBasic(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("cluster-%s", suffix)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test",
						tfjsonpath.New("cluster_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test",
						tfjsonpath.New("min_workers"),
						knownvalue.Int64Exact(1),
					),
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test",
						tfjsonpath.New("max_workers"),
						knownvalue.Int64Exact(2),
					),
				},
			},
			// Update and Read testing
			{
				Config: testAccClusterConfigUpdate(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("cluster-%s-updated", suffix)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test",
						tfjsonpath.New("max_workers"),
						knownvalue.Int64Exact(3),
					),
				},
			},
		},
	})
}

func TestAccResourceCluster_ResultCache(t *testing.T) {
	// Generate a short random suffix to avoid conflicts with leftover resources
	uniqueId := id.UniqueId()
	// Take only the last 8 characters to keep the name short
	if len(uniqueId) > 8 {
		uniqueId = uniqueId[len(uniqueId)-8:]
	}
	suffix := uniqueId
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create cluster with result cache enabled
			{
				Config: testAccClusterConfigResultCacheEnabled(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test_cache",
						tfjsonpath.New("result_cache_enabled"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test_cache",
						tfjsonpath.New("result_cache_default_visibility_seconds"),
						knownvalue.Int64Exact(3600),
					),
				},
			},
		},
	})
}

// TestAccResourceCluster_WarpSpeed is skipped due to provider bug where processing_mode
// becomes null after apply. See: https://github.com/starburstdata/terraform-provider-galaxy-generation/issues/XX
func TestAccResourceCluster_WarpSpeed(t *testing.T) {
	// Generate a short random suffix to avoid conflicts with leftover resources
	uniqueId := id.UniqueId()
	// Take only the last 8 characters to keep the name short
	if len(uniqueId) > 8 {
		uniqueId = uniqueId[len(uniqueId)-8:]
	}
	suffix := uniqueId
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create cluster with WarpSpeed processing mode
			{
				Config: testAccClusterConfigWarpSpeed(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test_warp",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("cluster-%s", suffix)),
					),
					// WarpResiliency should be automatically enabled when WarpSpeed is used
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test_warp",
						tfjsonpath.New("warp_resiliency_enabled"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

// testAccClusterConfigBasic returns a basic cluster configuration
func testAccClusterConfigBasic(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_cluster" "test" {
  name                  = "cluster-%[1]s"
  cloud_region_id       = "aws-us-east1"
  min_workers           = 1
  max_workers           = 2
  idle_stop_minutes     = 15
  private_link_cluster  = false
  result_cache_enabled  = false
  warp_resiliency_enabled = false
  catalog_refs          = []
}
`, suffix)
}

// testAccClusterConfigUpdate returns an updated cluster configuration
func testAccClusterConfigUpdate(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_cluster" "test" {
  name                  = "cluster-%[1]s-updated"
  cloud_region_id       = "aws-us-east1"
  min_workers           = 1
  max_workers           = 3
  idle_stop_minutes     = 30
  private_link_cluster  = false
  result_cache_enabled  = false
  warp_resiliency_enabled = false
  catalog_refs          = []
}
`, suffix)
}

// testAccClusterConfigResultCacheEnabled returns a cluster configuration with result cache enabled
func testAccClusterConfigResultCacheEnabled(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_cluster" "test_cache" {
  name                                    = "cluster-%[1]s-cache"
  cloud_region_id                         = "aws-us-east1"
  min_workers                             = 1
  max_workers                             = 2
  idle_stop_minutes                       = 15
  private_link_cluster                    = false
  result_cache_enabled                    = true
  result_cache_default_visibility_seconds = 3600
  warp_resiliency_enabled                 = false
  catalog_refs                            = []
}
`, suffix)
}

// testAccClusterConfigWarpSpeed returns a cluster configuration with WarpSpeed processing mode
func testAccClusterConfigWarpSpeed(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_cluster" "test_warp" {
  name                    = "cluster-%[1]s"
  cloud_region_id         = "aws-us-east1"
  min_workers             = 1
  max_workers             = 2
  idle_stop_minutes       = 15
  private_link_cluster    = false
  result_cache_enabled    = false
  processing_mode         = "WarpSpeed"
  warp_resiliency_enabled = true
  catalog_refs            = []
}
`, suffix)
}

// TestAccResourceCluster_MinimalConfig tests creating a cluster with only required fields,
// omitting all optional parameters like processing_mode, result_cache_default_visibility_seconds, etc.
func TestAccResourceCluster_MinimalConfig(t *testing.T) {
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
				Config: testAccClusterConfigMinimal(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("cluster-min-%s", suffix)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test",
						tfjsonpath.New("cluster_id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// testAccClusterConfigMinimal returns a minimal cluster configuration
func testAccClusterConfigMinimal(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_cluster" "test" {
  name                    = "cluster-min-%[1]s"
  cloud_region_id         = "aws-us-east1"
  min_workers             = 1
  max_workers             = 1
  idle_stop_minutes       = 15
  private_link_cluster    = false
  result_cache_enabled    = false
  warp_resiliency_enabled = false
  catalog_refs            = []
  # No optional fields - testing that omitting them doesn't cause API errors
  # Specifically: processing_mode, result_cache_default_visibility_seconds, notes, auto_stop_idle_cluster, etc.
}
`, suffix)
}
