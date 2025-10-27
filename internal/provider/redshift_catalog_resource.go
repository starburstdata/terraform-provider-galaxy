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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_redshift_catalog"
)

var _ resource.Resource = (*redshift_catalogResource)(nil)
var _ resource.ResourceWithConfigure = (*redshift_catalogResource)(nil)

func NewRedshiftCatalogResource() resource.Resource {
	return &redshift_catalogResource{}
}

type redshift_catalogResource struct {
	client *client.GalaxyClient
}

// Use base model directly

func (r *redshift_catalogResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redshift_catalog"
}

func (r *redshift_catalogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_redshift_catalog.RedshiftCatalogResourceSchema(ctx)
}

func (r *redshift_catalogResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *redshift_catalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_redshift_catalog.RedshiftCatalogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Initialize unknown values to null for optional fields
	if plan.AccessKey.IsUnknown() {
		plan.AccessKey = types.StringNull()
	}
	if plan.Region.IsUnknown() {
		plan.Region = types.StringNull()
	}
	if plan.RoleArn.IsUnknown() {
		plan.RoleArn = types.StringNull()
	}
	if plan.SecretKey.IsUnknown() {
		plan.SecretKey = types.StringNull()
	}
	if plan.SshTunnelId.IsUnknown() {
		plan.SshTunnelId = types.StringNull()
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating redshift_catalog")
	response, err := r.client.CreateCatalog(ctx, "redshift", request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating redshift_catalog",
			"Could not create redshift_catalog: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created redshift_catalog", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *redshift_catalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_redshift_catalog.RedshiftCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Reading redshift_catalog", map[string]interface{}{"id": id})
	response, err := r.client.GetCatalog(ctx, "redshift", id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "RedshiftCatalog not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading redshift_catalog",
			"Could not read redshift_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *redshift_catalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_redshift_catalog.RedshiftCatalogModel
	var state resource_redshift_catalog.RedshiftCatalogModel

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

	tflog.Debug(ctx, "Updating redshift_catalog", map[string]interface{}{"id": id})
	response, err := r.client.UpdateCatalog(ctx, "redshift", id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating redshift_catalog",
			"Could not update redshift_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated redshift_catalog", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *redshift_catalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_redshift_catalog.RedshiftCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Deleting redshift_catalog", map[string]interface{}{"id": id})
	err := r.client.DeleteCatalog(ctx, "redshift", id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting redshift_catalog",
				"Could not delete redshift_catalog "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted redshift_catalog", map[string]interface{}{"id": id})
}

// Helper methods
func (r *redshift_catalogResource) modelToCreateRequest(ctx context.Context, model *resource_redshift_catalog.RedshiftCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	request["name"] = model.Name.ValueString()
	request["readOnly"] = model.ReadOnly.ValueBool()
	request["authType"] = model.AuthType.ValueString()

	// Use endpoint directly (includes host:port/database)
	request["endpoint"] = model.Endpoint.ValueString()

	request["username"] = model.Username.ValueString()
	request["password"] = model.Password.ValueString()

	// Optional fields
	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		request["description"] = model.Description.ValueString()
	}

	if !model.SshTunnelId.IsNull() && !model.SshTunnelId.IsUnknown() {
		request["sshTunnelId"] = model.SshTunnelId.ValueString()
	}

	if !model.Validate.IsNull() && !model.Validate.IsUnknown() {
		request["validate"] = model.Validate.ValueBool()
	}

	return request
}

func (r *redshift_catalogResource) modelToUpdateRequest(ctx context.Context, model *resource_redshift_catalog.RedshiftCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *redshift_catalogResource) updateModelFromResponse(ctx context.Context, model *resource_redshift_catalog.RedshiftCatalogModel, response map[string]interface{}, diags *diag.Diagnostics) {
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

	if authType, ok := response["authType"].(string); ok {
		model.AuthType = types.StringValue(authType)
	}

	if endpoint, ok := response["endpoint"].(string); ok {
		// Store the full endpoint directly (includes host:port/database)
		model.Endpoint = types.StringValue(endpoint)
	}

	if username, ok := response["username"].(string); ok {
		model.Username = types.StringValue(username)
	}

	// Password is write-only, keep existing value

	if accessKey, ok := response["accessKey"].(string); ok && accessKey != "" {
		model.AccessKey = types.StringValue(accessKey)
	} else {
		model.AccessKey = types.StringNull()
	}

	if region, ok := response["region"].(string); ok && region != "" {
		model.Region = types.StringValue(region)
	} else {
		model.Region = types.StringNull()
	}

	if roleArn, ok := response["roleArn"].(string); ok && roleArn != "" {
		model.RoleArn = types.StringValue(roleArn)
	} else {
		model.RoleArn = types.StringNull()
	}

	if secretKey, ok := response["secretKey"].(string); ok && secretKey != "" {
		model.SecretKey = types.StringValue(secretKey)
	} else {
		model.SecretKey = types.StringNull()
	}

	if sshTunnelId, ok := response["sshTunnelId"].(string); ok && sshTunnelId != "" {
		model.SshTunnelId = types.StringValue(sshTunnelId)
	} else {
		model.SshTunnelId = types.StringNull()
	}

	// Handle validate field
	if validate, ok := response["validate"].(bool); ok {
		model.Validate = types.BoolValue(validate)
	} else if model.Validate.IsUnknown() {
		model.Validate = types.BoolNull()
	}

}
