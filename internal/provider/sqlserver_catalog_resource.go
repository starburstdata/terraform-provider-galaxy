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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_sqlserver_catalog"
)

var _ resource.Resource = (*sqlserver_catalogResource)(nil)
var _ resource.ResourceWithConfigure = (*sqlserver_catalogResource)(nil)

func NewSqlserverCatalogResource() resource.Resource {
	return &sqlserver_catalogResource{}
}

type sqlserver_catalogResource struct {
	client *client.GalaxyClient
}

// SQLServer uses the base model as it already has all required fields

func (r *sqlserver_catalogResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sqlserver_catalog"
}

func (r *sqlserver_catalogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// SQLServer already has endpoint, database_name, username, password in the generated schema
	s := resource_sqlserver_catalog.SqlserverCatalogResourceSchema(ctx)

	// Fix: validate is a request-only parameter, not returned by API.
	// Setting Computed=false ensures it's sent with update requests.
	if attr, ok := s.Attributes["validate"].(schema.BoolAttribute); ok {
		attr.Computed = false
		s.Attributes["validate"] = attr
	}

	resp.Schema = s
}

func (r *sqlserver_catalogResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *sqlserver_catalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_sqlserver_catalog.SqlserverCatalogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating sqlserver_catalog")
	response, err := r.client.CreateCatalog(ctx, "sqlserver", request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating sqlserver_catalog",
			"Could not create sqlserver_catalog: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created sqlserver_catalog", map[string]interface{}{"id": plan.CatalogId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sqlserver_catalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_sqlserver_catalog.SqlserverCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	tflog.Debug(ctx, "Reading sqlserver_catalog", map[string]interface{}{"id": id})
	response, err := r.client.GetCatalog(ctx, "sqlserver", id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "SqlserverCatalog not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading sqlserver_catalog",
			"Could not read sqlserver_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *sqlserver_catalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_sqlserver_catalog.SqlserverCatalogModel
	var state resource_sqlserver_catalog.SqlserverCatalogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	request := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating sqlserver_catalog", map[string]interface{}{"id": id})
	response, err := r.client.UpdateCatalog(ctx, "sqlserver", id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating sqlserver_catalog",
			"Could not update sqlserver_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated sqlserver_catalog", map[string]interface{}{"id": plan.CatalogId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *sqlserver_catalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_sqlserver_catalog.SqlserverCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	tflog.Debug(ctx, "Deleting sqlserver_catalog", map[string]interface{}{"id": id})
	err := r.client.DeleteCatalog(ctx, "sqlserver", id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting sqlserver_catalog",
				"Could not delete sqlserver_catalog "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted sqlserver_catalog", map[string]interface{}{"id": id})
}

// Helper methods
func (r *sqlserver_catalogResource) modelToCreateRequest(ctx context.Context, model *resource_sqlserver_catalog.SqlserverCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	request["name"] = model.Name.ValueString()
	request["readOnly"] = model.ReadOnly.ValueBool()
	request["endpoint"] = model.Endpoint.ValueString()
	request["databaseName"] = model.DatabaseName.ValueString()
	request["username"] = model.Username.ValueString()
	request["password"] = model.Password.ValueString()

	// Optional fields
	if !model.Port.IsNull() && !model.Port.IsUnknown() {
		request["port"] = model.Port.ValueInt64()
	} else {
		request["port"] = 1433 // Default SQL Server port
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() && model.Description.ValueString() != "" {
		request["description"] = model.Description.ValueString()
	}

	if !model.CloudKind.IsNull() && !model.CloudKind.IsUnknown() && model.CloudKind.ValueString() != "" {
		request["cloudKind"] = model.CloudKind.ValueString()
	}

	if !model.SshTunnelId.IsNull() && !model.SshTunnelId.IsUnknown() && model.SshTunnelId.ValueString() != "" {
		request["sshTunnelId"] = model.SshTunnelId.ValueString()
	}

	if !model.Validate.IsNull() && !model.Validate.IsUnknown() {
		request["validate"] = model.Validate.ValueBool()
	}

	return request
}

func (r *sqlserver_catalogResource) modelToUpdateRequest(ctx context.Context, model *resource_sqlserver_catalog.SqlserverCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *sqlserver_catalogResource) updateModelFromResponse(ctx context.Context, model *resource_sqlserver_catalog.SqlserverCatalogModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
	if catalogId, ok := response["catalogId"].(string); ok {
		model.CatalogId = types.StringValue(catalogId)
	}

	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}

	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else if model.Description.IsUnknown() {
		model.Description = types.StringNull()
	}

	if readOnly, ok := response["readOnly"].(bool); ok {
		model.ReadOnly = types.BoolValue(readOnly)
	}

	if endpoint, ok := response["endpoint"].(string); ok {
		model.Endpoint = types.StringValue(endpoint)
	}

	if databaseName, ok := response["databaseName"].(string); ok {
		model.DatabaseName = types.StringValue(databaseName)
	}

	if port, ok := response["port"].(float64); ok {
		model.Port = types.Int64Value(int64(port))
	} else if model.Port.IsUnknown() {
		model.Port = types.Int64Value(1433) // Default SQL Server port
	}

	if username, ok := response["username"].(string); ok {
		model.Username = types.StringValue(username)
	}

	// Password is write-only, keep existing value

	if cloudKind, ok := response["cloudKind"].(string); ok {
		model.CloudKind = types.StringValue(cloudKind)
	} else if model.CloudKind.IsUnknown() {
		model.CloudKind = types.StringNull()
	}

	if sshTunnelId, ok := response["sshTunnelId"].(string); ok {
		model.SshTunnelId = types.StringValue(sshTunnelId)
	} else if model.SshTunnelId.IsUnknown() {
		model.SshTunnelId = types.StringNull()
	}

	// Handle validate field
	if validate, ok := response["validate"].(bool); ok {
		model.Validate = types.BoolValue(validate)
	} else if model.Validate.IsUnknown() {
		model.Validate = types.BoolNull()
	}

}
