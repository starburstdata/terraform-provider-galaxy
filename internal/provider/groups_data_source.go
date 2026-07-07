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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_groups"
)

var _ datasource.DataSource = (*groupsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*groupsDataSource)(nil)

func NewGroupsDataSource() datasource.DataSource {
	return &groupsDataSource{}
}

type groupsDataSource struct {
	client *client.GalaxyClient
}

func (d *groupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *groupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_groups.GroupsDataSourceSchema(ctx)
}

func (d *groupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *groupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_groups.GroupsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading groups with automatic pagination")

	allGroups, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/group")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading groups",
			"Could not read groups: "+err.Error(),
		)
		return
	}

	var groupMaps []map[string]interface{}
	for _, groupInterface := range allGroups {
		if groupMap, ok := groupInterface.(map[string]interface{}); ok {
			groupMaps = append(groupMaps, groupMap)
		}
	}

	if len(groupMaps) > 0 {
		groups, err := d.mapGroupsResult(ctx, groupMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping groups response",
				"Could not map groups response: "+err.Error(),
			)
			return
		}
		config.Result = groups
	} else {
		elementType := datasource_groups.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_groups.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_groups.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *groupsDataSource) mapGroupsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	groups := make([]datasource_groups.ResultValue, 0)

	for _, groupMap := range result {
		group := d.mapSingleGroup(ctx, groupMap)
		groups = append(groups, group)
	}

	elementType := datasource_groups.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_groups.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, groups)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *groupsDataSource) mapSingleGroup(ctx context.Context, groupMap map[string]interface{}) datasource_groups.ResultValue {
	attributeTypes := datasource_groups.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	if createdOn, ok := groupMap["createdOn"].(string); ok {
		attributes["created_on"] = types.StringValue(createdOn)
	} else {
		attributes["created_on"] = types.StringNull()
	}

	if externalId, ok := groupMap["externalId"].(string); ok {
		attributes["external_id"] = types.StringValue(externalId)
	} else {
		attributes["external_id"] = types.StringNull()
	}

	if groupId, ok := groupMap["groupId"].(string); ok {
		attributes["group_id"] = types.StringValue(groupId)
	} else {
		attributes["group_id"] = types.StringNull()
	}

	if groupName, ok := groupMap["groupName"].(string); ok {
		attributes["group_name"] = types.StringValue(groupName)
	} else {
		attributes["group_name"] = types.StringNull()
	}

	if modifiedOn, ok := groupMap["modifiedOn"].(string); ok {
		attributes["modified_on"] = types.StringValue(modifiedOn)
	} else {
		attributes["modified_on"] = types.StringNull()
	}

	attributes["roles"] = d.mapRolesList(ctx, groupMap)
	attributes["users"] = d.mapUsersList(ctx, groupMap)

	group, diags := datasource_groups.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating group ResultValue: %v", diags))
		return datasource_groups.NewResultValueNull()
	}

	return group
}

func (d *groupsDataSource) mapRolesList(ctx context.Context, groupMap map[string]interface{}) types.List {
	elementType := datasource_groups.RolesType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_groups.RolesValue{}.AttributeTypes(ctx),
		},
	}
	if roles, ok := groupMap["roles"].([]interface{}); ok {
		rolesList := make([]datasource_groups.RolesValue, 0, len(roles))
		for _, roleInterface := range roles {
			if rm, ok := roleInterface.(map[string]interface{}); ok {
				roleValue := datasource_groups.RolesValue{
					RoleId:   types.StringNull(),
					RoleName: types.StringNull(),
				}
				if roleId, ok := rm["roleId"].(string); ok {
					roleValue.RoleId = types.StringValue(roleId)
				}
				if roleName, ok := rm["roleName"].(string); ok {
					roleValue.RoleName = types.StringValue(roleName)
				}
				rolesList = append(rolesList, roleValue)
			}
		}
		listValue, d := types.ListValueFrom(ctx, elementType, rolesList)
		if d.HasError() {
			tflog.Error(ctx, fmt.Sprintf("Error creating roles list: %v", d))
		} else {
			return listValue
		}
	}
	emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_groups.RolesValue{})
	return emptyList
}

func (d *groupsDataSource) mapUsersList(ctx context.Context, groupMap map[string]interface{}) types.List {
	elementType := datasource_groups.UsersType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_groups.UsersValue{}.AttributeTypes(ctx),
		},
	}
	if users, ok := groupMap["users"].([]interface{}); ok {
		usersList := make([]datasource_groups.UsersValue, 0, len(users))
		for _, userInterface := range users {
			if um, ok := userInterface.(map[string]interface{}); ok {
				userValue := datasource_groups.UsersValue{
					Email:  types.StringNull(),
					UserId: types.StringNull(),
				}
				if email, ok := um["email"].(string); ok {
					userValue.Email = types.StringValue(email)
				}
				if userId, ok := um["userId"].(string); ok {
					userValue.UserId = types.StringValue(userId)
				}
				usersList = append(usersList, userValue)
			}
		}
		listValue, d := types.ListValueFrom(ctx, elementType, usersList)
		if d.HasError() {
			tflog.Error(ctx, fmt.Sprintf("Error creating users list: %v", d))
		} else {
			return listValue
		}
	}
	emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_groups.UsersValue{})
	return emptyList
}
