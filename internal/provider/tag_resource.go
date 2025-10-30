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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_tag"
)

var _ resource.Resource = (*tagResource)(nil)
var _ resource.ResourceWithConfigure = (*tagResource)(nil)

func NewTagResource() resource.Resource {
	return &tagResource{}
}

type tagResource struct {
	client *client.GalaxyClient
}

func (r *tagResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (r *tagResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_tag.TagResourceSchema(ctx)
}

func (r *tagResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_tag.TagModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating tag")
	response, err := r.client.CreateTag(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating tag",
			"Could not create tag: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created tag", map[string]interface{}{"id": plan.TagId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_tag.TagModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.TagId.ValueString()
	tflog.Debug(ctx, "Reading tag", map[string]interface{}{"id": id})
	response, err := r.client.GetTag(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Tag not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading tag",
			"Could not read tag "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_tag.TagModel
	var state resource_tag.TagModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.TagId.ValueString()
	request := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating tag", map[string]interface{}{"id": id})
	response, err := r.client.UpdateTag(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating tag",
			"Could not update tag "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated tag", map[string]interface{}{"id": plan.TagId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_tag.TagModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.TagId.ValueString()
	tflog.Debug(ctx, "Deleting tag", map[string]interface{}{"id": id})
	err := r.client.DeleteTag(ctx, id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting tag",
				"Could not delete tag "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted tag", map[string]interface{}{"id": id})
}

// Helper methods
func (r *tagResource) modelToCreateRequest(ctx context.Context, model *resource_tag.TagModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	if !model.Name.IsNull() {
		request["name"] = model.Name.ValueString()
	}
	if !model.Color.IsNull() {
		request["color"] = model.Color.ValueString()
	}

	// Optional fields
	if !model.Description.IsNull() {
		request["description"] = model.Description.ValueString()
	}

	return request
}

func (r *tagResource) modelToUpdateRequest(ctx context.Context, model *resource_tag.TagModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	// The API will return an error if we try to update it

	return request
}

func (r *tagResource) updateModelFromResponse(ctx context.Context, model *resource_tag.TagModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
	if id, ok := response["tagId"].(string); ok {
		model.TagId = types.StringValue(id)
	}

	if tagId, ok := response["tagId"].(string); ok {
		model.TagId = types.StringValue(tagId)
	}

	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}

	if color, ok := response["color"].(string); ok {
		model.Color = types.StringValue(color)
	}

	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else {
		model.Description = types.StringNull()
	}

}
