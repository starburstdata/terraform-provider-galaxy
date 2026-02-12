// Copyright (c) Starburst Data, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"galaxy": providerserver.NewProtocol6WithError(New("test")()),
}

// importStateIdFunc returns a function that extracts a single attribute from state
// to use as the import ID. Use this for resources with simple single-value IDs.
func importStateIdFunc(resourceName, attrName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		id := rs.Primary.Attributes[attrName]
		if id == "" {
			return "", fmt.Errorf("attribute %s not set", attrName)
		}
		return id, nil
	}
}

// testAccPreCheck validates that the required environment variables are set for
// acceptance tests. This function is called before every acceptance test to ensure
// the provider can be properly configured.
func testAccPreCheck(t *testing.T) {
	t.Helper()

	// Check for required environment variables
	if v := os.Getenv("GALAXY_CLIENT_ID"); v == "" {
		t.Fatal("GALAXY_CLIENT_ID must be set for acceptance tests")
	}

	if v := os.Getenv("GALAXY_CLIENT_SECRET"); v == "" {
		t.Fatal("GALAXY_CLIENT_SECRET must be set for acceptance tests")
	}

	if v := os.Getenv("GALAXY_DOMAIN"); v == "" {
		t.Fatal("GALAXY_DOMAIN must be set for acceptance tests")
	}
}
