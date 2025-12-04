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

func TestAccResourceColumnMask_Basic(t *testing.T) {
	// Generate a short random suffix to avoid conflicts with leftover resources
	uniqueId := id.UniqueId()
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
				Config: testAccColumnMaskConfigBasic(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_column_mask.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("ssnmask_%s", suffix)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_column_mask.test",
						tfjsonpath.New("column_mask_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"galaxy_column_mask.test",
						tfjsonpath.New("column_mask_type"),
						knownvalue.StringExact("Varchar"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_column_mask.test",
						tfjsonpath.New("expression"),
						knownvalue.StringExact("CONCAT('XXX-XX-', SUBSTRING(ssn, 8, 4))"),
					),
				},
			},
			// Update and Read testing
			{
				Config: testAccColumnMaskConfigUpdate(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_column_mask.test",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Updated SSN mask showing only last 4 digits"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_column_mask.test",
						tfjsonpath.New("expression"),
						knownvalue.StringExact("CONCAT('***-**-', SUBSTRING(ssn, 8, 4))"),
					),
				},
			},
		},
	})
}

func TestAccResourceColumnMask_EmailMask(t *testing.T) {
	// Generate a short random suffix to avoid conflicts with leftover resources
	uniqueId := id.UniqueId()
	if len(uniqueId) > 8 {
		uniqueId = uniqueId[len(uniqueId)-8:]
	}
	suffix := uniqueId
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing with email mask
			{
				Config: testAccColumnMaskConfigEmail(suffix),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"galaxy_column_mask.email",
						tfjsonpath.New("name"),
						knownvalue.StringExact(fmt.Sprintf("emailmask_%s", suffix)),
					),
					statecheck.ExpectKnownValue(
						"galaxy_column_mask.email",
						tfjsonpath.New("column_mask_type"),
						knownvalue.StringExact("Varchar"),
					),
					statecheck.ExpectKnownValue(
						"galaxy_column_mask.email",
						tfjsonpath.New("description"),
						knownvalue.StringExact("Show only email domain"),
					),
				},
			},
		},
	})
}

// testAccColumnMaskConfigBasic returns a basic column mask configuration
func testAccColumnMaskConfigBasic(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_column_mask" "test" {
  name             = "ssnmask_%[1]s"
  description      = "Mask SSN showing only last 4 digits"
  column_mask_type = "Varchar"
  expression       = "CONCAT('XXX-XX-', SUBSTRING(ssn, 8, 4))"
}
`, suffix)
}

// testAccColumnMaskConfigUpdate returns an updated column mask configuration
func testAccColumnMaskConfigUpdate(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_column_mask" "test" {
  name             = "ssnmask_%[1]s"
  description      = "Updated SSN mask showing only last 4 digits"
  column_mask_type = "Varchar"
  expression       = "CONCAT('***-**-', SUBSTRING(ssn, 8, 4))"
}
`, suffix)
}

// testAccColumnMaskConfigEmail returns a column mask configuration for email masking
func testAccColumnMaskConfigEmail(suffix string) string {
	return fmt.Sprintf(`
resource "galaxy_column_mask" "email" {
  name             = "emailmask_%[1]s"
  description      = "Show only email domain"
  column_mask_type = "Varchar"
  expression       = "CONCAT('***@', SPLIT_PART(email, '@', 2))"
}
`, suffix)
}
