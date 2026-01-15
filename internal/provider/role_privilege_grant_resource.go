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

	// Role privilege grants are uniquely identified by (roleId, entityId, privilege, grantKind)
	// The API doesn't provide a way to read individual grants, so we assume it exists if it's in state
	// The grant will be validated during the next Update or when it's explicitly deleted
	roleId := state.RoleId.ValueString()
	entityId := state.EntityId.ValueString()
	privilege := state.Privilege.ValueString()
	grantKind := state.GrantKind.ValueString()

	tflog.Debug(ctx, "Reading role_privilege_grant (no-op - assumes grant exists)", map[string]interface{}{
		"roleId":    roleId,
		"entityId":  entityId,
		"privilege": privilege,
		"grantKind": grantKind,
	})

	// No API call needed - the grant is managed through Create/Update/Delete operations
	// The state remains unchanged
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

	// Role privilege grants don't support direct updates (API returns 405)
	// Instead, we delete the old grant and create a new one
	// First, delete the existing grant using the composite key
	roleId := state.RoleId.ValueString()
	entityId := state.EntityId.ValueString()
	privilege := state.Privilege.ValueString()
	grantKind := state.GrantKind.ValueString()

	tflog.Debug(ctx, "Updating role_privilege_grant via delete+create", map[string]interface{}{
		"roleId":    roleId,
		"entityId":  entityId,
		"privilege": privilege,
		"grantKind": grantKind,
	})

	// Create the revoke request with the unique identifying fields
	// Grants are uniquely identified by (roleId, entityId, privilege, grantKind)
	entityKind := state.EntityKind.ValueString()
	deleteRequest := make(map[string]interface{})
	deleteRequest["entityId"] = entityId
	deleteRequest["entityKind"] = entityKind
	deleteRequest["privilege"] = privilege
	// Use RemoveRoleGrant to completely remove the grant (not just the grant option)
	deleteRequest["revokeAction"] = "RemoveRoleGrant"

	// Include optional scope fields if set
	if !state.ColumnName.IsNull() && state.ColumnName.ValueString() != "" {
		deleteRequest["columnName"] = state.ColumnName.ValueString()
	}
	if !state.SchemaName.IsNull() && state.SchemaName.ValueString() != "" {
		deleteRequest["schemaName"] = state.SchemaName.ValueString()
	}
	if !state.TableName.IsNull() && state.TableName.ValueString() != "" {
		deleteRequest["tableName"] = state.TableName.ValueString()
	}

	// Delete the old grant
	tflog.Debug(ctx, "Deleting existing role_privilege_grant before update")
	err := r.client.RevokeRolePrivilege(ctx, roleId, deleteRequest)
	if err != nil && !client.IsNotFound(err) {
		resp.Diagnostics.AddError(
			"Error deleting role_privilege_grant during update",
			"Could not delete existing grant: "+err.Error(),
		)
		return
	}

	// Create the new grant with updated values
	createRequest := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating new role_privilege_grant with updated values")
	response, err := r.client.CreateRolePrivilegeGrant(ctx, createRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating role_privilege_grant during update",
			"Could not create grant with new values: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated role_privilege_grant via delete+create")
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *role_privilege_grantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_role_privilege_grant.RolePrivilegeGrantModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the revoke request using the same approach as Update
	roleId := state.RoleId.ValueString()
	entityId := state.EntityId.ValueString()
	entityKind := state.EntityKind.ValueString()
	privilege := state.Privilege.ValueString()

	tflog.Debug(ctx, "Deleting role_privilege_grant", map[string]interface{}{
		"roleId":    roleId,
		"entityId":  entityId,
		"privilege": privilege,
	})

	// Create the revoke request
	revokeRequest := make(map[string]interface{})
	revokeRequest["entityId"] = entityId
	revokeRequest["entityKind"] = entityKind
	revokeRequest["privilege"] = privilege
	revokeRequest["revokeAction"] = "RemoveRoleGrant"

	// Include optional scope fields if set
	if !state.ColumnName.IsNull() && state.ColumnName.ValueString() != "" {
		revokeRequest["columnName"] = state.ColumnName.ValueString()
	}
	if !state.SchemaName.IsNull() && state.SchemaName.ValueString() != "" {
		revokeRequest["schemaName"] = state.SchemaName.ValueString()
	}
	if !state.TableName.IsNull() && state.TableName.ValueString() != "" {
		revokeRequest["tableName"] = state.TableName.ValueString()
	}

	err := r.client.RevokeRolePrivilege(ctx, roleId, revokeRequest)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting role_privilege_grant",
				"Could not delete role_privilege_grant: "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted role_privilege_grant", map[string]interface{}{
		"roleId":    roleId,
		"entityId":  entityId,
		"privilege": privilege,
	})
}

// Helper methods
func (r *role_privilege_grantResource) modelToCreateRequest(ctx context.Context, model *resource_role_privilege_grant.RolePrivilegeGrantModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	if !model.RoleId.IsNull() && !model.RoleId.IsUnknown() && model.RoleId.ValueString() != "" {
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
	if !model.ColumnName.IsNull() && !model.ColumnName.IsUnknown() && model.ColumnName.ValueString() != "" {
		request["columnName"] = model.ColumnName.ValueString()
	}
	if !model.SchemaName.IsNull() && !model.SchemaName.IsUnknown() && model.SchemaName.ValueString() != "" {
		request["schemaName"] = model.SchemaName.ValueString()
	}
	if !model.TableName.IsNull() && !model.TableName.IsUnknown() && model.TableName.ValueString() != "" {
		request["tableName"] = model.TableName.ValueString()
	}

	// Note: Pagination fields are handled at the framework level and not part of the model
	if !model.ListAllPrivileges.IsNull() && !model.ListAllPrivileges.IsUnknown() {
		request["listAllPrivileges"] = model.ListAllPrivileges.ValueBool()
	}

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

	// For optional scope fields (columnName, schemaName, tableName), preserve plan values
	// when the API doesn't return them. This handles the case where users specify wildcard
	// values like "*" that the API accepts but doesn't echo back in the response.
	// If the plan value is null/unknown, set to null (must be known after apply).
	if columnName, ok := response["columnName"].(string); ok {
		model.ColumnName = types.StringValue(columnName)
	} else if model.ColumnName.IsNull() || model.ColumnName.IsUnknown() {
		model.ColumnName = types.StringNull()
	}
	// Otherwise keep existing model value (user-specified value like "*")

	if schemaName, ok := response["schemaName"].(string); ok {
		model.SchemaName = types.StringValue(schemaName)
	} else if model.SchemaName.IsNull() || model.SchemaName.IsUnknown() {
		model.SchemaName = types.StringNull()
	}
	// Otherwise keep existing model value (user-specified value like "*")

	if tableName, ok := response["tableName"].(string); ok {
		model.TableName = types.StringValue(tableName)
	} else if model.TableName.IsNull() || model.TableName.IsUnknown() {
		model.TableName = types.StringNull()
	}
	// Otherwise keep existing model value (user-specified value like "*")

	// Note: Pagination fields are not part of the model as they are handled at the framework level

	if listAllPrivileges, ok := response["listAllPrivileges"].(bool); ok {
		model.ListAllPrivileges = types.BoolValue(listAllPrivileges)
	} else if model.ListAllPrivileges.IsNull() || model.ListAllPrivileges.IsUnknown() {
		model.ListAllPrivileges = types.BoolNull()
	}
	// Otherwise keep existing model value

	// Note: Role privilege grants are individual operations, not list operations
}
