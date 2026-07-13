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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_data_quality_check"
)

var _ resource.Resource = (*dataQualityCheckResource)(nil)
var _ resource.ResourceWithConfigure = (*dataQualityCheckResource)(nil)
var _ resource.ResourceWithImportState = (*dataQualityCheckResource)(nil)

func NewDataQualityCheckResource() resource.Resource {
	return &dataQualityCheckResource{}
}

type dataQualityCheckResource struct {
	client *client.GalaxyClient
}

func (r *dataQualityCheckResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_quality_check"
}

func (r *dataQualityCheckResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_data_quality_check.DataQualityCheckResourceSchema(ctx)

	// The Galaxy UpdateDataQualityCheckPatch schema only accepts name, description, query,
	// severity, category, and clusterId. catalog_id, schema_id, table_id, and kind are
	// immutable at the API boundary, so changing them must destroy and recreate the check.
	for _, name := range []string{"catalog_id", "schema_id", "table_id", "kind"} {
		if attr, ok := s.Attributes[name].(schema.StringAttribute); ok {
			attr.PlanModifiers = append(attr.PlanModifiers, stringplanmodifier.RequiresReplace())
			s.Attributes[name] = attr
		}
	}

	// cluster_id and description are Optional+Computed; add UseStateForUnknown to prevent
	// recurring "(known after apply)" diffs
	for _, name := range []string{"cluster_id", "description"} {
		if attr, ok := s.Attributes[name].(schema.StringAttribute); ok {
			attr.PlanModifiers = append(attr.PlanModifiers, stringplanmodifier.UseStateForUnknown())
			s.Attributes[name] = attr
		}
	}

	resp.Schema = s
}

func (r *dataQualityCheckResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dataQualityCheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_data_quality_check.DataQualityCheckModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(&plan)

	tflog.Debug(ctx, "Creating data quality check")
	response, err := r.client.CreateDataQualityCheck(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating data quality check",
			"Could not create data quality check: "+err.Error(),
		)
		return
	}

	// Set defaults for computed fields that were unknown before creation
	if plan.DataQualityCheckId.IsUnknown() {
		plan.DataQualityCheckId = types.StringNull()
	}
	if plan.Description.IsUnknown() {
		plan.Description = types.StringNull()
	}
	if plan.ClusterId.IsUnknown() {
		plan.ClusterId = types.StringNull()
	}

	r.updateModelFromResponse(&plan, response)

	tflog.Debug(ctx, "Created data quality check", map[string]interface{}{"id": plan.DataQualityCheckId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dataQualityCheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_data_quality_check.DataQualityCheckModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.DataQualityCheckId.ValueString()
	tflog.Debug(ctx, "Reading data quality check", map[string]interface{}{"id": id})
	response, err := r.client.GetDataQualityCheck(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Data quality check not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading data quality check",
			"Could not read data quality check "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(&state, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dataQualityCheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_data_quality_check.DataQualityCheckModel
	var state resource_data_quality_check.DataQualityCheckModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.DataQualityCheckId.ValueString()
	request := r.modelToCreateRequest(&plan)

	tflog.Debug(ctx, "Updating data quality check", map[string]interface{}{"id": id})
	response, err := r.client.UpdateDataQualityCheck(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating data quality check",
			"Could not update data quality check "+id+": "+err.Error(),
		)
		return
	}

	plan.DataQualityCheckId = state.DataQualityCheckId
	r.updateModelFromResponse(&plan, response)

	tflog.Debug(ctx, "Updated data quality check", map[string]interface{}{"id": plan.DataQualityCheckId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dataQualityCheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_data_quality_check.DataQualityCheckModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.DataQualityCheckId.ValueString()
	tflog.Debug(ctx, "Deleting data quality check", map[string]interface{}{"id": id})
	err := r.client.DeleteDataQualityCheck(ctx, id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting data quality check",
				"Could not delete data quality check "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted data quality check", map[string]interface{}{"id": id})
}

func (r *dataQualityCheckResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("data_quality_check_id"), req, resp)
}

// Helper methods
func (r *dataQualityCheckResource) modelToCreateRequest(model *resource_data_quality_check.DataQualityCheckModel) map[string]interface{} {
	request := make(map[string]interface{})

	if !model.CatalogId.IsNull() && !model.CatalogId.IsUnknown() && model.CatalogId.ValueString() != "" {
		request["catalogId"] = model.CatalogId.ValueString()
	}

	if !model.Category.IsNull() && !model.Category.IsUnknown() && model.Category.ValueString() != "" {
		request["category"] = model.Category.ValueString()
	}

	if !model.ClusterId.IsNull() && !model.ClusterId.IsUnknown() && model.ClusterId.ValueString() != "" {
		request["clusterId"] = model.ClusterId.ValueString()
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() && model.Description.ValueString() != "" {
		request["description"] = model.Description.ValueString()
	}

	if !model.Kind.IsNull() && !model.Kind.IsUnknown() && model.Kind.ValueString() != "" {
		request["kind"] = model.Kind.ValueString()
	}

	if !model.Name.IsNull() && !model.Name.IsUnknown() && model.Name.ValueString() != "" {
		request["name"] = model.Name.ValueString()
	}

	if !model.Query.IsNull() && !model.Query.IsUnknown() && model.Query.ValueString() != "" {
		request["query"] = model.Query.ValueString()
	}

	if !model.SchemaId.IsNull() && !model.SchemaId.IsUnknown() && model.SchemaId.ValueString() != "" {
		request["schemaId"] = model.SchemaId.ValueString()
	}

	if !model.Severity.IsNull() && !model.Severity.IsUnknown() && model.Severity.ValueString() != "" {
		request["severity"] = model.Severity.ValueString()
	}

	if !model.TableId.IsNull() && !model.TableId.IsUnknown() && model.TableId.ValueString() != "" {
		request["tableId"] = model.TableId.ValueString()
	}

	return request
}

func (r *dataQualityCheckResource) updateModelFromResponse(model *resource_data_quality_check.DataQualityCheckModel, response map[string]interface{}) {
	if id, ok := response["dataQualityCheckId"].(string); ok {
		model.DataQualityCheckId = types.StringValue(id)
	}

	if catalogId, ok := response["catalogId"].(string); ok {
		model.CatalogId = types.StringValue(catalogId)
	}

	if category, ok := response["category"].(string); ok {
		model.Category = types.StringValue(category)
	}

	if clusterId, ok := response["clusterId"].(string); ok {
		model.ClusterId = types.StringValue(clusterId)
	} else if model.ClusterId.IsUnknown() {
		model.ClusterId = types.StringNull()
	}

	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else if model.Description.IsUnknown() {
		model.Description = types.StringNull()
	}

	if kind, ok := response["kind"].(string); ok {
		model.Kind = types.StringValue(kind)
	}

	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}

	if query, ok := response["query"].(string); ok {
		model.Query = types.StringValue(query)
	}

	if schemaId, ok := response["schemaId"].(string); ok {
		model.SchemaId = types.StringValue(schemaId)
	}

	if severity, ok := response["severity"].(string); ok {
		model.Severity = types.StringValue(severity)
	}

	if tableId, ok := response["tableId"].(string); ok {
		model.TableId = types.StringValue(tableId)
	}
}
