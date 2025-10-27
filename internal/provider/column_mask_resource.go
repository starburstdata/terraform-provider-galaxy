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
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_column_mask"
)

var _ resource.Resource = (*column_maskResource)(nil)
var _ resource.ResourceWithConfigure = (*column_maskResource)(nil)

func NewColumnMaskResource() resource.Resource {
	return &column_maskResource{}
}

type column_maskResource struct {
	client *client.GalaxyClient
}

func (r *column_maskResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_column_mask"
}

func (r *column_maskResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_column_mask.ColumnMaskResourceSchema(ctx)
}

func (r *column_maskResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.GalaxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.GalaxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *column_maskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_column_mask.ColumnMaskModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating column_mask")
	response, err := r.client.CreateColumnMask(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating column_mask",
			"Could not create column_mask: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created column_mask", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *column_maskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_column_mask.ColumnMaskModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Reading column_mask", map[string]interface{}{"id": id})
	response, err := r.client.GetColumnMask(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "ColumnMask not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading column_mask",
			"Could not read column_mask "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *column_maskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_column_mask.ColumnMaskModel
	var state resource_column_mask.ColumnMaskModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	request := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating column_mask", map[string]interface{}{"id": id})
	response, err := r.client.UpdateColumnMask(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating column_mask",
			"Could not update column_mask "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated column_mask", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *column_maskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_column_mask.ColumnMaskModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Deleting column_mask", map[string]interface{}{"id": id})
	err := r.client.DeleteColumnMask(ctx, id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting column_mask",
				"Could not delete column_mask "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted column_mask", map[string]interface{}{"id": id})
}

// Helper methods
func (r *column_maskResource) modelToCreateRequest(ctx context.Context, model *resource_column_mask.ColumnMaskModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	if !model.Name.IsNull() {
		request["name"] = model.Name.ValueString()
	}
	if !model.Description.IsNull() {
		request["description"] = model.Description.ValueString()
	}
	if !model.Expression.IsNull() {
		request["expression"] = model.Expression.ValueString()
	}
	if !model.ColumnMaskType.IsNull() {
		request["columnMaskType"] = model.ColumnMaskType.ValueString()
	}

	return request
}

func (r *column_maskResource) modelToUpdateRequest(ctx context.Context, model *resource_column_mask.ColumnMaskModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *column_maskResource) updateModelFromResponse(ctx context.Context, model *resource_column_mask.ColumnMaskModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	} else if id, ok := response["columnMaskId"].(string); ok {
		model.Id = types.StringValue(id)
	}

	if columnMaskId, ok := response["columnMaskId"].(string); ok {
		model.ColumnMaskId = types.StringValue(columnMaskId)
	}

	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}

	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	}

	if expression, ok := response["expression"].(string); ok {
		model.Expression = types.StringValue(expression)
	}

	if columnMaskType, ok := response["columnMaskType"].(string); ok {
		model.ColumnMaskType = types.StringValue(columnMaskType)
	}

	if created, ok := response["created"].(string); ok {
		model.Created = types.StringValue(created)
	}

	if modified, ok := response["modified"].(string); ok {
		model.Modified = types.StringValue(modified)
	}

}
