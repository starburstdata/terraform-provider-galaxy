// Copyright (c) Starburst Data, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func liveTableTypePropertiesRequiresReplace() planmodifier.Object {
	return objectplanmodifier.RequiresReplaceIf(
		func(ctx context.Context, req planmodifier.ObjectRequest, resp *objectplanmodifier.RequiresReplaceIfFuncResponse) {
			raw, err := liveTablePlanWithPriorComputed(ctx, req.Config, req.Plan, req.State)
			if err != nil {
				resp.Diagnostics.AddError("Error comparing live table plan", err.Error())
				return
			}
			value, _, err := tftypes.WalkAttributePath(raw, tftypes.NewAttributePath().WithAttributeName("type_properties"))
			if err != nil {
				resp.Diagnostics.AddError("Error comparing live table plan", err.Error())
				return
			}
			prior, err := req.StateValue.ToTerraformValue(ctx)
			if err != nil {
				resp.Diagnostics.AddError("Error comparing live table state", err.Error())
				return
			}
			resp.RequiresReplace = !keepPriorNestedColumnsJSON(value.(tftypes.Value), prior, "type_properties").Equal(prior)
		},
		"Substantive changes require replacement; equivalent nested-column JSON does not.",
		"Substantive changes require replacement; equivalent nested-column JSON does not.",
	)
}

func liveTablePlanWithPriorComputed(ctx context.Context, config tfsdk.Config, plan tfsdk.Plan, state tfsdk.State) (tftypes.Value, error) {
	return tftypes.Transform(plan.Raw, func(p *tftypes.AttributePath, value tftypes.Value) (tftypes.Value, error) {
		if value.IsKnown() {
			return value, nil
		}
		a, err := plan.Schema.AttributeAtTerraformPath(ctx, p)
		if err != nil || !a.IsComputed() {
			return value, nil
		}
		configured, _, err := tftypes.WalkAttributePath(config.Raw, p)
		if err != nil || !configured.(tftypes.Value).IsNull() {
			return value, nil
		}
		prior, _, err := tftypes.WalkAttributePath(state.Raw, p)
		if err != nil {
			return value, nil
		}
		return prior.(tftypes.Value), nil
	})
}

func (r *liveTableResource) updateEquivalentJSON(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	raw, err := liveTablePlanWithPriorComputed(ctx, req.Config, req.Plan, req.State)
	if err != nil {
		resp.Diagnostics.AddError("Error comparing live table plan", err.Error())
		return
	}
	if !keepPriorNestedColumnsJSON(raw, req.State.Raw, "").Equal(req.State.Raw) {
		resp.Diagnostics.AddError("Update not supported", "galaxy_live_table resources cannot be updated. Delete and recreate the resource.")
		return
	}
	// Only the representation changed; retain the configured text without touching the remote table.
	resp.State = tfsdk.State{Schema: req.Plan.Schema, Raw: raw}
}
