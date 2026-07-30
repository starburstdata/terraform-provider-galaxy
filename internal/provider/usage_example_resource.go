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

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_usage_example"
)

var _ resource.Resource = (*usageExampleResource)(nil)
var _ resource.ResourceWithConfigure = (*usageExampleResource)(nil)
var _ resource.ResourceWithImportState = (*usageExampleResource)(nil)

func NewUsageExampleResource() resource.Resource {
	return &usageExampleResource{}
}

type usageExampleResource struct {
	client *client.GalaxyClient
}

func (r *usageExampleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_usage_example"
}

func (r *usageExampleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_usage_example.UsageExampleResourceSchema(ctx)

	// data_product_id is required in path for all API calls, override to Required with RequiresReplace
	if attr, ok := s.Attributes["data_product_id"].(schema.StringAttribute); ok {
		attr.Required = true
		attr.Computed = false
		attr.PlanModifiers = []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		}
		s.Attributes["data_product_id"] = attr
	}

	// usage_example_id is assigned at creation and never changes
	if attr, ok := s.Attributes["usage_example_id"].(schema.StringAttribute); ok {
		attr.PlanModifiers = []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		}
		s.Attributes["usage_example_id"] = attr
	}

	// id is assigned at creation and never changes
	if attr, ok := s.Attributes["id"].(schema.StringAttribute); ok {
		attr.PlanModifiers = []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		}
		s.Attributes["id"] = attr
	}

	// Optional+Computed fields: use state for unknown.
	// modified_on is intentionally excluded because the API updates it on every PATCH.
	for _, name := range []string{"description", "suggested_prompt", "created_on"} {
		if attr, ok := s.Attributes[name].(schema.StringAttribute); ok {
			attr.PlanModifiers = []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			}
			s.Attributes[name] = attr
		}
	}

	resp.Schema = s
}

func (r *usageExampleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *usageExampleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_usage_example.UsageExampleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataProductID := plan.DataProductId.ValueString()

	request := map[string]interface{}{
		"name": plan.Name.ValueString(),
		"code": plan.Code.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() && plan.Description.ValueString() != "" {
		request["description"] = plan.Description.ValueString()
	}
	if !plan.SuggestedPrompt.IsNull() && !plan.SuggestedPrompt.IsUnknown() && plan.SuggestedPrompt.ValueString() != "" {
		request["suggestedPrompt"] = plan.SuggestedPrompt.ValueString()
	}

	tflog.Debug(ctx, "Creating usage example", map[string]interface{}{"data_product_id": dataProductID})
	response, err := r.client.CreateUsageExample(ctx, dataProductID, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating usage example",
			"Could not create usage example: "+err.Error(),
		)
		return
	}

	// Set defaults for computed fields that were unknown before creation
	if plan.Description.IsUnknown() {
		plan.Description = types.StringNull()
	}
	if plan.SuggestedPrompt.IsUnknown() {
		plan.SuggestedPrompt = types.StringNull()
	}
	if plan.CreatedOn.IsUnknown() {
		plan.CreatedOn = types.StringNull()
	}
	if plan.ModifiedOn.IsUnknown() {
		plan.ModifiedOn = types.StringNull()
	}

	r.updateModelFromResponse(&plan, response)

	tflog.Debug(ctx, "Created usage example", map[string]interface{}{"id": plan.UsageExampleId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *usageExampleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_usage_example.UsageExampleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataProductID := state.DataProductId.ValueString()
	usageExampleID := state.UsageExampleId.ValueString()

	tflog.Debug(ctx, "Reading usage example", map[string]interface{}{"id": usageExampleID})
	response, err := r.client.GetUsageExample(ctx, dataProductID, usageExampleID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Usage example not found, removing from state", map[string]interface{}{"id": usageExampleID})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading usage example",
			"Could not read usage example "+usageExampleID+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(&state, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *usageExampleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_usage_example.UsageExampleModel
	var state resource_usage_example.UsageExampleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataProductID := state.DataProductId.ValueString()
	usageExampleID := state.UsageExampleId.ValueString()

	request := map[string]interface{}{
		"name": plan.Name.ValueString(),
		"code": plan.Code.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() && plan.Description.ValueString() != "" {
		request["description"] = plan.Description.ValueString()
	}
	if !plan.SuggestedPrompt.IsNull() && !plan.SuggestedPrompt.IsUnknown() && plan.SuggestedPrompt.ValueString() != "" {
		request["suggestedPrompt"] = plan.SuggestedPrompt.ValueString()
	}

	tflog.Debug(ctx, "Updating usage example", map[string]interface{}{"id": usageExampleID})
	response, err := r.client.UpdateUsageExample(ctx, dataProductID, usageExampleID, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating usage example",
			"Could not update usage example "+usageExampleID+": "+err.Error(),
		)
		return
	}

	plan.DataProductId = state.DataProductId
	plan.UsageExampleId = state.UsageExampleId
	r.updateModelFromResponse(&plan, response)

	tflog.Debug(ctx, "Updated usage example", map[string]interface{}{"id": plan.UsageExampleId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *usageExampleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_usage_example.UsageExampleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataProductID := state.DataProductId.ValueString()
	usageExampleID := state.UsageExampleId.ValueString()

	tflog.Debug(ctx, "Deleting usage example", map[string]interface{}{"id": usageExampleID})
	err := r.client.DeleteUsageExample(ctx, dataProductID, usageExampleID)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting usage example",
				"Could not delete usage example "+usageExampleID+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted usage example", map[string]interface{}{"id": usageExampleID})
}

func (r *usageExampleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *usageExampleResource) updateModelFromResponse(model *resource_usage_example.UsageExampleModel, response map[string]interface{}) {
	if usageExampleId, ok := response["usageExampleId"].(string); ok {
		model.UsageExampleId = types.StringValue(usageExampleId)
		model.Id = types.StringValue(usageExampleId)
	}
	if dataProductId, ok := response["dataProductId"].(string); ok {
		model.DataProductId = types.StringValue(dataProductId)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if code, ok := response["code"].(string); ok {
		model.Code = types.StringValue(code)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else {
		model.Description = types.StringNull()
	}
	if suggestedPrompt, ok := response["suggestedPrompt"].(string); ok {
		model.SuggestedPrompt = types.StringValue(suggestedPrompt)
	} else {
		model.SuggestedPrompt = types.StringNull()
	}
	if createdOn, ok := response["createdOn"].(string); ok {
		model.CreatedOn = types.StringValue(createdOn)
	} else {
		model.CreatedOn = types.StringNull()
	}
	if modifiedOn, ok := response["modifiedOn"].(string); ok {
		model.ModifiedOn = types.StringValue(modifiedOn)
	} else {
		model.ModifiedOn = types.StringNull()
	}
}
