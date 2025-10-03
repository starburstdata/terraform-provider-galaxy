package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_rolegrant"
)

var _ datasource.DataSource = (*rolegrantDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*rolegrantDataSource)(nil)

func NewRolegrantDataSource() datasource.DataSource {
	return &rolegrantDataSource{}
}

type rolegrantDataSource struct {
	client *client.GalaxyClient
}

func (d *rolegrantDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rolegrant"
}

func (d *rolegrantDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_rolegrant.RolegrantDataSourceSchema(ctx)
}

func (d *rolegrantDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *rolegrantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_rolegrant.RolegrantModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := config.Id.ValueString()
	tflog.Debug(ctx, "Reading role grants", map[string]interface{}{"roleId": roleID})

	response, err := d.client.ListRoleGrants(ctx, roleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role grants",
			"Could not read role grants for role "+roleID+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *rolegrantDataSource) updateModelFromResponse(ctx context.Context, model *datasource_rolegrant.RolegrantModel, response map[string]interface{}) {
	// The id (role ID) is already set from the configuration

	// Map the result array - role grants have AdminOption, Principal, RoleId, RoleName
	if resultArray, ok := response["result"].([]interface{}); ok {
		resultList := make([]datasource_rolegrant.ResultValue, 0, len(resultArray))
		for _, item := range resultArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				resultItem := datasource_rolegrant.ResultValue{
					AdminOption: types.BoolValue(getBoolFromMap(itemMap, "adminOption")),
					RoleId:      types.StringValue(getStringFromMap(itemMap, "roleId")),
					RoleName:    types.StringValue(getStringFromMap(itemMap, "roleName")),
				}

				// Map principal object
				if principal, ok := itemMap["principal"].(map[string]interface{}); ok {
					principalAttrs := map[string]attr.Value{
						"type": types.StringValue(getStringFromMap(principal, "type")),
						"name": types.StringValue(getStringFromMap(principal, "name")),
					}
					principalObj, _ := types.ObjectValue(map[string]attr.Type{
						"type": types.StringType,
						"name": types.StringType,
					}, principalAttrs)
					resultItem.Principal = principalObj
				} else {
					resultItem.Principal = types.ObjectNull(map[string]attr.Type{
						"type": types.StringType,
						"name": types.StringType,
					})
				}

				resultList = append(resultList, resultItem)
			}
		}
		resultListValue, _ := types.ListValueFrom(ctx, datasource_rolegrant.ResultType{}, resultList)
		model.Result = resultListValue
	} else {
		model.Result = types.ListNull(datasource_rolegrant.ResultType{})
	}
}

// Helper function to safely extract boolean values from maps
func getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}
