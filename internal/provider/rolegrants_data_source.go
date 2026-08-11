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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_rolegrants"
)

var _ datasource.DataSource = (*rolegrantsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*rolegrantsDataSource)(nil)

func NewRolegrantsDataSource() datasource.DataSource {
	return &rolegrantsDataSource{}
}

type rolegrantsDataSource struct {
	client *client.GalaxyClient
}

func (d *rolegrantsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rolegrants"
}

func (d *rolegrantsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_rolegrants.RolegrantsDataSourceSchema(ctx)
}

func (d *rolegrantsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *rolegrantsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_rolegrants.RolegrantsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := config.RoleId.ValueString()
	tflog.Debug(ctx, "Reading rolegrants", map[string]interface{}{"role_id": roleID})

	response, err := d.client.ListRoleGrants(ctx, roleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading rolegrants",
			"Could not read rolegrants for role "+roleID+": "+err.Error(),
		)
		return
	}

	// Defaults
	if config.Type.IsNull() || config.Type.IsUnknown() {
		config.Type = types.StringNull()
	}

	// Map results
	elementType := datasource_rolegrants.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_rolegrants.ResultValue{}.AttributeTypes(ctx),
		},
	}

	var results []interface{}
	if items, ok := response["items"].([]interface{}); ok {
		results = items
	} else if arr, ok := response["result"].([]interface{}); ok {
		results = arr
	}

	if len(results) > 0 {
		var resultElements []attr.Value
		for _, itemRaw := range results {
			if item, ok := itemRaw.(map[string]interface{}); ok {
				resultValue, diags := d.mapSingleResult(ctx, item)
				resp.Diagnostics.Append(diags...)
				if diags.HasError() {
					continue
				}
				resultElements = append(resultElements, resultValue)
			}
		}
		config.Result, _ = types.ListValue(elementType, resultElements)
	} else {
		config.Result, _ = types.ListValue(elementType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *rolegrantsDataSource) mapSingleResult(ctx context.Context, item map[string]interface{}) (datasource_rolegrants.ResultValue, diag.Diagnostics) {
	attributeTypes := datasource_rolegrants.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	if roleId, ok := item["roleId"].(string); ok {
		attributes["role_id"] = types.StringValue(roleId)
	} else {
		attributes["role_id"] = types.StringNull()
	}
	if roleName, ok := item["roleName"].(string); ok {
		attributes["role_name"] = types.StringValue(roleName)
	} else {
		attributes["role_name"] = types.StringNull()
	}
	if adminOption, ok := item["adminOption"].(bool); ok {
		attributes["admin_option"] = types.BoolValue(adminOption)
	} else {
		attributes["admin_option"] = types.BoolNull()
	}

	// Map principal
	principalAttributeTypes := datasource_rolegrants.PrincipalValue{}.AttributeTypes(ctx)
	if principalData, ok := item["principal"].(map[string]interface{}); ok {
		principalAttributes := map[string]attr.Value{}
		if id, ok := principalData["id"].(string); ok {
			principalAttributes["id"] = types.StringValue(id)
		} else {
			principalAttributes["id"] = types.StringNull()
		}
		if pType, ok := principalData["type"].(string); ok {
			principalAttributes["type"] = types.StringValue(pType)
		} else {
			principalAttributes["type"] = types.StringNull()
		}
		principalValue, diags := datasource_rolegrants.NewPrincipalValue(principalAttributeTypes, principalAttributes)
		if diags.HasError() {
			return datasource_rolegrants.ResultValue{}, diags
		}
		principalObj, diags := principalValue.ToObjectValue(ctx)
		if diags.HasError() {
			return datasource_rolegrants.ResultValue{}, diags
		}
		attributes["principal"] = principalObj
	} else {
		attributes["principal"] = types.ObjectNull(principalAttributeTypes)
	}

	return datasource_rolegrants.NewResultValue(attributeTypes, attributes)
}
