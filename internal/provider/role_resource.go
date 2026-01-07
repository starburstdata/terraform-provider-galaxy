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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_role"
)

var _ resource.Resource = (*roleResource)(nil)
var _ resource.ResourceWithConfigure = (*roleResource)(nil)

func NewRoleResource() resource.Resource {
	return &roleResource{}
}

type roleResource struct {
	client *client.GalaxyClient
}

func (r *roleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_role.RoleResourceSchema(ctx)
}

func (r *roleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_role.RoleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating role")
	response, err := r.client.CreateRole(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating role",
			"Could not create role: "+err.Error(),
		)
		return
	}

	// Set defaults for computed fields that were unknown before creation
	// to ensure they are not left as unknown after apply
	if plan.RoleDescription.IsUnknown() {
		plan.RoleDescription = types.StringNull()
	}
	if plan.CreatedOn.IsUnknown() {
		plan.CreatedOn = types.StringNull()
	}
	if plan.ModifiedOn.IsUnknown() {
		plan.ModifiedOn = types.StringNull()
	}
	if plan.OwningRoleId.IsUnknown() {
		plan.OwningRoleId = types.StringNull()
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created role", map[string]interface{}{"id": plan.RoleId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_role.RoleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.RoleId.ValueString()
	tflog.Debug(ctx, "Reading role", map[string]interface{}{"id": id})
	response, err := r.client.GetRole(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Role not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading role",
			"Could not read role "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_role.RoleModel
	var state resource_role.RoleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.RoleId.ValueString()
	request := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating role", map[string]interface{}{"id": id})
	response, err := r.client.UpdateRole(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating role",
			"Could not update role "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated role", map[string]interface{}{"id": plan.RoleId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_role.RoleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.RoleId.ValueString()
	tflog.Debug(ctx, "Deleting role", map[string]interface{}{"id": id})
	err := r.client.DeleteRole(ctx, id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting role",
				"Could not delete role "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted role", map[string]interface{}{"id": id})
}

// Helper methods
func (r *roleResource) modelToCreateRequest(ctx context.Context, model *resource_role.RoleModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Map required fields
	if !model.RoleName.IsNull() && !model.RoleName.IsUnknown() && model.RoleName.ValueString() != "" {
		request["roleName"] = model.RoleName.ValueString()
	}

	if !model.RoleDescription.IsNull() && !model.RoleDescription.IsUnknown() && model.RoleDescription.ValueString() != "" {
		request["roleDescription"] = model.RoleDescription.ValueString()
	}

	// Handle the boolean field properly - use value instead of pointer
	if !model.GrantToCreatingRole.IsNull() && !model.GrantToCreatingRole.IsUnknown() {
		request["grantToCreatingRole"] = model.GrantToCreatingRole.ValueBool()
	}

	return request
}

func (r *roleResource) modelToUpdateRequest(ctx context.Context, model *resource_role.RoleModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	// Remove computed-only fields that should not be included in updates
	delete(request, "roleId")
	delete(request, "createdOn")
	delete(request, "modifiedOn")
	delete(request, "owningRoleId")
	delete(request, "allRoles")
	delete(request, "directlyGrantedRoles")

	return request
}

func (r *roleResource) updateModelFromResponse(ctx context.Context, model *resource_role.RoleModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
	if id, ok := response["roleId"].(string); ok {
		model.RoleId = types.StringValue(id)
	}

	if roleName, ok := response["roleName"].(string); ok {
		model.RoleName = types.StringValue(roleName)
	}

	if roleDescription, ok := response["roleDescription"].(string); ok {
		model.RoleDescription = types.StringValue(roleDescription)
	} else if model.RoleDescription.IsUnknown() {
		model.RoleDescription = types.StringNull()
	}

	if grantToCreatingRole, ok := response["grantToCreatingRole"].(bool); ok {
		model.GrantToCreatingRole = types.BoolValue(grantToCreatingRole)
	}

	if createdOn, ok := response["createdOn"].(string); ok {
		model.CreatedOn = types.StringValue(createdOn)
	} else if model.CreatedOn.IsUnknown() {
		model.CreatedOn = types.StringNull()
	}

	if modifiedOn, ok := response["modifiedOn"].(string); ok {
		model.ModifiedOn = types.StringValue(modifiedOn)
	} else if model.ModifiedOn.IsUnknown() {
		model.ModifiedOn = types.StringNull()
	}

	if owningRoleId, ok := response["owningRoleId"].(string); ok {
		model.OwningRoleId = types.StringValue(owningRoleId)
	} else if model.OwningRoleId.IsUnknown() {
		model.OwningRoleId = types.StringNull()
	}

	// Set complex list fields - these are computed fields that may not be returned by Create API
	// Set to empty lists to indicate they are known but empty
	if allRoles, ok := response["allRoles"].([]interface{}); ok {
		rolesList := make([]resource_role.AllRolesValue, 0, len(allRoles))
		for _, role := range allRoles {
			if roleMap, ok := role.(map[string]interface{}); ok {
				roleValue := resource_role.AllRolesValue{}

				if adminOption, ok := roleMap["adminOption"].(bool); ok {
					roleValue.AdminOption = types.BoolValue(adminOption)
				} else {
					roleValue.AdminOption = types.BoolNull()
				}

				if roleId, ok := roleMap["roleId"].(string); ok {
					roleValue.RoleId = types.StringValue(roleId)
				} else {
					roleValue.RoleId = types.StringNull()
				}

				if roleName, ok := roleMap["roleName"].(string); ok {
					roleValue.RoleName = types.StringValue(roleName)
				} else {
					roleValue.RoleName = types.StringNull()
				}

				// Set principal as null for now (complex nested object)
				roleValue.Principal = types.ObjectNull(resource_role.PrincipalValue{}.AttributeTypes(ctx))

				rolesList = append(rolesList, roleValue)
			}
		}
		allRolesType := resource_role.AllRolesType{
			ObjectType: types.ObjectType{
				AttrTypes: resource_role.AllRolesValue{}.AttributeTypes(ctx),
			},
		}
		listValue, d := types.ListValueFrom(ctx, allRolesType, rolesList)
		diags.Append(d...)
		model.AllRoles = listValue
	} else {
		allRolesType := resource_role.AllRolesType{
			ObjectType: types.ObjectType{
				AttrTypes: resource_role.AllRolesValue{}.AttributeTypes(ctx),
			},
		}
		listValue, d := types.ListValueFrom(ctx, allRolesType, []resource_role.AllRolesValue{})
		diags.Append(d...)
		model.AllRoles = listValue
	}

	if directlyGrantedRoles, ok := response["directlyGrantedRoles"].([]interface{}); ok {
		rolesList := make([]resource_role.DirectlyGrantedRolesValue, 0, len(directlyGrantedRoles))
		for _, role := range directlyGrantedRoles {
			if roleMap, ok := role.(map[string]interface{}); ok {
				roleValue := resource_role.DirectlyGrantedRolesValue{}

				if adminOption, ok := roleMap["adminOption"].(bool); ok {
					roleValue.AdminOption = types.BoolValue(adminOption)
				} else {
					roleValue.AdminOption = types.BoolNull()
				}

				if roleId, ok := roleMap["roleId"].(string); ok {
					roleValue.RoleId = types.StringValue(roleId)
				} else {
					roleValue.RoleId = types.StringNull()
				}

				if roleName, ok := roleMap["roleName"].(string); ok {
					roleValue.RoleName = types.StringValue(roleName)
				} else {
					roleValue.RoleName = types.StringNull()
				}

				// Set principal as null for now (complex nested object)
				roleValue.Principal = types.ObjectNull(resource_role.PrincipalValue{}.AttributeTypes(ctx))

				rolesList = append(rolesList, roleValue)
			}
		}
		directlyGrantedRolesType := resource_role.DirectlyGrantedRolesType{
			ObjectType: types.ObjectType{
				AttrTypes: resource_role.DirectlyGrantedRolesValue{}.AttributeTypes(ctx),
			},
		}
		listValue, d := types.ListValueFrom(ctx, directlyGrantedRolesType, rolesList)
		diags.Append(d...)
		model.DirectlyGrantedRoles = listValue
	} else {
		directlyGrantedRolesType := resource_role.DirectlyGrantedRolesType{
			ObjectType: types.ObjectType{
				AttrTypes: resource_role.DirectlyGrantedRolesValue{}.AttributeTypes(ctx),
			},
		}
		listValue, d := types.ListValueFrom(ctx, directlyGrantedRolesType, []resource_role.DirectlyGrantedRolesValue{})
		diags.Append(d...)
		model.DirectlyGrantedRoles = listValue
	}
}
