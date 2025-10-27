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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_cassandra_catalog"
)

var _ resource.Resource = (*cassandra_catalogResource)(nil)
var _ resource.ResourceWithConfigure = (*cassandra_catalogResource)(nil)

func NewCassandraCatalogResource() resource.Resource {
	return &cassandra_catalogResource{}
}

type cassandra_catalogResource struct {
	client *client.GalaxyClient
}

// Use the generated model directly
type CassandraCatalogModel = resource_cassandra_catalog.CassandraCatalogModel

func (r *cassandra_catalogResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cassandra_catalog"
}

func (r *cassandra_catalogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_cassandra_catalog.CassandraCatalogResourceSchema(ctx)
}

func (r *cassandra_catalogResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *cassandra_catalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CassandraCatalogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating cassandra_catalog")
	response, err := r.client.CreateCatalog(ctx, "cassandra", request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating cassandra_catalog",
			"Could not create cassandra_catalog: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created cassandra_catalog", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cassandra_catalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CassandraCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Reading cassandra_catalog", map[string]interface{}{"id": id})
	response, err := r.client.GetCatalog(ctx, "cassandra", id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "CassandraCatalog not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading cassandra_catalog",
			"Could not read cassandra_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cassandra_catalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CassandraCatalogModel
	var state CassandraCatalogModel

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

	tflog.Debug(ctx, "Updating cassandra_catalog", map[string]interface{}{"id": id})
	response, err := r.client.UpdateCatalog(ctx, "cassandra", id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating cassandra_catalog",
			"Could not update cassandra_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated cassandra_catalog", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cassandra_catalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CassandraCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Deleting cassandra_catalog", map[string]interface{}{"id": id})
	err := r.client.DeleteCatalog(ctx, "cassandra", id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting cassandra_catalog",
				"Could not delete cassandra_catalog "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted cassandra_catalog", map[string]interface{}{"id": id})
}

// Helper methods
func (r *cassandra_catalogResource) modelToCreateRequest(ctx context.Context, model *CassandraCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	request["name"] = model.Name.ValueString()
	request["readOnly"] = model.ReadOnly.ValueBool()
	request["deploymentType"] = model.DeploymentType.ValueString()

	// Optional authentication fields
	if !model.Username.IsNull() && !model.Username.IsUnknown() {
		request["username"] = model.Username.ValueString()
	}
	if !model.Password.IsNull() && !model.Password.IsUnknown() {
		request["password"] = model.Password.ValueString()
	}

	// Connection details
	if !model.ContactPoints.IsNull() && !model.ContactPoints.IsUnknown() {
		request["contactPoints"] = model.ContactPoints.ValueString()
	}
	if !model.LocalDatacenter.IsNull() && !model.LocalDatacenter.IsUnknown() {
		request["localDatacenter"] = model.LocalDatacenter.ValueString()
	}

	// Optional fields
	if !model.Port.IsNull() && !model.Port.IsUnknown() {
		request["port"] = model.Port.ValueInt64()
	}

	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		request["description"] = model.Description.ValueString()
	}

	if !model.CloudKind.IsNull() && !model.CloudKind.IsUnknown() {
		request["cloudKind"] = model.CloudKind.ValueString()
	}

	if !model.DatabaseId.IsNull() && !model.DatabaseId.IsUnknown() {
		request["databaseId"] = model.DatabaseId.ValueString()
	}

	if !model.Region.IsNull() && !model.Region.IsUnknown() {
		request["region"] = model.Region.ValueString()
	}

	if !model.SshTunnelId.IsNull() && !model.SshTunnelId.IsUnknown() {
		request["sshTunnelId"] = model.SshTunnelId.ValueString()
	}

	if !model.Token.IsNull() && !model.Token.IsUnknown() {
		request["token"] = model.Token.ValueString()
	}

	if !model.Validate.IsNull() && !model.Validate.IsUnknown() {
		request["validate"] = model.Validate.ValueBool()
	}

	return request
}

func (r *cassandra_catalogResource) modelToUpdateRequest(ctx context.Context, model *CassandraCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *cassandra_catalogResource) updateModelFromResponse(ctx context.Context, model *CassandraCatalogModel, response map[string]interface{}, diags *diag.Diagnostics) {
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

	if deploymentType, ok := response["deploymentType"].(string); ok {
		model.DeploymentType = types.StringValue(deploymentType)
	}

	if contactPoints, ok := response["contactPoints"].(string); ok {
		model.ContactPoints = types.StringValue(contactPoints)
	}

	if localDatacenter, ok := response["localDatacenter"].(string); ok {
		model.LocalDatacenter = types.StringValue(localDatacenter)
	}

	if port, ok := response["port"].(float64); ok {
		model.Port = types.Int64Value(int64(port))
	}

	if username, ok := response["username"].(string); ok {
		model.Username = types.StringValue(username)
	}

	// Password is write-only from the API perspective - the API returns "<Value is encrypted>"
	// We don't update the password field from the API response since it's not the actual value.
	// The password field should remain as whatever value was originally planned/configured.

	if cloudKind, ok := response["cloudKind"].(string); ok {
		model.CloudKind = types.StringValue(cloudKind)
	} else if model.CloudKind.IsUnknown() {
		model.CloudKind = types.StringNull()
	}

	if databaseId, ok := response["databaseId"].(string); ok {
		model.DatabaseId = types.StringValue(databaseId)
	} else if model.DatabaseId.IsUnknown() {
		model.DatabaseId = types.StringNull()
	}

	if region, ok := response["region"].(string); ok {
		model.Region = types.StringValue(region)
	} else if model.Region.IsUnknown() {
		model.Region = types.StringNull()
	}

	if sshTunnelId, ok := response["sshTunnelId"].(string); ok {
		model.SshTunnelId = types.StringValue(sshTunnelId)
	} else if model.SshTunnelId.IsUnknown() {
		model.SshTunnelId = types.StringNull()
	}

	if token, ok := response["token"].(string); ok {
		model.Token = types.StringValue(token)
	} else if model.Token.IsUnknown() {
		model.Token = types.StringNull()
	}

	// Handle validate field
	if validate, ok := response["validate"].(bool); ok {
		model.Validate = types.BoolValue(validate)
	} else if model.Validate.IsUnknown() {
		model.Validate = types.BoolNull()
	}
}
