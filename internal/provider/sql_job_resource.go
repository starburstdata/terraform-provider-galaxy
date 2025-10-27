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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_sql_job"
)

var _ resource.Resource = (*sqlJobResource)(nil)
var _ resource.ResourceWithConfigure = (*sqlJobResource)(nil)

func NewSqlJobResource() resource.Resource {
	return &sqlJobResource{}
}

type sqlJobResource struct {
	client *client.GalaxyClient
}

func (r *sqlJobResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_job"
}

func (r *sqlJobResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_sql_job.SqlJobResourceSchema(ctx)
}

func (r *sqlJobResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sqlJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_sql_job.SqlJobModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating sql_job")
	response, err := r.client.CreateSqlJob(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating sql_job",
			"Could not create sql_job: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created sql_job", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sqlJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_sql_job.SqlJobModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the SQL job ID from state
	id := state.SqlJobId.ValueString()
	if id == "" {
		id = state.Id.ValueString()
	}

	tflog.Debug(ctx, "Reading sql_job", map[string]interface{}{"id": id})
	response, err := r.client.GetSqlJob(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "SqlJob not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading sql_job",
			"Could not read sql_job "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sqlJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_sql_job.SqlJobModel
	var state resource_sql_job.SqlJobModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the SQL job ID from state
	id := state.SqlJobId.ValueString()
	if id == "" {
		id = state.Id.ValueString()
	}

	request := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating sql_job", map[string]interface{}{"id": id})
	response, err := r.client.UpdateSqlJob(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating sql_job",
			"Could not update sql_job "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sqlJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_sql_job.SqlJobModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the SQL job ID from state
	id := state.SqlJobId.ValueString()
	if id == "" {
		id = state.Id.ValueString()
	}

	tflog.Debug(ctx, "Deleting sql_job", map[string]interface{}{"id": id})
	err := r.client.DeleteSqlJob(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting sql_job",
			"Could not delete sql_job "+id+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted sql_job", map[string]interface{}{"id": id})
}

// Helper functions to convert between model and API request/response
func (r *sqlJobResource) modelToCreateRequest(ctx context.Context, model *resource_sql_job.SqlJobModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	if !model.Name.IsNull() && !model.Name.IsUnknown() {
		request["name"] = model.Name.ValueString()
	}
	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		request["description"] = model.Description.ValueString()
	}
	if !model.ClusterId.IsNull() && !model.ClusterId.IsUnknown() {
		request["clusterId"] = model.ClusterId.ValueString()
	}
	if !model.RoleId.IsNull() && !model.RoleId.IsUnknown() {
		request["roleId"] = model.RoleId.ValueString()
	}
	if !model.Query.IsNull() && !model.Query.IsUnknown() {
		request["query"] = model.Query.ValueString()
	}
	if !model.CronExpression.IsNull() && !model.CronExpression.IsUnknown() {
		request["cronExpression"] = model.CronExpression.ValueString()
	}
	if !model.Timezone.IsNull() && !model.Timezone.IsUnknown() {
		request["timezone"] = model.Timezone.ValueString()
	}

	return request
}

func (r *sqlJobResource) modelToUpdateRequest(ctx context.Context, model *resource_sql_job.SqlJobModel, diags *diag.Diagnostics) map[string]interface{} {
	// For update operations, use the same structure as create
	return r.modelToCreateRequest(ctx, model, diags)
}

func (r *sqlJobResource) updateModelFromResponse(ctx context.Context, model *resource_sql_job.SqlJobModel, response map[string]interface{}, diags *diag.Diagnostics) {
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	}
	if sqlJobId, ok := response["sqlJobId"].(string); ok {
		model.SqlJobId = types.StringValue(sqlJobId)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	}
	if clusterId, ok := response["clusterId"].(string); ok {
		model.ClusterId = types.StringValue(clusterId)
	}
	if roleId, ok := response["roleId"].(string); ok {
		model.RoleId = types.StringValue(roleId)
	}
	if query, ok := response["query"].(string); ok {
		model.Query = types.StringValue(query)
	}
	if cronExpression, ok := response["cronExpression"].(string); ok {
		model.CronExpression = types.StringValue(cronExpression)
	}
	if timezone, ok := response["timezone"].(string); ok {
		model.Timezone = types.StringValue(timezone)
	}
	if nextExecution, ok := response["nextExecution"].(string); ok {
		model.NextExecution = types.StringValue(nextExecution)
	}
}
