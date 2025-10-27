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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_service_accounts"
)

var _ datasource.DataSource = (*service_accountsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*service_accountsDataSource)(nil)

func NewServiceAccountsDataSource() datasource.DataSource {
	return &service_accountsDataSource{}
}

type service_accountsDataSource struct {
	client *client.GalaxyClient
}

func (d *service_accountsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_accounts"
}

func (d *service_accountsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_service_accounts.ServiceAccountsDataSourceSchema(ctx)
}

func (d *service_accountsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *service_accountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_service_accounts.ServiceAccountsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading service_accounts with automatic pagination")

	// Use automatic pagination to get ALL service accounts across all pages
	allServiceAccounts, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/serviceAccount")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading service_accounts",
			"Could not read service_accounts: "+err.Error(),
		)
		return
	}

	// Convert []interface{} to []map[string]interface{} for mapping
	var serviceAccountMaps []map[string]interface{}
	for _, serviceAccountInterface := range allServiceAccounts {
		if serviceAccountMap, ok := serviceAccountInterface.(map[string]interface{}); ok {
			serviceAccountMaps = append(serviceAccountMaps, serviceAccountMap)
		}
	}

	// Map API response to model
	if len(serviceAccountMaps) > 0 {
		serviceAccounts, err := d.mapServiceAccountsResult(ctx, serviceAccountMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping service_accounts response",
				"Could not map service_accounts response: "+err.Error(),
			)
			return
		}
		config.Result = serviceAccounts
	} else {
		elementType := datasource_service_accounts.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_service_accounts.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_service_accounts.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *service_accountsDataSource) mapServiceAccountsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	serviceAccounts := make([]datasource_service_accounts.ResultValue, 0)

	for _, serviceAccountMap := range result {
		serviceAccount := d.mapSingleServiceAccount(ctx, serviceAccountMap)
		serviceAccounts = append(serviceAccounts, serviceAccount)
	}

	elementType := datasource_service_accounts.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_service_accounts.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, serviceAccounts)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *service_accountsDataSource) mapSingleServiceAccount(ctx context.Context, serviceAccountMap map[string]interface{}) datasource_service_accounts.ResultValue {
	attributeTypes := datasource_service_accounts.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map service account ID
	if serviceAccountId, ok := serviceAccountMap["serviceAccountId"].(string); ok {
		attributes["service_account_id"] = types.StringValue(serviceAccountId)
	} else {
		attributes["service_account_id"] = types.StringNull()
	}

	// Map username
	if userName, ok := serviceAccountMap["userName"].(string); ok {
		attributes["user_name"] = types.StringValue(userName)
	} else {
		attributes["user_name"] = types.StringNull()
	}

	// Map role ID
	if roleId, ok := serviceAccountMap["roleId"].(string); ok {
		attributes["role_id"] = types.StringValue(roleId)
	} else {
		attributes["role_id"] = types.StringNull()
	}

	// Map additional role IDs list
	attributes["additional_role_ids"] = d.mapAdditionalRoleIds(ctx, serviceAccountMap)

	// Map passwords list
	attributes["passwords"] = d.mapPasswords(ctx, serviceAccountMap)

	// Create the ResultValue using the constructor
	serviceAccount, diags := datasource_service_accounts.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating service account ResultValue: %v", diags))
		return datasource_service_accounts.NewResultValueNull()
	}

	return serviceAccount
}

func (d *service_accountsDataSource) mapAdditionalRoleIds(ctx context.Context, serviceAccountMap map[string]interface{}) types.List {
	if additionalRoleIds, ok := serviceAccountMap["additionalRoleIds"].([]interface{}); ok {
		roleIdList := make([]attr.Value, 0, len(additionalRoleIds))
		for _, roleId := range additionalRoleIds {
			if roleIdStr, ok := roleId.(string); ok {
				roleIdList = append(roleIdList, types.StringValue(roleIdStr))
			}
		}
		roleIdListValue, _ := types.ListValue(types.StringType, roleIdList)
		return roleIdListValue
	} else {
		return types.ListNull(types.StringType)
	}
}

func (d *service_accountsDataSource) mapPasswords(ctx context.Context, serviceAccountMap map[string]interface{}) types.List {
	if passwords, ok := serviceAccountMap["passwords"].([]interface{}); ok {
		passwordsList := make([]datasource_service_accounts.PasswordsValue, 0, len(passwords))
		for _, passwordInterface := range passwords {
			if passwordMap, ok := passwordInterface.(map[string]interface{}); ok {
				password := d.mapSinglePassword(ctx, passwordMap)
				passwordsList = append(passwordsList, password)
			}
		}

		elementType := datasource_service_accounts.PasswordsType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_service_accounts.PasswordsValue{}.AttributeTypes(ctx),
			},
		}
		passwordsListValue, _ := types.ListValueFrom(ctx, elementType, passwordsList)
		return passwordsListValue
	} else {
		elementType := datasource_service_accounts.PasswordsType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_service_accounts.PasswordsValue{}.AttributeTypes(ctx),
			},
		}
		return types.ListNull(elementType)
	}
}

func (d *service_accountsDataSource) mapSinglePassword(ctx context.Context, passwordMap map[string]interface{}) datasource_service_accounts.PasswordsValue {
	attributeTypes := datasource_service_accounts.PasswordsValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	if serviceAccountPasswordId, ok := passwordMap["serviceAccountPasswordId"].(string); ok {
		attributes["service_account_password_id"] = types.StringValue(serviceAccountPasswordId)
	} else {
		attributes["service_account_password_id"] = types.StringNull()
	}

	if created, ok := passwordMap["created"].(string); ok {
		attributes["created"] = types.StringValue(created)
	} else {
		attributes["created"] = types.StringNull()
	}

	if description, ok := passwordMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	if lastLogin, ok := passwordMap["lastLogin"].(string); ok {
		attributes["last_login"] = types.StringValue(lastLogin)
	} else {
		attributes["last_login"] = types.StringNull()
	}

	if passwordValue, ok := passwordMap["password"].(string); ok {
		attributes["password"] = types.StringValue(passwordValue)
	} else {
		attributes["password"] = types.StringNull()
	}

	if passwordPrefix, ok := passwordMap["passwordPrefix"].(string); ok {
		attributes["password_prefix"] = types.StringValue(passwordPrefix)
	} else {
		attributes["password_prefix"] = types.StringNull()
	}

	password, diags := datasource_service_accounts.NewPasswordsValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating PasswordsValue: %v", diags))
		return datasource_service_accounts.NewPasswordsValueNull()
	}

	return password
}
