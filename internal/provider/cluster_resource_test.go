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
			// Import testing
			{
				ResourceName:                         "galaxy_cluster.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateIdFunc("galaxy_cluster.test", "cluster_id"),
				ImportStateVerifyIdentifierAttribute: "cluster_id",
				ImportStateVerifyIgnore: []string{
					// These fields are not returned by the Galaxy API
					"private_link_cluster",
					"result_cache_enabled",
					"warp_resiliency_enabled",
					"trino_uri", // computed, only available when cluster is ENABLED
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
			// Import testing
			{
				ResourceName:                         "galaxy_cluster.test_cache",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateIdFunc("galaxy_cluster.test_cache", "cluster_id"),
				ImportStateVerifyIdentifierAttribute: "cluster_id",
				ImportStateVerifyIgnore: []string{
					// These fields are not returned by the Galaxy API
					"private_link_cluster",
					"result_cache_enabled",
					"result_cache_default_visibility_seconds",
					"warp_resiliency_enabled",
					"trino_uri", // computed, only available when cluster is ENABLED
				},
			},
		},
	})
}

// TestAccResourceCluster_WarpSpeed verifies that a WarpSpeed cluster applies cleanly when
// warp_resiliency_enabled is set explicitly to true. The backend ignores the value (it hardcodes
// false server-side) and the response omits the field, so the Bool(true) assertion below passes
// because the Known plan value flows through to state untouched - NOT because the API honored it.
// The DeprecationMessage on the schema is what actually tells users this field is a no-op.
func TestAccResourceCluster_WarpSpeed(t *testing.T) {
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
				Config: testAccClusterConfigWarpSpeed(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test_warp",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("cluster-%s", suffix)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test_warp",
						tfjsonpath.New("warp_resiliency_enabled"),
						knownvalue.Bool(true),
					),
				},
			},
			// Import testing
			{
				ResourceName:                         "galaxy_cluster.test_warp",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateIdFunc("galaxy_cluster.test_warp", "cluster_id"),
				ImportStateVerifyIdentifierAttribute: "cluster_id",
				ImportStateVerifyIgnore: []string{
					// These fields are not returned by the Galaxy API
					"private_link_cluster",
					"result_cache_enabled",
					"processing_mode",
					"warp_resiliency_enabled",
					"trino_uri", // computed, only available when cluster is ENABLED
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

// testAccClusterConfigWarpSpeedOmitted returns a WarpSpeed cluster configuration
// that omits the deprecated warp_resiliency_enabled attribute entirely.
func testAccClusterConfigWarpSpeedOmitted(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_cluster" "test_warp_omitted" {
  name                 = "cluster-%[1]s"
  cloud_region_id      = "aws-us-east1"
  min_workers          = 1
  max_workers          = 2
  idle_stop_minutes    = 15
  private_link_cluster = false
  result_cache_enabled = false
  processing_mode      = "WarpSpeed"
  catalog_refs         = []
}
`, suffix)
}

// TestAccResourceCluster_WarpSpeedOmitted verifies that a WarpSpeed cluster can be
// created without specifying warp_resiliency_enabled and that apply completes without
// the "inconsistent result after apply" error reported in
// https://github.com/starburstdata/terraform-provider-galaxy/issues/72.
func TestAccResourceCluster_WarpSpeedOmitted(t *testing.T) {
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
				Config: testAccClusterConfigWarpSpeedOmitted(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test_warp_omitted",
						tfjsonpath.New("warp_resiliency_enabled"),
						knownvalue.Bool(false),
					),
				},
			},
			{
				ResourceName:                         "galaxy_cluster.test_warp_omitted",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateIdFunc("galaxy_cluster.test_warp_omitted", "cluster_id"),
				ImportStateVerifyIdentifierAttribute: "cluster_id",
				ImportStateVerifyIgnore: []string{
					"private_link_cluster",
					"result_cache_enabled",
					"processing_mode",
					"warp_resiliency_enabled",
					"trino_uri",
				},
			},
		},
	})
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
			// Import testing
			{
				ResourceName:                         "galaxy_cluster.test",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    importStateIdFunc("galaxy_cluster.test", "cluster_id"),
				ImportStateVerifyIdentifierAttribute: "cluster_id",
				ImportStateVerifyIgnore: []string{
					// These fields are not returned by the Galaxy API
					"private_link_cluster",
					"result_cache_enabled",
					"warp_resiliency_enabled",
					"trino_uri", // computed, only available when cluster is ENABLED
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

// TestAccResourceCluster_ResultCacheDisableAfterEnable verifies the gating logic in
// modelToCreateRequest: result_cache_default_visibility_seconds must not be sent when
// result_cache_enabled is false. The Galaxy API rejects such combinations with a 400.
// Creates a cluster with cache enabled and visibility set, then updates to cache disabled
// (omitting visibility from config). Apply must succeed.
func TestAccResourceCluster_ResultCacheDisableAfterEnable(t *testing.T) {
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
			{
				Config: testAccClusterConfigResultCacheDisabled(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test_cache",
						tfjsonpath.New("result_cache_enabled"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"galaxy_cluster.test_cache",
						tfjsonpath.New("result_cache_default_visibility_seconds"),
						knownvalue.Null(),
					),
				},
			},
		},
	})
}

// testAccClusterConfigResultCacheDisabled returns the same cluster as
// testAccClusterConfigResultCacheEnabled but with cache disabled and visibility omitted.
func testAccClusterConfigResultCacheDisabled(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_cluster" "test_cache" {
  name                 = "cluster-%[1]s-cache"
  cloud_region_id      = "aws-us-east1"
  min_workers          = 1
  max_workers          = 2
  idle_stop_minutes    = 15
  private_link_cluster = false
  result_cache_enabled = false
  catalog_refs         = []
}
`, suffix)
}
