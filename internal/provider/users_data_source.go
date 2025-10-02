package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_users"
)

var _ datasource.DataSource = (*usersDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*usersDataSource)(nil)

func NewUsersDataSource() datasource.DataSource {
	return &usersDataSource{}
}

type usersDataSource struct {
	client *client.GalaxyClient
}

func (d *usersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_users.UsersDataSourceSchema(ctx)
}

func (d *usersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_users.UsersModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading users with automatic pagination")

	// Use automatic pagination to get ALL users across all pages
	allUsers, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/user")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading users",
			"Could not read users: "+err.Error(),
		)
		return
	}

	// Convert []interface{} to []map[string]interface{} for mapping
	var userMaps []map[string]interface{}
	for _, userInterface := range allUsers {
		if userMap, ok := userInterface.(map[string]interface{}); ok {
			userMaps = append(userMaps, userMap)
		}
	}

	// Map API response to model
	if len(userMaps) > 0 {
		users, err := d.mapUsersResult(ctx, userMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping users response",
				"Could not map users response: "+err.Error(),
			)
			return
		}
		config.Result = users
	} else {
		elementType := datasource_users.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_users.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_users.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *usersDataSource) mapUsersResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	users := make([]datasource_users.ResultValue, 0, len(result))

	for _, userMap := range result {
		user := d.mapSingleUser(ctx, userMap)
		users = append(users, user)
	}

	elementType := datasource_users.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_users.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, users)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *usersDataSource) mapSingleUser(ctx context.Context, userMap map[string]interface{}) datasource_users.ResultValue {
	// Use the generated AttributeTypes method
	attributeTypes := datasource_users.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map basic user fields
	if userId, ok := userMap["userId"].(string); ok {
		attributes["user_id"] = types.StringValue(userId)
	} else {
		attributes["user_id"] = types.StringNull()
	}

	if email, ok := userMap["email"].(string); ok {
		attributes["email"] = types.StringValue(email)
	} else {
		attributes["email"] = types.StringNull()
	}

	if createdOn, ok := userMap["createdOn"].(string); ok {
		attributes["created_on"] = types.StringValue(createdOn)
	} else {
		attributes["created_on"] = types.StringNull()
	}

	if defaultRoleId, ok := userMap["defaultRoleId"].(string); ok {
		attributes["default_role_id"] = types.StringValue(defaultRoleId)
	} else {
		attributes["default_role_id"] = types.StringNull()
	}

	if scimManaged, ok := userMap["scimManaged"].(bool); ok {
		attributes["scim_managed"] = types.BoolValue(scimManaged)
	} else {
		attributes["scim_managed"] = types.BoolNull()
	}

	// Map role lists - simplified to null for now to get basic functionality working
	allRolesElementType := datasource_users.AllRolesType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_users.AllRolesValue{}.AttributeTypes(ctx),
		},
	}
	attributes["all_roles"] = types.ListNull(allRolesElementType)

	directRolesElementType := datasource_users.DirectlyGrantedRolesType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_users.DirectlyGrantedRolesValue{}.AttributeTypes(ctx),
		},
	}
	attributes["directly_granted_roles"] = types.ListNull(directRolesElementType)

	// Create the ResultValue using the constructor
	user, diags := datasource_users.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating user ResultValue: %v", diags))
		return datasource_users.NewResultValueNull()
	}

	return user
}
