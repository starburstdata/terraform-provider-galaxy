// Copyright (c) Starburst Data, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"

// testSuffix is a package-level timestamp-based suffix shared across all acceptance tests.
// It is unique per test run (YYYYMMDDHHMMSS) and unique per resource type because each
// test uses a different name prefix (e.g. "role-", "tag-", "s3cat-").
var testSuffix = id.UniqueId()[10:24]
