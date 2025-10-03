package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_user"
)

var _ datasource.DataSource = (*userDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*userDataSource)(nil)

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

type userDataSource struct {
	client *client.GalaxyClient
}

func (d *userDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_user.UserDataSourceSchema(ctx)
}

func (d *userDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_user.UserModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading user", map[string]interface{}{"id": id})

	response, err := d.client.GetUser(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading user",
			"Could not read user "+id+": "+err.Error(),
		)
		return
	}

	// Debug logging
	tflog.Debug(ctx, "API response received", map[string]interface{}{"response": response})

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *userDataSource) updateModelFromResponse(ctx context.Context, model *datasource_user.UserModel, response map[string]interface{}) {
	// Map response fields to model
	if userId, ok := response["userId"].(string); ok {
		model.Id = types.StringValue(userId)
		model.UserId = types.StringValue(userId)
	}
	if email, ok := response["email"].(string); ok {
		model.Email = types.StringValue(email)
	}
	if createdOn, ok := response["createdOn"].(string); ok {
		model.CreatedOn = types.StringValue(createdOn)
	}
	if defaultRoleId, ok := response["defaultRoleId"].(string); ok {
		model.DefaultRoleId = types.StringValue(defaultRoleId)
	}
	if scimManaged, ok := response["scimManaged"].(bool); ok {
		model.ScimManaged = types.BoolValue(scimManaged)
	}

	// Map role lists
	d.mapRolesToModel(ctx, model, response)
}

func (d *userDataSource) mapRolesToModel(ctx context.Context, model *datasource_user.UserModel, response map[string]interface{}) {
	// Map all roles
	if allRoles, ok := response["allRoles"].([]interface{}); ok {
		if len(allRoles) > 0 {
			// Convert roles to the expected structure
			var roleElements []attr.Value
			for _, roleInterface := range allRoles {
				if roleMap, ok := roleInterface.(map[string]interface{}); ok {
					role := d.mapSingleRole(ctx, roleMap)
					roleElements = append(roleElements, role)
				}
			}
			elementType := datasource_user.AllRolesType{
				ObjectType: types.ObjectType{
					AttrTypes: datasource_user.AllRolesValue{}.AttributeTypes(ctx),
				},
			}
			model.AllRoles, _ = types.ListValue(elementType, roleElements)
		} else {
			// Empty roles list with proper type
			elementType := datasource_user.AllRolesType{
				ObjectType: types.ObjectType{
					AttrTypes: datasource_user.AllRolesValue{}.AttributeTypes(ctx),
				},
			}
			model.AllRoles, _ = types.ListValue(elementType, []attr.Value{})
		}
	} else {
		elementType := datasource_user.AllRolesType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_user.AllRolesValue{}.AttributeTypes(ctx),
			},
		}
		model.AllRoles, _ = types.ListValue(elementType, []attr.Value{})
	}

	// Map directly granted roles
	if directlyGrantedRoles, ok := response["directlyGrantedRoles"].([]interface{}); ok {
		if len(directlyGrantedRoles) > 0 {
			// Convert roles to the expected structure
			var directRoleElements []attr.Value
			for _, roleInterface := range directlyGrantedRoles {
				if roleMap, ok := roleInterface.(map[string]interface{}); ok {
					role := d.mapSingleDirectRole(ctx, roleMap)
					directRoleElements = append(directRoleElements, role)
				}
			}
			elementType := datasource_user.DirectlyGrantedRolesType{
				ObjectType: types.ObjectType{
					AttrTypes: datasource_user.DirectlyGrantedRolesValue{}.AttributeTypes(ctx),
				},
			}
			model.DirectlyGrantedRoles, _ = types.ListValue(elementType, directRoleElements)
		} else {
			// Empty roles list with proper type
			elementType := datasource_user.DirectlyGrantedRolesType{
				ObjectType: types.ObjectType{
					AttrTypes: datasource_user.DirectlyGrantedRolesValue{}.AttributeTypes(ctx),
				},
			}
			model.DirectlyGrantedRoles, _ = types.ListValue(elementType, []attr.Value{})
		}
	} else {
		elementType := datasource_user.DirectlyGrantedRolesType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_user.DirectlyGrantedRolesValue{}.AttributeTypes(ctx),
			},
		}
		model.DirectlyGrantedRoles, _ = types.ListValue(elementType, []attr.Value{})
	}
}

func (d *userDataSource) mapSingleRole(ctx context.Context, roleMap map[string]interface{}) datasource_user.AllRolesValue {
	roleValue := datasource_user.AllRolesValue{
		RoleId:      types.StringNull(),
		RoleName:    types.StringNull(),
		AdminOption: types.BoolNull(),
		Principal:   types.ObjectNull(datasource_user.PrincipalValue{}.AttributeTypes(ctx)),
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
		principalValue := datasource_user.PrincipalValue{
			Id:            types.StringNull(),
			PrincipalType: types.StringNull(),
		}
		if principalId, ok := principal["id"].(string); ok {
			principalValue.Id = types.StringValue(principalId)
		}
		if principalType, ok := principal["type"].(string); ok {
			principalValue.PrincipalType = types.StringValue(principalType)
		}
		principalObjectValue, diag := principalValue.ToObjectValue(ctx)
		if !diag.HasError() {
			roleValue.Principal = principalObjectValue
		}
	}

	return roleValue
}

func (d *userDataSource) mapSingleDirectRole(ctx context.Context, roleMap map[string]interface{}) datasource_user.DirectlyGrantedRolesValue {
	roleValue := datasource_user.DirectlyGrantedRolesValue{
		RoleId:      types.StringNull(),
		RoleName:    types.StringNull(),
		AdminOption: types.BoolNull(),
		Principal:   types.ObjectNull(datasource_user.PrincipalValue{}.AttributeTypes(ctx)),
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
		principalValue := datasource_user.PrincipalValue{
			Id:            types.StringNull(),
			PrincipalType: types.StringNull(),
		}
		if principalId, ok := principal["id"].(string); ok {
			principalValue.Id = types.StringValue(principalId)
		}
		if principalType, ok := principal["type"].(string); ok {
			principalValue.PrincipalType = types.StringValue(principalType)
		}
		principalObjectValue, diag := principalValue.ToObjectValue(ctx)
		if !diag.HasError() {
			roleValue.Principal = principalObjectValue
		}
	}

	return roleValue
}
