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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_opensearch_catalog"
)

var _ resource.Resource = (*opensearch_catalogResource)(nil)
var _ resource.ResourceWithConfigure = (*opensearch_catalogResource)(nil)

func NewOpensearchCatalogResource() resource.Resource {
	return &opensearch_catalogResource{}
}

type opensearch_catalogResource struct {
	client *client.GalaxyClient
}

// Use base model directly

func (r *opensearch_catalogResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_opensearch_catalog"
}

func (r *opensearch_catalogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_opensearch_catalog.OpensearchCatalogResourceSchema(ctx)

	// Fix: validate is a request-only parameter, not returned by API.
	// Setting Computed=false ensures it's sent with update requests.
	if attr, ok := s.Attributes["validate"].(schema.BoolAttribute); ok {
		attr.Computed = false
		s.Attributes["validate"] = attr
	}

	resp.Schema = s
}

func (r *opensearch_catalogResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *opensearch_catalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_opensearch_catalog.OpensearchCatalogModel

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

	tflog.Debug(ctx, "Creating opensearch_catalog")
	response, err := r.client.CreateCatalog(ctx, "opensearch", request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating opensearch_catalog",
			"Could not create opensearch_catalog: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created opensearch_catalog", map[string]interface{}{"id": plan.CatalogId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *opensearch_catalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_opensearch_catalog.OpensearchCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	tflog.Debug(ctx, "Reading opensearch_catalog", map[string]interface{}{"id": id})
	response, err := r.client.GetCatalog(ctx, "opensearch", id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "OpensearchCatalog not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading opensearch_catalog",
			"Could not read opensearch_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *opensearch_catalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_opensearch_catalog.OpensearchCatalogModel
	var state resource_opensearch_catalog.OpensearchCatalogModel

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

	tflog.Debug(ctx, "Updating opensearch_catalog", map[string]interface{}{"id": id})
	response, err := r.client.UpdateCatalog(ctx, "opensearch", id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating opensearch_catalog",
			"Could not update opensearch_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated opensearch_catalog", map[string]interface{}{"id": plan.CatalogId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *opensearch_catalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_opensearch_catalog.OpensearchCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	tflog.Debug(ctx, "Deleting opensearch_catalog", map[string]interface{}{"id": id})
	err := r.client.DeleteCatalog(ctx, "opensearch", id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting opensearch_catalog",
				"Could not delete opensearch_catalog "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted opensearch_catalog", map[string]interface{}{"id": id})
}

// Helper methods
func (r *opensearch_catalogResource) modelToCreateRequest(ctx context.Context, model *resource_opensearch_catalog.OpensearchCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	request["name"] = model.Name.ValueString()
	request["endpoint"] = model.Endpoint.ValueString()
	request["authType"] = model.AuthType.ValueString()
	request["readOnly"] = model.ReadOnly.ValueBool()

	// Optional fields
	if !model.Port.IsNull() && !model.Port.IsUnknown() {
		request["port"] = model.Port.ValueInt64()
	} else {
		request["port"] = 443 // Default OpenSearch port
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() && model.Description.ValueString() != "" {
		request["description"] = model.Description.ValueString()
	}

	// Handle authentication based on authType
	if model.AuthType.ValueString() == "basic" {
		if !model.Username.IsNull() && !model.Username.IsUnknown() && model.Username.ValueString() != "" {
			request["username"] = model.Username.ValueString()
		}
		if !model.Password.IsNull() && !model.Password.IsUnknown() && model.Password.ValueString() != "" {
			request["password"] = model.Password.ValueString()
		}
	}

	if !model.SshTunnelId.IsNull() && !model.SshTunnelId.IsUnknown() && model.SshTunnelId.ValueString() != "" {
		request["sshTunnelId"] = model.SshTunnelId.ValueString()
	}

	if !model.Validate.IsNull() && !model.Validate.IsUnknown() {
		request["validate"] = model.Validate.ValueBool()
	}

	return request
}

func (r *opensearch_catalogResource) modelToUpdateRequest(ctx context.Context, model *resource_opensearch_catalog.OpensearchCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *opensearch_catalogResource) updateModelFromResponse(ctx context.Context, model *resource_opensearch_catalog.OpensearchCatalogModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
	if id, ok := response["catalogId"].(string); ok {
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

	if authType, ok := response["authType"].(string); ok {
		model.AuthType = types.StringValue(authType)
	}

	if endpoint, ok := response["endpoint"].(string); ok {
		model.Endpoint = types.StringValue(endpoint)
	}

	if port, ok := response["port"].(float64); ok {
		model.Port = types.Int64Value(int64(port))
	}

	if readOnly, ok := response["readOnly"].(bool); ok {
		model.ReadOnly = types.BoolValue(readOnly)
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
