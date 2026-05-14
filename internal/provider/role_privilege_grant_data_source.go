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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_role_privilege_grant"
)

var _ datasource.DataSource = (*rolePrivilegeGrantDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*rolePrivilegeGrantDataSource)(nil)

func NewRolePrivilegeGrantDataSource() datasource.DataSource {
	return &rolePrivilegeGrantDataSource{}
}

type rolePrivilegeGrantDataSource struct {
	client *client.GalaxyClient
}

func (d *rolePrivilegeGrantDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_privilege_grants"
}

func (d *rolePrivilegeGrantDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_role_privilege_grant.RolePrivilegeGrantDataSourceSchema(ctx)
}

func (d *rolePrivilegeGrantDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *rolePrivilegeGrantDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_role_privilege_grant.RolePrivilegeGrantModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleId := config.RoleId.ValueString()

	path := "/public/api/v1/role/" + roleId + "/privilege"
	listAll := !config.ListAllPrivileges.IsNull() && !config.ListAllPrivileges.IsUnknown() && config.ListAllPrivileges.ValueBool()
	if listAll {
		path += "?listAllPrivileges=true"
	}

	tflog.Debug(ctx, "Reading role privilege grants", map[string]interface{}{
		"roleId":            roleId,
		"listAllPrivileges": listAll,
	})

	allGrants, err := d.client.GetAllPaginatedResults(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role privilege grants",
			"Could not read role privilege grants: "+err.Error(),
		)
		return
	}

	elementType := datasource_role_privilege_grant.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_role_privilege_grant.ResultValue{}.AttributeTypes(ctx),
		},
	}

	grants := make([]datasource_role_privilege_grant.ResultValue, 0, len(allGrants))
	for _, g := range allGrants {
		grantMap, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		grant, diags := d.mapGrant(ctx, grantMap)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		grants = append(grants, grant)
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, grants)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	config.Result = listValue
	config.ListAllPrivileges = types.BoolValue(listAll)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *rolePrivilegeGrantDataSource) mapGrant(ctx context.Context, grantMap map[string]interface{}) (datasource_role_privilege_grant.ResultValue, diag.Diagnostics) {
	attributeTypes := datasource_role_privilege_grant.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	if v, ok := grantMap["entityId"].(string); ok {
		attributes["entity_id"] = types.StringValue(v)
	} else {
		attributes["entity_id"] = types.StringNull()
	}

	if v, ok := grantMap["entityKind"].(string); ok {
		attributes["entity_kind"] = types.StringValue(v)
	} else {
		attributes["entity_kind"] = types.StringNull()
	}

	if v, ok := grantMap["grantKind"].(string); ok {
		attributes["grant_kind"] = types.StringValue(v)
	} else {
		attributes["grant_kind"] = types.StringNull()
	}

	if v, ok := grantMap["grantOption"].(bool); ok {
		attributes["grant_option"] = types.BoolValue(v)
	} else {
		attributes["grant_option"] = types.BoolValue(false)
	}

	if v, ok := grantMap["privilege"].(string); ok {
		attributes["privilege"] = types.StringValue(v)
	} else {
		attributes["privilege"] = types.StringNull()
	}

	if v, ok := grantMap["schemaName"].(string); ok {
		attributes["schema_name"] = types.StringValue(v)
	} else {
		attributes["schema_name"] = types.StringNull()
	}

	if v, ok := grantMap["tableName"].(string); ok {
		attributes["table_name"] = types.StringValue(v)
	} else {
		attributes["table_name"] = types.StringNull()
	}

	if v, ok := grantMap["columnName"].(string); ok {
		attributes["column_name"] = types.StringValue(v)
	} else {
		attributes["column_name"] = types.StringNull()
	}

	return datasource_role_privilege_grant.NewResultValue(attributeTypes, attributes)
}
