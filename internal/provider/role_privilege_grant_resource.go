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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_role_privilege_grant"
)

var _ resource.Resource = (*role_privilege_grantResource)(nil)
var _ resource.ResourceWithConfigure = (*role_privilege_grantResource)(nil)

func NewRolePrivilegeGrantResource() resource.Resource {
	return &role_privilege_grantResource{}
}

type role_privilege_grantResource struct {
	client *client.GalaxyClient
}

func (r *role_privilege_grantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_privilege_grant"
}

func (r *role_privilege_grantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_role_privilege_grant.RolePrivilegeGrantResourceSchema(ctx)
}

func (r *role_privilege_grantResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *role_privilege_grantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_role_privilege_grant.RolePrivilegeGrantModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating role_privilege_grant")
	response, err := r.client.CreateRolePrivilegeGrant(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating role_privilege_grant",
			"Could not create role_privilege_grant: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created role_privilege_grant", map[string]interface{}{"id": plan.EntityId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *role_privilege_grantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_role_privilege_grant.RolePrivilegeGrantModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.EntityId.ValueString()
	tflog.Debug(ctx, "Reading role_privilege_grant", map[string]interface{}{"id": id})
	response, err := r.client.GetRolePrivilegeGrant(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "RolePrivilegeGrant not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading role_privilege_grant",
			"Could not read role_privilege_grant "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *role_privilege_grantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_role_privilege_grant.RolePrivilegeGrantModel
	var state resource_role_privilege_grant.RolePrivilegeGrantModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.EntityId.ValueString()
	request := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating role_privilege_grant", map[string]interface{}{"id": id})
	response, err := r.client.UpdateRolePrivilegeGrant(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating role_privilege_grant",
			"Could not update role_privilege_grant "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated role_privilege_grant", map[string]interface{}{"id": plan.EntityId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *role_privilege_grantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_role_privilege_grant.RolePrivilegeGrantModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.EntityId.ValueString()
	tflog.Debug(ctx, "Deleting role_privilege_grant", map[string]interface{}{"id": id})
	err := r.client.DeleteRolePrivilegeGrant(ctx, id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting role_privilege_grant",
				"Could not delete role_privilege_grant "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted role_privilege_grant", map[string]interface{}{"id": id})
}

// Helper methods
func (r *role_privilege_grantResource) modelToCreateRequest(ctx context.Context, model *resource_role_privilege_grant.RolePrivilegeGrantModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	if !model.RoleId.IsNull() {
		request["roleId"] = model.RoleId.ValueString()
	}
	if !model.EntityId.IsNull() {
		request["entityId"] = model.EntityId.ValueString()
	}
	if !model.EntityKind.IsNull() {
		request["entityKind"] = model.EntityKind.ValueString()
	}
	if !model.GrantKind.IsNull() {
		request["grantKind"] = model.GrantKind.ValueString()
	}
	if !model.Privilege.IsNull() {
		request["privilege"] = model.Privilege.ValueString()
	}
	if !model.GrantOption.IsNull() {
		request["grantOption"] = model.GrantOption.ValueBool()
	}

	// Optional fields
	if !model.ColumnName.IsNull() {
		request["columnName"] = model.ColumnName.ValueString()
	}
	if !model.SchemaName.IsNull() {
		request["schemaName"] = model.SchemaName.ValueString()
	}
	if !model.TableName.IsNull() {
		request["tableName"] = model.TableName.ValueString()
	}

	// Note: Pagination fields are handled at the framework level and not part of the model
	if !model.ListAllPrivileges.IsNull() {
		request["listAllPrivileges"] = model.ListAllPrivileges.ValueBool()
	}

	return request
}

func (r *role_privilege_grantResource) modelToUpdateRequest(ctx context.Context, model *resource_role_privilege_grant.RolePrivilegeGrantModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	// Note: SyncToken not available in this model

	return request
}

func (r *role_privilege_grantResource) updateModelFromResponse(ctx context.Context, model *resource_role_privilege_grant.RolePrivilegeGrantModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// For privilege grants, we don't modify the EntityId since it's an input parameter
	// that should remain consistent (it represents the entity being granted privileges on)

	// Update other fields from response if they exist
	if roleId, ok := response["roleId"].(string); ok {
		model.RoleId = types.StringValue(roleId)
	}

	if entityId, ok := response["entityId"].(string); ok {
		model.EntityId = types.StringValue(entityId)
	}

	if entityKind, ok := response["entityKind"].(string); ok {
		model.EntityKind = types.StringValue(entityKind)
	}

	if grantKind, ok := response["grantKind"].(string); ok {
		model.GrantKind = types.StringValue(grantKind)
	}

	if privilege, ok := response["privilege"].(string); ok {
		model.Privilege = types.StringValue(privilege)
	}

	if grantOption, ok := response["grantOption"].(bool); ok {
		model.GrantOption = types.BoolValue(grantOption)
	}

	if columnName, ok := response["columnName"].(string); ok {
		model.ColumnName = types.StringValue(columnName)
	} else {
		model.ColumnName = types.StringNull()
	}

	if schemaName, ok := response["schemaName"].(string); ok {
		model.SchemaName = types.StringValue(schemaName)
	} else {
		model.SchemaName = types.StringNull()
	}

	if tableName, ok := response["tableName"].(string); ok {
		model.TableName = types.StringValue(tableName)
	} else {
		model.TableName = types.StringNull()
	}

	// Note: Pagination fields are not part of the model as they are handled at the framework level

	if listAllPrivileges, ok := response["listAllPrivileges"].(bool); ok {
		model.ListAllPrivileges = types.BoolValue(listAllPrivileges)
	} else {
		model.ListAllPrivileges = types.BoolNull()
	}

	// Note: Role privilege grants are individual operations, not list operations
}
