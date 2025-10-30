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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_service_account"
)

var _ datasource.DataSource = (*service_accountDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*service_accountDataSource)(nil)

func NewServiceAccountDataSource() datasource.DataSource {
	return &service_accountDataSource{}
}

type service_accountDataSource struct {
	client *client.GalaxyClient
}

func (d *service_accountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (d *service_accountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_service_account.ServiceAccountDataSourceSchema(ctx)
}

func (d *service_accountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *service_accountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_service_account.ServiceAccountModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.ServiceAccountId.ValueString()
	tflog.Debug(ctx, "Reading service_account", map[string]interface{}{"id": id})

	response, err := d.client.GetServiceAccount(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service_account",
			"Could not read service_account "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *service_accountDataSource) updateModelFromResponse(ctx context.Context, model *datasource_service_account.ServiceAccountModel, response map[string]interface{}) {
	// Map response fields to model
	if serviceAccountId, ok := response["serviceAccountId"].(string); ok {
		model.ServiceAccountId = types.StringValue(serviceAccountId)
	}
	// Map UserName and RoleId
	if userName, ok := response["userName"].(string); ok {
		model.UserName = types.StringValue(userName)
	} else if name, ok := response["name"].(string); ok {
		model.UserName = types.StringValue(name)
	}
	if roleId, ok := response["roleId"].(string); ok {
		model.RoleId = types.StringValue(roleId)
	}
	// Map additional role IDs if present
	if additionalRoleIds, ok := response["additionalRoleIds"].([]interface{}); ok {
		roleList := make([]types.String, 0, len(additionalRoleIds))
		for _, roleId := range additionalRoleIds {
			if roleStr, ok := roleId.(string); ok {
				roleList = append(roleList, types.StringValue(roleStr))
			}
		}
		model.AdditionalRoleIds, _ = types.ListValueFrom(ctx, types.StringType, roleList)
	} else {
		model.AdditionalRoleIds = types.ListNull(types.StringType)
	}
	// Passwords field is typically not returned in GET operations
	passwordsType := datasource_service_account.PasswordsType{
		ObjectType: types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"created":                     types.StringType,
				"description":                 types.StringType,
				"last_login":                  types.StringType,
				"password":                    types.StringType,
				"password_prefix":             types.StringType,
				"service_account_password_id": types.StringType,
			},
		},
	}
	model.Passwords = types.ListNull(passwordsType)
}
