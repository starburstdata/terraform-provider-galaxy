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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_snowflake_catalogs"
)

var _ datasource.DataSource = (*snowflake_catalogsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*snowflake_catalogsDataSource)(nil)

func NewSnowflakeCatalogsDataSource() datasource.DataSource {
	return &snowflake_catalogsDataSource{}
}

type snowflake_catalogsDataSource struct {
	client *client.GalaxyClient
}

func (d *snowflake_catalogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snowflake_catalogs"
}

func (d *snowflake_catalogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_snowflake_catalogs.SnowflakeCatalogsDataSourceSchema(ctx)
}

func (d *snowflake_catalogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *snowflake_catalogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_snowflake_catalogs.SnowflakeCatalogsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading snowflake_catalogs with automatic pagination")

	// Use automatic pagination to get ALL snowflake catalogs across all pages
	allCatalogs, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/catalog?catalogType=SNOWFLAKE")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading snowflake_catalogs",
			"Could not read snowflake_catalogs: "+err.Error(),
		)
		return
	}

	// Convert []interface{} to []map[string]interface{} for mapping
	var catalogMaps []map[string]interface{}
	for _, catalogInterface := range allCatalogs {
		if catalogMap, ok := catalogInterface.(map[string]interface{}); ok {
			catalogMaps = append(catalogMaps, catalogMap)
		}
	}

	// Map API response to model
	if len(catalogMaps) > 0 {
		catalogs, err := d.mapSnowflakeCatalogsResult(ctx, catalogMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping snowflake_catalogs response",
				"Could not map snowflake_catalogs response: "+err.Error(),
			)
			return
		}
		config.Result = catalogs
	} else {
		elementType := datasource_snowflake_catalogs.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_snowflake_catalogs.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_snowflake_catalogs.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *snowflake_catalogsDataSource) mapSnowflakeCatalogsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	catalogs := make([]datasource_snowflake_catalogs.ResultValue, 0, len(result))

	for _, catalogMap := range result {
		catalog := d.mapSingleSnowflakeCatalog(ctx, catalogMap)
		catalogs = append(catalogs, catalog)
	}

	elementType := datasource_snowflake_catalogs.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_snowflake_catalogs.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, catalogs)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *snowflake_catalogsDataSource) mapSingleSnowflakeCatalog(ctx context.Context, catalogMap map[string]interface{}) datasource_snowflake_catalogs.ResultValue {
	attributeTypes := datasource_snowflake_catalogs.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map catalog ID
	if catalogId, ok := catalogMap["catalogId"].(string); ok {
		attributes["catalog_id"] = types.StringValue(catalogId)
	} else {
		attributes["catalog_id"] = types.StringNull()
	}

	// Map other required fields
	if name, ok := catalogMap["name"].(string); ok {
		attributes["name"] = types.StringValue(name)
	} else {
		attributes["name"] = types.StringNull()
	}

	if description, ok := catalogMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	if accountIdentifier, ok := catalogMap["accountIdentifier"].(string); ok {
		attributes["account_identifier"] = types.StringValue(accountIdentifier)
	} else {
		attributes["account_identifier"] = types.StringNull()
	}

	if cloudKind, ok := catalogMap["cloudKind"].(string); ok {
		attributes["cloud_kind"] = types.StringValue(cloudKind)
	} else {
		attributes["cloud_kind"] = types.StringNull()
	}

	if databaseName, ok := catalogMap["databaseName"].(string); ok {
		attributes["database_name"] = types.StringValue(databaseName)
	} else {
		attributes["database_name"] = types.StringNull()
	}

	if password, ok := catalogMap["password"].(string); ok {
		attributes["password"] = types.StringValue(password)
	} else {
		attributes["password"] = types.StringNull()
	}

	if readOnly, ok := catalogMap["readOnly"].(bool); ok {
		attributes["read_only"] = types.BoolValue(readOnly)
	} else {
		attributes["read_only"] = types.BoolNull()
	}

	if role, ok := catalogMap["role"].(string); ok {
		attributes["role"] = types.StringValue(role)
	} else {
		attributes["role"] = types.StringNull()
	}

	if useParallelMode, ok := catalogMap["useParallelMode"].(bool); ok {
		attributes["use_parallel_mode"] = types.BoolValue(useParallelMode)
	} else {
		attributes["use_parallel_mode"] = types.BoolNull()
	}

	if username, ok := catalogMap["username"].(string); ok {
		attributes["username"] = types.StringValue(username)
	} else {
		attributes["username"] = types.StringNull()
	}

	if warehouse, ok := catalogMap["warehouse"].(string); ok {
		attributes["warehouse"] = types.StringValue(warehouse)
	} else {
		attributes["warehouse"] = types.StringNull()
	}

	// Handle validate field
	if validate, ok := catalogMap["validate"].(bool); ok {
		attributes["validate"] = types.BoolValue(validate)
	} else {
		attributes["validate"] = types.BoolNull()
	}
	// Create the ResultValue using the constructor
	catalog, diags := datasource_snowflake_catalogs.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating catalog ResultValue: %v", diags))
		return datasource_snowflake_catalogs.NewResultValueNull()
	}

	return catalog
}
