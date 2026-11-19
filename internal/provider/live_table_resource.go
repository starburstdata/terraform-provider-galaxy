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
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"unicode"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_live_table"
)

var _ resource.Resource = (*liveTableResource)(nil)
var _ resource.ResourceWithConfigure = (*liveTableResource)(nil)
var _ resource.ResourceWithImportState = (*liveTableResource)(nil)

func NewLiveTableResource() resource.Resource {
	return &liveTableResource{}
}

type liveTableResource struct {
	client *client.GalaxyClient
}

func (r *liveTableResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_live_table"
}

func (r *liveTableResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_live_table.LiveTableResourceSchema(ctx)

	if attr, ok := s.Attributes["live_table_id"].(schema.StringAttribute); ok {
		attr.PlanModifiers = []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
		s.Attributes["live_table_id"] = attr
	}

	resp.Schema = s
}

func (r *liveTableResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.GalaxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.GalaxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *liveTableResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Convert the full tftypes-shaped plan into a camelCase map compatible with the Galaxy API.
	rawMap, diags := tftypesValueToMap(ctx, req.Plan.Raw)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	request, ok := rawMap.(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("Internal error", "plan did not decode to an object")
		return
	}

	// Strip computed-only fields the API doesn't accept on create
	for _, k := range []string{
		"catalogName", "cloudRegionId", "errorNotificationCount", "id",
		"liveTableId", "liveTableState", "stopMessage", "stopType",
		"warnNotificationCount",
	} {
		delete(request, k)
	}

	tflog.Debug(ctx, "Creating live_table")
	response, err := r.client.CreateLiveTable(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating live_table",
			"Could not create live_table: "+err.Error(),
		)
		return
	}

	id, _ := response["liveTableId"].(string)
	getResp, err := r.client.GetLiveTable(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading live_table after create",
			"Could not read live_table "+id+": "+err.Error(),
		)
		return
	}

	r.writeResponseToState(ctx, getResp, &resp.State, &resp.Diagnostics)
}

func (r *liveTableResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_live_table.LiveTableModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.LiveTableId.ValueString()
	response, err := r.client.GetLiveTable(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading live_table",
			"Could not read live_table "+id+": "+err.Error(),
		)
		return
	}

	r.writeResponseToState(ctx, response, &resp.State, &resp.Diagnostics)
}

// Update is not supported by the Galaxy API (no PATCH endpoint).
func (r *liveTableResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Update not supported",
		"galaxy_live_table resources cannot be updated. Delete and recreate the resource.",
	)
}

func (r *liveTableResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_live_table.LiveTableModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.LiveTableId.ValueString()
	if err := r.client.DeleteLiveTable(ctx, id); err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting live_table",
				"Could not delete live_table "+id+": "+err.Error(),
			)
			return
		}
	}
}

func (r *liveTableResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("live_table_id"), req, resp)
}

// writeResponseToState takes a Galaxy API response (camelCase JSON) and writes it into state attribute-by-attribute.
func (r *liveTableResource) writeResponseToState(ctx context.Context, response map[string]interface{}, state *tfsdk.State, diags *diag.Diagnostics) {
	s := resource_live_table.LiveTableResourceSchema(ctx)
	if id, ok := response["liveTableId"].(string); ok {
		response["id"] = id
	}

	for attrName, attrSchema := range s.Attributes {
		attrFwType := attrSchema.GetType()
		attrType := attrFwType.TerraformType(ctx)
		key := snakeToCamel(attrName)
		if attrName == "id" {
			key = "id"
		}
		raw, has := response[key]
		if !has {
			raw = nil
		}
		tfval, convDiags := mapToTftypesValue(ctx, raw, attrType)
		diags.Append(convDiags...)
		if diags.HasError() {
			return
		}
		attrVal, err := attrFwType.ValueFromTerraform(ctx, tfval)
		if err != nil {
			diags.AddError("conversion error", fmt.Sprintf("%s: %v", attrName, err))
			return
		}
		diags.Append(state.SetAttribute(ctx, path.Root(attrName), attrVal)...)
		if diags.HasError() {
			return
		}
	}
}

// tftypesValueToMap walks a tftypes.Value and returns its Go representation, converting snake_case attribute keys to camelCase.
func tftypesValueToMap(ctx context.Context, v tftypes.Value) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	if v.IsNull() || !v.IsKnown() {
		return nil, diags
	}
	ty := v.Type()
	switch {
	case ty.Is(tftypes.String):
		var s string
		if err := v.As(&s); err != nil {
			diags.AddError("decode string", err.Error())
			return nil, diags
		}
		return s, diags
	case ty.Is(tftypes.Bool):
		var b bool
		if err := v.As(&b); err != nil {
			diags.AddError("decode bool", err.Error())
			return nil, diags
		}
		return b, diags
	case ty.Is(tftypes.Number):
		var n *big.Float
		if err := v.As(&n); err != nil {
			diags.AddError("decode number", err.Error())
			return nil, diags
		}
		if n == nil {
			return nil, diags
		}
		if n.IsInt() {
			i, _ := n.Int64()
			return i, diags
		}
		f, _ := n.Float64()
		return f, diags
	}

	if ty.Is(tftypes.List{}) || ty.Is(tftypes.Set{}) || ty.Is(tftypes.Tuple{}) {
		var items []tftypes.Value
		if err := v.As(&items); err != nil {
			diags.AddError("decode list", err.Error())
			return nil, diags
		}
		out := make([]interface{}, 0, len(items))
		for _, item := range items {
			x, d := tftypesValueToMap(ctx, item)
			diags.Append(d...)
			out = append(out, x)
		}
		return out, diags
	}

	// Object or Map
	if ty.Is(tftypes.Map{}) {
		var items map[string]tftypes.Value
		if err := v.As(&items); err != nil {
			diags.AddError("decode map", err.Error())
			return nil, diags
		}
		out := make(map[string]interface{}, len(items))
		for k, item := range items {
			x, d := tftypesValueToMap(ctx, item)
			diags.Append(d...)
			out[k] = x
		}
		return out, diags
	}

	// Object
	var items map[string]tftypes.Value
	if err := v.As(&items); err != nil {
		diags.AddError("decode object", err.Error())
		return nil, diags
	}
	out := make(map[string]interface{}, len(items))
	for k, item := range items {
		x, d := tftypesValueToMap(ctx, item)
		diags.Append(d...)
		if x == nil {
			continue
		}
		out[snakeToCamel(k)] = x
	}
	return out, diags
}

