package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_cross_account_iam_roles"
)

var _ datasource.DataSource = (*cross_account_iam_rolesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*cross_account_iam_rolesDataSource)(nil)

func NewCrossAccountIamRolesDataSource() datasource.DataSource {
	return &cross_account_iam_rolesDataSource{}
}

type cross_account_iam_rolesDataSource struct {
	client *client.GalaxyClient
}

func (d *cross_account_iam_rolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cross_account_iam_roles"
}

func (d *cross_account_iam_rolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_cross_account_iam_roles.CrossAccountIamRolesDataSourceSchema(ctx)
}

func (d *cross_account_iam_rolesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cross_account_iam_rolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_cross_account_iam_roles.CrossAccountIamRolesModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading cross_account_iam_roles with automatic pagination")

	// Use automatic pagination to get ALL cross-account IAM roles across all pages
	allRoles, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/crossAccountIamRole")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading cross_account_iam_roles",
			"Could not read cross_account_iam_roles: "+err.Error(),
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
		roles, err := d.mapCrossAccountIamRolesResult(ctx, roleMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping cross_account_iam_roles response",
				"Could not map cross_account_iam_roles response: "+err.Error(),
			)
			return
		}
		config.Result = roles
	} else {
		elementType := datasource_cross_account_iam_roles.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_cross_account_iam_roles.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_cross_account_iam_roles.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *cross_account_iam_rolesDataSource) mapCrossAccountIamRolesResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	roles := make([]datasource_cross_account_iam_roles.ResultValue, 0, len(result))

	for _, roleMap := range result {
		role := d.mapSingleCrossAccountIamRole(ctx, roleMap)
		roles = append(roles, role)
	}

	elementType := datasource_cross_account_iam_roles.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_cross_account_iam_roles.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, roles)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *cross_account_iam_rolesDataSource) mapSingleCrossAccountIamRole(ctx context.Context, roleMap map[string]interface{}) datasource_cross_account_iam_roles.ResultValue {
	attributeTypes := datasource_cross_account_iam_roles.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map alias_name
	if aliasName, ok := roleMap["aliasName"].(string); ok {
		attributes["alias_name"] = types.StringValue(aliasName)
	} else {
		attributes["alias_name"] = types.StringNull()
	}

	// Map aws_iam_arn
	if awsIamArn, ok := roleMap["awsIamArn"].(string); ok {
		attributes["aws_iam_arn"] = types.StringValue(awsIamArn)
	} else {
		attributes["aws_iam_arn"] = types.StringNull()
	}

	// Handle dependants list
	if dependants, ok := roleMap["dependants"].([]interface{}); ok {
		dependantsStrings := make([]string, 0, len(dependants))
		for _, dependantInterface := range dependants {
			if dependantStr, ok := dependantInterface.(string); ok {
				dependantsStrings = append(dependantsStrings, dependantStr)
			}
		}
		dependantsListValue, _ := types.ListValueFrom(ctx, types.StringType, dependantsStrings)
		attributes["dependants"] = dependantsListValue
	} else {
		attributes["dependants"] = types.ListNull(types.StringType)
	}

	// Create the ResultValue using the constructor
	role, diags := datasource_cross_account_iam_roles.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating role ResultValue: %v", diags))
		return datasource_cross_account_iam_roles.NewResultValueNull()
	}

	return role
}
