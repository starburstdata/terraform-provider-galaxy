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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_postgresql_catalog"
)

var _ resource.Resource = (*postgresql_catalogResource)(nil)
var _ resource.ResourceWithConfigure = (*postgresql_catalogResource)(nil)

func NewPostgresqlCatalogResource() resource.Resource {
	return &postgresql_catalogResource{}
}

type postgresql_catalogResource struct {
	client *client.GalaxyClient
}

// PostgreSQL uses the base model as it already has all required fields

func (r *postgresql_catalogResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql_catalog"
}

func (r *postgresql_catalogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	// PostgreSQL already has endpoint and database_name in the generated schema
	resp.Schema = resource_postgresql_catalog.PostgresqlCatalogResourceSchema(ctx)
}

func (r *postgresql_catalogResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *postgresql_catalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_postgresql_catalog.PostgresqlCatalogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating postgresql_catalog")
	response, err := r.client.CreateCatalog(ctx, "postgresql", request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating postgresql_catalog",
			"Could not create postgresql_catalog: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created postgresql_catalog", map[string]interface{}{"id": plan.CatalogId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *postgresql_catalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_postgresql_catalog.PostgresqlCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	tflog.Debug(ctx, "Reading postgresql_catalog", map[string]interface{}{"id": id})
	response, err := r.client.GetCatalog(ctx, "postgresql", id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "PostgresqlCatalog not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading postgresql_catalog",
			"Could not read postgresql_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *postgresql_catalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_postgresql_catalog.PostgresqlCatalogModel
	var state resource_postgresql_catalog.PostgresqlCatalogModel

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

	tflog.Debug(ctx, "Updating postgresql_catalog", map[string]interface{}{"id": id})
	response, err := r.client.UpdateCatalog(ctx, "postgresql", id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating postgresql_catalog",
			"Could not update postgresql_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated postgresql_catalog", map[string]interface{}{"id": plan.CatalogId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *postgresql_catalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_postgresql_catalog.PostgresqlCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	tflog.Debug(ctx, "Deleting postgresql_catalog", map[string]interface{}{"id": id})
	err := r.client.DeleteCatalog(ctx, "postgresql", id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting postgresql_catalog",
				"Could not delete postgresql_catalog "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted postgresql_catalog", map[string]interface{}{"id": id})
}

// Helper methods
func (r *postgresql_catalogResource) modelToCreateRequest(ctx context.Context, model *resource_postgresql_catalog.PostgresqlCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	request["name"] = model.Name.ValueString()
	request["readOnly"] = model.ReadOnly.ValueBool()
	request["username"] = model.Username.ValueString()
	request["password"] = model.Password.ValueString()
	request["endpoint"] = model.Endpoint.ValueString()
	request["databaseName"] = model.DatabaseName.ValueString()

	// Optional fields
	if !model.Port.IsNull() && !model.Port.IsUnknown() {
		request["port"] = model.Port.ValueInt64()
	} else {
		request["port"] = 5432 // Default PostgreSQL port
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() && model.Description.ValueString() != "" {
		request["description"] = model.Description.ValueString()
	}

	if !model.Validate.IsNull() && !model.Validate.IsUnknown() {
		request["validate"] = model.Validate.ValueBool()
	}

	return request
}

func (r *postgresql_catalogResource) modelToUpdateRequest(ctx context.Context, model *resource_postgresql_catalog.PostgresqlCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *postgresql_catalogResource) updateModelFromResponse(ctx context.Context, model *resource_postgresql_catalog.PostgresqlCatalogModel, response map[string]interface{}, diags *diag.Diagnostics) {
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
	}

	if username, ok := response["username"].(string); ok {
		model.Username = types.StringValue(username)
	}

	// Password is write-only, keep existing value

	// Handle computed fields
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

	if tlsEnabled, ok := response["tlsEnabled"].(bool); ok {
		model.TlsEnabled = types.BoolValue(tlsEnabled)
	} else if model.TlsEnabled.IsUnknown() {
		model.TlsEnabled = types.BoolNull()
	}

	// Handle validate field
	if validate, ok := response["validate"].(bool); ok {
		model.Validate = types.BoolValue(validate)
	} else if model.Validate.IsUnknown() {
		model.Validate = types.BoolNull()
	}

}
