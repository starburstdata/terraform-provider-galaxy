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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_row_filter"
)

var _ resource.Resource = (*row_filterResource)(nil)
var _ resource.ResourceWithConfigure = (*row_filterResource)(nil)

func NewRowFilterResource() resource.Resource {
	return &row_filterResource{}
}

type row_filterResource struct {
	client *client.GalaxyClient
}

func (r *row_filterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_row_filter"
}

func (r *row_filterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_row_filter.RowFilterResourceSchema(ctx)
}

func (r *row_filterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *row_filterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_row_filter.RowFilterModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating row_filter")
	response, err := r.client.CreateRowFilter(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating row_filter",
			"Could not create row_filter: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created row_filter", map[string]interface{}{"id": plan.RowFilterId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *row_filterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_row_filter.RowFilterModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.RowFilterId.ValueString()
	tflog.Debug(ctx, "Reading row_filter", map[string]interface{}{"id": id})
	response, err := r.client.GetRowFilter(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "RowFilter not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading row_filter",
			"Could not read row_filter "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *row_filterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_row_filter.RowFilterModel
	var state resource_row_filter.RowFilterModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.RowFilterId.ValueString()
	request := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating row_filter", map[string]interface{}{"id": id})
	response, err := r.client.UpdateRowFilter(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating row_filter",
			"Could not update row_filter "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated row_filter", map[string]interface{}{"id": plan.RowFilterId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *row_filterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_row_filter.RowFilterModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.RowFilterId.ValueString()
	tflog.Debug(ctx, "Deleting row_filter", map[string]interface{}{"id": id})
	err := r.client.DeleteRowFilter(ctx, id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting row_filter",
				"Could not delete row_filter "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted row_filter", map[string]interface{}{"id": id})
}

// Helper methods
func (r *row_filterResource) modelToCreateRequest(ctx context.Context, model *resource_row_filter.RowFilterModel, diags *diag.Diagnostics) map[string]interface{} {
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

	return request
}

func (r *row_filterResource) modelToUpdateRequest(ctx context.Context, model *resource_row_filter.RowFilterModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *row_filterResource) updateModelFromResponse(ctx context.Context, model *resource_row_filter.RowFilterModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
	if id, ok := response["rowFilterId"].(string); ok {
		model.RowFilterId = types.StringValue(id)
	}

	if rowFilterId, ok := response["rowFilterId"].(string); ok {
		model.RowFilterId = types.StringValue(rowFilterId)
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

	if created, ok := response["created"].(string); ok {
		model.Created = types.StringValue(created)
	}

	if modified, ok := response["modified"].(string); ok {
		model.Modified = types.StringValue(modified)
	}

}
