// Copyright Starburst Data, Inc. All rights reserved.
//
// The source code is the proprietary and confidential information of Starburst Data, Inc. and
// may be used only for reference purposes in connection with the Terraform Registry. All rights,
// title, interest and ownership of the code and any derivatives, updates, upgrades, enhancements
// and modifications thereof remain with Starburst Data, Inc. You are not permitted to distribute,
// disclose, sell, lease, transfer, assign, modify, create derivative works of, or sublicense the
// code, or use the code to create or develop any products or services.

package provider

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// reorderToMatchPlanBy reorders actual to match the relative order of planned according to
// key(). Items whose key isn't present in planned keep their original relative order and sort
// after all planned-matching items. Required, non-computed list attributes must have their
// applied state exactly match the plan, but the Galaxy API does not preserve submission order -
// this avoids "Provider produced inconsistent result after apply" on every apply.
func reorderToMatchPlanBy[T any](planned, actual []T, key func(T) string) []T {
	if len(planned) == 0 || len(actual) == 0 {
		return slices.Clone(actual)
	}

	rank := make(map[string]int, len(planned))
	for i, p := range planned {
		k := key(p)
		if _, exists := rank[k]; !exists {
			rank[k] = i
		}
	}

	result := slices.Clone(actual)
	slices.SortStableFunc(result, func(a, b T) int {
		ra, aOk := rank[key(a)]
		rb, bOk := rank[key(b)]
		switch {
		case aOk && bOk:
			return ra - rb
		case aOk:
			return -1
		case bOk:
			return 1
		default:
			return 0
		}
	})
	return result
}

// reorderToMatchPlan reorders actual string IDs to match the order of planned.
func reorderToMatchPlan(planned, actual []string) []string {
	return reorderToMatchPlanBy(planned, actual, func(s string) string { return s })
}

// stringListElements extracts a []string from a types.List, returning nil for null/unknown lists.
// Any diagnostics from ElementsAs are returned to the caller so extraction failures can be
// surfaced rather than silently masked.
func stringListElements(ctx context.Context, list types.List) ([]string, diag.Diagnostics) {
	if list.IsNull() || list.IsUnknown() {
		return nil, nil
	}
	elements := make([]types.String, 0, len(list.Elements()))
	if d := list.ElementsAs(ctx, &elements, false); d.HasError() {
		return nil, d
	}
	result := make([]string, 0, len(elements))
	for _, elem := range elements {
		if !elem.IsNull() && !elem.IsUnknown() {
			result = append(result, elem.ValueString())
		}
	}
	return result, nil
}
