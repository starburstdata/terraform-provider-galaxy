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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_roles"
)

var _ datasource.DataSource = (*rolesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*rolesDataSource)(nil)

func NewRolesDataSource() datasource.DataSource {
	return &rolesDataSource{}
}

type rolesDataSource struct {
	client *client.GalaxyClient
}

func (d *rolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_roles"
}

func (d *rolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_roles.RolesDataSourceSchema(ctx)
}

func (d *rolesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *rolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_roles.RolesModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading roles with automatic pagination")

	// Use automatic pagination to get ALL roles across all pages
	allRoles, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/role")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading roles",
			"Could not read roles: "+err.Error(),
		)
		return
	}

	// Convert []interface{} to []map[string]interface{} for mapping
	var roleMaps []map[string]interface{}
	for _, roleInterface := range allRoles {
		if roleMap, ok := roleInterface.(map[string]interface{}); ok {
			roleMaps = append(roleMaps, roleMap)
		}
	}

	// Map API response to model
	if len(roleMaps) > 0 {
		roles, err := d.mapRolesResult(ctx, roleMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping roles response",
				"Could not map roles response: "+err.Error(),
			)
			return
		}
		config.Result = roles
	} else {
		elementType := datasource_roles.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_roles.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_roles.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *rolesDataSource) mapRolesResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	roles := make([]datasource_roles.ResultValue, 0)

	for _, roleMap := range result {
		role := d.mapSingleRole(ctx, roleMap)
		roles = append(roles, role)
	}

	elementType := datasource_roles.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_roles.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, roles)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *rolesDataSource) mapSingleRole(ctx context.Context, roleMap map[string]interface{}) datasource_roles.ResultValue {
	attributeTypes := datasource_roles.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map role ID
	if roleId, ok := roleMap["roleId"].(string); ok {
		attributes["role_id"] = types.StringValue(roleId)
	} else {
		attributes["role_id"] = types.StringNull()
	}

	// Map role name
	if roleName, ok := roleMap["roleName"].(string); ok {
		attributes["role_name"] = types.StringValue(roleName)
	} else if name, ok := roleMap["name"].(string); ok {
		attributes["role_name"] = types.StringValue(name)
	} else {
		attributes["role_name"] = types.StringNull()
	}

	// Map role description
	if roleDescription, ok := roleMap["roleDescription"].(string); ok {
		attributes["role_description"] = types.StringValue(roleDescription)
	} else if description, ok := roleMap["description"].(string); ok {
		attributes["role_description"] = types.StringValue(description)
	} else {
		attributes["role_description"] = types.StringNull()
	}

	// Map owning role ID
	if owningRoleId, ok := roleMap["owningRoleId"].(string); ok {
		attributes["owning_role_id"] = types.StringValue(owningRoleId)
	} else {
		attributes["owning_role_id"] = types.StringNull()
	}

	// Map created on
	if createdOn, ok := roleMap["createdOn"].(string); ok {
		attributes["created_on"] = types.StringValue(createdOn)
	} else {
		attributes["created_on"] = types.StringNull()
	}

	// Map modified on
	if modifiedOn, ok := roleMap["modifiedOn"].(string); ok {
		attributes["modified_on"] = types.StringValue(modifiedOn)
	} else {
		attributes["modified_on"] = types.StringNull()
	}

	// Map all roles list
	attributes["all_roles"] = d.mapAllRolesList(roleMap)

	// Map directly granted roles list
	attributes["directly_granted_roles"] = d.mapDirectlyGrantedRolesList(roleMap)

	// Create the ResultValue using the constructor
	role, diags := datasource_roles.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating role ResultValue: %v", diags))
		return datasource_roles.NewResultValueNull()
	}

	return role
}

func (d *rolesDataSource) mapAllRolesList(roleMap map[string]interface{}) types.List {
	if allRoles, ok := roleMap["allRoles"].([]interface{}); ok {
		allRolesList := make([]attr.Value, 0, len(allRoles))
		for _, roleIdInterface := range allRoles {
			if roleIdStr, ok := roleIdInterface.(string); ok {
				allRolesList = append(allRolesList, types.StringValue(roleIdStr))
			}
		}
		allRolesListValue, _ := types.ListValue(types.StringType, allRolesList)
		return allRolesListValue
	} else {
		return types.ListNull(types.StringType)
	}
}

func (d *rolesDataSource) mapDirectlyGrantedRolesList(roleMap map[string]interface{}) types.List {
	if directlyGrantedRoles, ok := roleMap["directlyGrantedRoles"].([]interface{}); ok {
		directlyGrantedRolesList := make([]attr.Value, 0, len(directlyGrantedRoles))
		for _, roleIdInterface := range directlyGrantedRoles {
			if roleIdStr, ok := roleIdInterface.(string); ok {
				directlyGrantedRolesList = append(directlyGrantedRolesList, types.StringValue(roleIdStr))
			}
		}
		directlyGrantedRolesListValue, _ := types.ListValue(types.StringType, directlyGrantedRolesList)
		return directlyGrantedRolesListValue
	} else {
		return types.ListNull(types.StringType)
	}
}