// mapToTftypesValue is the inverse of tftypesValueToMap: given a camelCase Galaxy API response and a target tftypes.Type (derived
// from the resource schema), produce a tftypes.Value shaped to that type. Missing keys yield null values of the expected type.
func mapToTftypesValue(ctx context.Context, m interface{}, ty tftypes.Type) (tftypes.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if m == nil {
		return tftypes.NewValue(ty, nil), diags
	}

	switch {
	case ty.Is(tftypes.String):
		if s, ok := m.(string); ok {
			return tftypes.NewValue(ty, s), diags
		}
		return tftypes.NewValue(ty, nil), diags
	case ty.Is(tftypes.Bool):
		if b, ok := m.(bool); ok {
			return tftypes.NewValue(ty, b), diags
		}
		return tftypes.NewValue(ty, nil), diags
	case ty.Is(tftypes.Number):
		switch n := m.(type) {
		case float64:
			return tftypes.NewValue(ty, big.NewFloat(n)), diags
		case int64:
			return tftypes.NewValue(ty, big.NewFloat(float64(n))), diags
		case int:
			return tftypes.NewValue(ty, big.NewFloat(float64(n))), diags
		case json.Number:
			f, err := n.Float64()
			if err != nil {
				diags.AddError("decode json.Number", err.Error())
				return tftypes.NewValue(ty, nil), diags
			}
			return tftypes.NewValue(ty, big.NewFloat(f)), diags
		}
		return tftypes.NewValue(ty, nil), diags
	}

	if lt, ok := ty.(tftypes.List); ok {
		arr, ok := m.([]interface{})
		if !ok {
			return tftypes.NewValue(ty, nil), diags
		}
		elems := make([]tftypes.Value, 0, len(arr))
		for _, item := range arr {
			v, d := mapToTftypesValue(ctx, item, lt.ElementType)
			diags.Append(d...)
			elems = append(elems, v)
		}
		return tftypes.NewValue(ty, elems), diags
	}
	if lt, ok := ty.(tftypes.Set); ok {
		arr, ok := m.([]interface{})
		if !ok {
			return tftypes.NewValue(ty, nil), diags
		}
		elems := make([]tftypes.Value, 0, len(arr))
		for _, item := range arr {
			v, d := mapToTftypesValue(ctx, item, lt.ElementType)
			diags.Append(d...)
			elems = append(elems, v)
		}
		return tftypes.NewValue(ty, elems), diags
	}
	if mt, ok := ty.(tftypes.Map); ok {
		obj, ok := m.(map[string]interface{})
		if !ok {
			return tftypes.NewValue(ty, nil), diags
		}
		elems := make(map[string]tftypes.Value, len(obj))
		for k, item := range obj {
			v, d := mapToTftypesValue(ctx, item, mt.ElementType)
			diags.Append(d...)
			elems[k] = v
		}
		return tftypes.NewValue(ty, elems), diags
	}
	if ot, ok := ty.(tftypes.Object); ok {
		obj, _ := m.(map[string]interface{})
		attrs := make(map[string]tftypes.Value, len(ot.AttributeTypes))
		for attrName, attrType := range ot.AttributeTypes {
			var value interface{}
			if obj != nil {
				// Try camelCase key first; "id" maps to "id"
				if attrName == "id" {
					value = obj["id"]
					if value == nil {
						value = obj["liveTableId"]
					}
				} else {
					value = obj[snakeToCamel(attrName)]
				}
			}
			v, d := mapToTftypesValue(ctx, value, attrType)
			diags.Append(d...)
			attrs[attrName] = v
		}
		return tftypes.NewValue(ty, attrs), diags
	}
	if tt, ok := ty.(tftypes.Tuple); ok {
		arr, _ := m.([]interface{})
		elems := make([]tftypes.Value, 0, len(tt.ElementTypes))
		for i, et := range tt.ElementTypes {
			var item interface{}
			if i < len(arr) {
				item = arr[i]
			}
			v, d := mapToTftypesValue(ctx, item, et)
			diags.Append(d...)
			elems = append(elems, v)
		}
		return tftypes.NewValue(ty, elems), diags
	}
	return tftypes.NewValue(ty, nil), diags
}

func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		return s
	}
	var b strings.Builder
	b.WriteString(parts[0])
	for _, p := range parts[1:] {
		if p == "" {
			continue
		}
		rs := []rune(p)
		rs[0] = unicode.ToUpper(rs[0])
		b.WriteString(string(rs))
	}
	return b.String()
}
