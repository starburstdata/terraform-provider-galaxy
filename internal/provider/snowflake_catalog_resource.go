package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_snowflake_catalog"
)

var _ resource.Resource = (*snowflake_catalogResource)(nil)
var _ resource.ResourceWithConfigure = (*snowflake_catalogResource)(nil)

func NewSnowflakeCatalogResource() resource.Resource {
	return &snowflake_catalogResource{}
}

type snowflake_catalogResource struct {
	client *client.GalaxyClient
}

func (r *snowflake_catalogResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snowflake_catalog"
}

func (r *snowflake_catalogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_snowflake_catalog.SnowflakeCatalogResourceSchema(ctx)
}

func (r *snowflake_catalogResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *snowflake_catalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_snowflake_catalog.SnowflakeCatalogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating snowflake_catalog")
	response, err := r.client.CreateCatalog(ctx, "snowflake", request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating snowflake_catalog",
			"Could not create snowflake_catalog: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created snowflake_catalog", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *snowflake_catalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_snowflake_catalog.SnowflakeCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use either id or catalog_id from state - prefer catalog_id if id is empty
	id := state.Id.ValueString()
	if id == "" && !state.CatalogId.IsNull() {
		id = state.CatalogId.ValueString()
	}
	if id == "" {
		resp.Diagnostics.AddError(
			"Missing ID for snowflake_catalog",
			"Both id and catalog_id are empty in state",
		)
		return
	}

	tflog.Debug(ctx, "Reading snowflake_catalog", map[string]interface{}{"id": id})
	response, err := r.client.GetCatalog(ctx, "snowflake", id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "SnowflakeCatalog not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading snowflake_catalog",
			"Could not read snowflake_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *snowflake_catalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_snowflake_catalog.SnowflakeCatalogModel
	var state resource_snowflake_catalog.SnowflakeCatalogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use either id or catalog_id from state - prefer catalog_id if id is empty
	id := state.Id.ValueString()
	if id == "" && !state.CatalogId.IsNull() {
		id = state.CatalogId.ValueString()
	}
	if id == "" {
		resp.Diagnostics.AddError(
			"Missing ID for snowflake_catalog",
			"Both id and catalog_id are empty in state",
		)
		return
	}

	request := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating snowflake_catalog", map[string]interface{}{"id": id})
	response, err := r.client.UpdateCatalog(ctx, "snowflake", id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating snowflake_catalog",
			"Could not update snowflake_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated snowflake_catalog", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *snowflake_catalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_snowflake_catalog.SnowflakeCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use either id or catalog_id from state - prefer catalog_id if id is empty
	id := state.Id.ValueString()
	if id == "" && !state.CatalogId.IsNull() {
		id = state.CatalogId.ValueString()
	}
	if id == "" {
		resp.Diagnostics.AddError(
			"Missing ID for snowflake_catalog",
			"Both id and catalog_id are empty in state",
		)
		return
	}

	tflog.Debug(ctx, "Deleting snowflake_catalog", map[string]interface{}{"id": id})
	err := r.client.DeleteCatalog(ctx, "snowflake", id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting snowflake_catalog",
				"Could not delete snowflake_catalog "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted snowflake_catalog", map[string]interface{}{"id": id})
}

// Helper methods
func (r *snowflake_catalogResource) modelToCreateRequest(ctx context.Context, model *resource_snowflake_catalog.SnowflakeCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields - matching the working curl script exactly
	if !model.Name.IsNull() {
		request["name"] = model.Name.ValueString()
	}
	if !model.AccountIdentifier.IsNull() {
		request["accountIdentifier"] = model.AccountIdentifier.ValueString()
	}
	if !model.DatabaseName.IsNull() {
		request["databaseName"] = model.DatabaseName.ValueString()
	}
	if !model.Username.IsNull() {
		request["username"] = model.Username.ValueString()
	}
	if !model.Password.IsNull() {
		request["password"] = model.Password.ValueString()
	}
	if !model.ReadOnly.IsNull() {
		request["readOnly"] = model.ReadOnly.ValueBool()
	}

	// Only include description if provided (like curl script)
	if !model.Description.IsNull() {
		request["description"] = model.Description.ValueString()
	}

	// SKIP optional fields that cause "Unrecognized entity" error:
	// - cloudKind (server will default to AWS)
	// - role (not needed for basic authentication)
	// - warehouse (not needed for basic authentication)

	// Optional authentication fields
	if !model.AuthenticationType.IsNull() && !model.AuthenticationType.IsUnknown() {
		request["authenticationType"] = model.AuthenticationType.ValueString()
	}

	if !model.PrivateKey.IsNull() && !model.PrivateKey.IsUnknown() {
		request["privateKey"] = model.PrivateKey.ValueString()
	}

	if !model.PrivateKeyPassphrase.IsNull() && !model.PrivateKeyPassphrase.IsUnknown() {
		request["privateKeyPassphrase"] = model.PrivateKeyPassphrase.ValueString()
	}

	if !model.Validate.IsNull() && !model.Validate.IsUnknown() {
		request["validate"] = model.Validate.ValueBool()
	}

	return request
}

func (r *snowflake_catalogResource) modelToUpdateRequest(ctx context.Context, model *resource_snowflake_catalog.SnowflakeCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *snowflake_catalogResource) updateModelFromResponse(ctx context.Context, model *resource_snowflake_catalog.SnowflakeCatalogModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
	// The API returns catalogId as the primary identifier
	if catalogId, ok := response["catalogId"].(string); ok {
		model.Id = types.StringValue(catalogId)        // Use catalogId as the main ID
		model.CatalogId = types.StringValue(catalogId) // Also set the catalogId field
	} else if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	} else if id, ok := response["snowflakeCatalogId"].(string); ok {
		model.Id = types.StringValue(id)
	}

	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}

	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else {
		model.Description = types.StringNull()
	}

	if accountIdentifier, ok := response["accountIdentifier"].(string); ok {
		model.AccountIdentifier = types.StringValue(accountIdentifier)
	}

	if databaseName, ok := response["databaseName"].(string); ok {
		model.DatabaseName = types.StringValue(databaseName)
	}

	if username, ok := response["username"].(string); ok {
		model.Username = types.StringValue(username)
	}

	if readOnly, ok := response["readOnly"].(bool); ok {
		model.ReadOnly = types.BoolValue(readOnly)
	}

	if cloudKind, ok := response["cloudKind"].(string); ok {
		model.CloudKind = types.StringValue(cloudKind)
	} else {
		model.CloudKind = types.StringValue("AWS") // Default value
	}

	if role, ok := response["role"].(string); ok {
		model.Role = types.StringValue(role)
	} else {
		model.Role = types.StringNull()
	}

	if warehouse, ok := response["warehouse"].(string); ok {
		model.Warehouse = types.StringValue(warehouse)
	} else {
		model.Warehouse = types.StringNull()
	}

	// Note: password is not returned in responses for security reasons

	// Handle validate field
	if validate, ok := response["validate"].(bool); ok {
		model.Validate = types.BoolValue(validate)
	} else if model.Validate.IsUnknown() {
		model.Validate = types.BoolNull()
	}

	// Handle authentication_type field
	if authenticationType, ok := response["authenticationType"].(string); ok {
		model.AuthenticationType = types.StringValue(authenticationType)
	} else if model.AuthenticationType.IsUnknown() {
		model.AuthenticationType = types.StringNull()
	}

	// Handle private_key field (typically not returned for security reasons)
	if privateKey, ok := response["privateKey"].(string); ok {
		model.PrivateKey = types.StringValue(privateKey)
	} else if model.PrivateKey.IsUnknown() {
		model.PrivateKey = types.StringNull()
	}

	// Handle private_key_passphrase field (typically not returned for security reasons)
	if privateKeyPassphrase, ok := response["privateKeyPassphrase"].(string); ok {
		model.PrivateKeyPassphrase = types.StringValue(privateKeyPassphrase)
	} else if model.PrivateKeyPassphrase.IsUnknown() {
		model.PrivateKeyPassphrase = types.StringNull()
	}

}
