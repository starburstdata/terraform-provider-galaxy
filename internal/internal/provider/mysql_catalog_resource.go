package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_mysql_catalog"
)

var _ resource.Resource = (*mysql_catalogResource)(nil)
var _ resource.ResourceWithConfigure = (*mysql_catalogResource)(nil)

func NewMysqlCatalogResource() resource.Resource {
	return &mysql_catalogResource{}
}

type mysql_catalogResource struct {
	client *client.GalaxyClient
}

// No need for extended model - use base model directly

func (r *mysql_catalogResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_catalog"
}

func (r *mysql_catalogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_mysql_catalog.MysqlCatalogResourceSchema(ctx)
}

func (r *mysql_catalogResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *mysql_catalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_mysql_catalog.MysqlCatalogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Initialize unknown values to null for optional fields
	if plan.PrivateLinkId.IsUnknown() {
		plan.PrivateLinkId = types.StringNull()
	}
	if plan.SshTunnelId.IsUnknown() {
		plan.SshTunnelId = types.StringNull()
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating mysql_catalog")
	response, err := r.client.CreateCatalog(ctx, "mysql", request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating mysql_catalog",
			"Could not create mysql_catalog: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created mysql_catalog", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mysql_catalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_mysql_catalog.MysqlCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Reading mysql_catalog", map[string]interface{}{"id": id})
	response, err := r.client.GetCatalog(ctx, "mysql", id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "MysqlCatalog not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading mysql_catalog",
			"Could not read mysql_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mysql_catalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_mysql_catalog.MysqlCatalogModel
	var state resource_mysql_catalog.MysqlCatalogModel

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

	tflog.Debug(ctx, "Updating mysql_catalog", map[string]interface{}{"id": id})
	response, err := r.client.UpdateCatalog(ctx, "mysql", id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating mysql_catalog",
			"Could not update mysql_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated mysql_catalog", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mysql_catalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_mysql_catalog.MysqlCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Deleting mysql_catalog", map[string]interface{}{"id": id})
	err := r.client.DeleteCatalog(ctx, "mysql", id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting mysql_catalog",
				"Could not delete mysql_catalog "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted mysql_catalog", map[string]interface{}{"id": id})
}

// Helper methods
func (r *mysql_catalogResource) modelToCreateRequest(ctx context.Context, model *resource_mysql_catalog.MysqlCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	request["name"] = model.Name.ValueString()
	request["readOnly"] = model.ReadOnly.ValueBool()
	request["connectionType"] = "direct" // For direct MySQL connection
	request["username"] = model.Username.ValueString()
	request["password"] = model.Password.ValueString()
	request["host"] = model.Host.ValueString()

	// Optional fields - MySQL doesn't use database field in this catalog type

	if !model.Port.IsNull() && !model.Port.IsUnknown() {
		request["port"] = model.Port.ValueInt64()
	} else {
		request["port"] = 3306 // Default MySQL port
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		request["description"] = model.Description.ValueString()
	}

	if !model.CloudKind.IsNull() && !model.CloudKind.IsUnknown() {
		request["cloudKind"] = model.CloudKind.ValueString()
	}

	if !model.Validate.IsNull() && !model.Validate.IsUnknown() {
		request["validate"] = model.Validate.ValueBool()
	}

	return request
}

func (r *mysql_catalogResource) modelToUpdateRequest(ctx context.Context, model *resource_mysql_catalog.MysqlCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *mysql_catalogResource) updateModelFromResponse(ctx context.Context, model *resource_mysql_catalog.MysqlCatalogModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
		model.CatalogId = types.StringValue(id)
	} else if id, ok := response["catalogId"].(string); ok {
		model.Id = types.StringValue(id)
		model.CatalogId = types.StringValue(id)
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

	if cloudKind, ok := response["cloudKind"].(string); ok {
		model.CloudKind = types.StringValue(cloudKind)
	} else if model.CloudKind.IsUnknown() {
		model.CloudKind = types.StringNull()
	}

	if connectionType, ok := response["connectionType"].(string); ok {
		model.ConnectionType = types.StringValue(connectionType)
	}

	if port, ok := response["port"].(float64); ok {
		model.Port = types.Int64Value(int64(port))
	}

	if username, ok := response["username"].(string); ok {
		model.Username = types.StringValue(username)
	}

	// Password is write-only, keep existing value
	// model.Password stays as-is

	if host, ok := response["host"].(string); ok {
		model.Host = types.StringValue(host)
	}

	if privateLinkId, ok := response["privateLinkId"].(string); ok && privateLinkId != "" {
		model.PrivateLinkId = types.StringValue(privateLinkId)
	} else {
		model.PrivateLinkId = types.StringNull()
	}

	if sshTunnelId, ok := response["sshTunnelId"].(string); ok && sshTunnelId != "" {
		model.SshTunnelId = types.StringValue(sshTunnelId)
	} else {
		model.SshTunnelId = types.StringNull()
	}

	// MySQL catalogs don't use separate database field

	// Handle validate field
	if validate, ok := response["validate"].(bool); ok {
		model.Validate = types.BoolValue(validate)
	} else if model.Validate.IsUnknown() {
		model.Validate = types.BoolNull()
	}

}
