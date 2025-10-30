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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_role"
)

var _ datasource.DataSource = (*roleDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*roleDataSource)(nil)

func NewRoleDataSource() datasource.DataSource {
	return &roleDataSource{}
}

type roleDataSource struct {
	client *client.GalaxyClient
}

func (d *roleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *roleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_role.RoleDataSourceSchema(ctx)
}

func (d *roleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.GalaxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.GalaxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *roleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_role.RoleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.RoleId.ValueString()
	tflog.Debug(ctx, "Reading role", map[string]interface{}{"id": id})

	response, err := d.client.GetRole(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role",
			"Could not read role "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *roleDataSource) updateModelFromResponse(ctx context.Context, model *datasource_role.RoleModel, response map[string]interface{}) {
	// Map response fields to model based on actual API response structure
	if roleId, ok := response["roleId"].(string); ok {
		model.RoleId = types.StringValue(roleId)
	}
	if roleName, ok := response["roleName"].(string); ok {
		model.RoleName = types.StringValue(roleName)
	}
	if roleDescription, ok := response["roleDescription"].(string); ok {
		model.RoleDescription = types.StringValue(roleDescription)
	}
	if createdOn, ok := response["createdOn"].(string); ok {
		model.CreatedOn = types.StringValue(createdOn)
	}
	if modifiedOn, ok := response["modifiedOn"].(string); ok {
		model.ModifiedOn = types.StringValue(modifiedOn)
	}
	if owningRoleId, ok := response["owningRoleId"].(string); ok {
		model.OwningRoleId = types.StringValue(owningRoleId)
	}

	// Map role lists (reusing the same structure as user data source)
	d.mapRolesToModel(ctx, model, response)
}

func (d *roleDataSource) mapRolesToModel(ctx context.Context, model *datasource_role.RoleModel, response map[string]interface{}) {
	// Map all roles
	if allRoles, ok := response["allRoles"].([]interface{}); ok {
		allRolesList := make([]datasource_role.AllRolesValue, 0, len(allRoles))
		for _, roleInterface := range allRoles {
			if roleMap, ok := roleInterface.(map[string]interface{}); ok {
				role := d.mapSingleRole(ctx, roleMap)
				allRolesList = append(allRolesList, role)
			}
		}
		allRolesType := datasource_role.AllRolesType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_role.AllRolesValue{}.AttributeTypes(ctx),
			},
		}
		allRolesListValue, _ := types.ListValueFrom(ctx, allRolesType, allRolesList)
		model.AllRoles = allRolesListValue
	} else {
		allRolesType := datasource_role.AllRolesType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_role.AllRolesValue{}.AttributeTypes(ctx),
			},
		}
		model.AllRoles = types.ListNull(allRolesType)
	}

	// Map directly granted roles
	if directlyGrantedRoles, ok := response["directlyGrantedRoles"].([]interface{}); ok {
		directRolesList := make([]datasource_role.DirectlyGrantedRolesValue, 0, len(directlyGrantedRoles))
		for _, roleInterface := range directlyGrantedRoles {
			if roleMap, ok := roleInterface.(map[string]interface{}); ok {
				role := d.mapSingleDirectRole(ctx, roleMap)
				directRolesList = append(directRolesList, role)
			}
		}
		directRolesType := datasource_role.DirectlyGrantedRolesType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_role.DirectlyGrantedRolesValue{}.AttributeTypes(ctx),
			},
		}
		directRolesListValue, _ := types.ListValueFrom(ctx, directRolesType, directRolesList)
		model.DirectlyGrantedRoles = directRolesListValue
	} else {
		directRolesType := datasource_role.DirectlyGrantedRolesType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_role.DirectlyGrantedRolesValue{}.AttributeTypes(ctx),
			},
		}
		model.DirectlyGrantedRoles = types.ListNull(directRolesType)
	}
}

func (d *roleDataSource) mapSingleRole(ctx context.Context, roleMap map[string]interface{}) datasource_role.AllRolesValue {
	// Initialize with null/unknown values using proper basetypes
	roleValue := datasource_role.AllRolesValue{
		RoleId:      types.StringNull(),
		RoleName:    types.StringNull(),
		AdminOption: types.BoolNull(),
		Principal:   types.ObjectNull(datasource_role.PrincipalValue{}.AttributeTypes(ctx)),
	}

	if roleId, ok := roleMap["roleId"].(string); ok {
		roleValue.RoleId = types.StringValue(roleId)
	}
	if roleName, ok := roleMap["roleName"].(string); ok {
		roleValue.RoleName = types.StringValue(roleName)
	}
	if adminOption, ok := roleMap["adminOption"].(bool); ok {
		roleValue.AdminOption = types.BoolValue(adminOption)
	}

	// Map principal
	if principal, ok := roleMap["principal"].(map[string]interface{}); ok {
		principalValue := datasource_role.PrincipalValue{
			Id:            types.StringNull(),
			PrincipalType: types.StringNull(),
		}
		if principalId, ok := principal["id"].(string); ok {
			principalValue.Id = types.StringValue(principalId)
		}
		if principalType, ok := principal["type"].(string); ok {
			principalValue.PrincipalType = types.StringValue(principalType)
		}
		principalObjectValue, diags := principalValue.ToObjectValue(ctx)
		if !diags.HasError() {
			roleValue.Principal = principalObjectValue
		}
	}

	return roleValue
}

func (d *roleDataSource) mapSingleDirectRole(ctx context.Context, roleMap map[string]interface{}) datasource_role.DirectlyGrantedRolesValue {
	// Initialize with null/unknown values
	roleValue := datasource_role.DirectlyGrantedRolesValue{
		RoleId:      types.StringNull(),
		RoleName:    types.StringNull(),
		AdminOption: types.BoolNull(),
		Principal:   types.ObjectNull(datasource_role.PrincipalValue{}.AttributeTypes(ctx)),
	}

	if roleId, ok := roleMap["roleId"].(string); ok {
		roleValue.RoleId = types.StringValue(roleId)
	}
	if roleName, ok := roleMap["roleName"].(string); ok {
		roleValue.RoleName = types.StringValue(roleName)
	}
	if adminOption, ok := roleMap["adminOption"].(bool); ok {
		roleValue.AdminOption = types.BoolValue(adminOption)
	}

	// Map principal
	if principal, ok := roleMap["principal"].(map[string]interface{}); ok {
		principalValue := datasource_role.PrincipalValue{
			Id:            types.StringNull(),
			PrincipalType: types.StringNull(),
		}
		if principalId, ok := principal["id"].(string); ok {
			principalValue.Id = types.StringValue(principalId)
		}
		if principalType, ok := principal["type"].(string); ok {
			principalValue.PrincipalType = types.StringValue(principalType)
		}
		principalObjectValue, diags := principalValue.ToObjectValue(ctx)
		if !diags.HasError() {
			roleValue.Principal = principalObjectValue
		}
	}

	return roleValue
}
