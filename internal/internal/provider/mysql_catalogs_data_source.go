package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_mysql_catalogs"
)

var _ datasource.DataSource = (*mysql_catalogsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*mysql_catalogsDataSource)(nil)

func NewMysqlCatalogsDataSource() datasource.DataSource {
	return &mysql_catalogsDataSource{}
}

type mysql_catalogsDataSource struct {
	client *client.GalaxyClient
}

func (d *mysql_catalogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_catalogs"
}

func (d *mysql_catalogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_mysql_catalogs.MysqlCatalogsDataSourceSchema(ctx)
}

func (d *mysql_catalogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *mysql_catalogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_mysql_catalogs.MysqlCatalogsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading mysql_catalogs with automatic pagination")

	// Use automatic pagination to get ALL mysql catalogs across all pages
	allCatalogs, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/catalog?catalogType=MYSQL")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading mysql_catalogs",
			"Could not read mysql_catalogs: "+err.Error(),
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
		catalogs, err := d.mapMysqlCatalogsResult(ctx, catalogMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping mysql_catalogs response",
				"Could not map mysql_catalogs response: "+err.Error(),
			)
			return
		}
		config.Result = catalogs
	} else {
		elementType := datasource_mysql_catalogs.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_mysql_catalogs.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_mysql_catalogs.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *mysql_catalogsDataSource) mapMysqlCatalogsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	catalogs := make([]datasource_mysql_catalogs.ResultValue, 0, len(result))

	for _, catalogMap := range result {
		catalog := d.mapSingleMysqlCatalog(ctx, catalogMap)
		catalogs = append(catalogs, catalog)
	}

	elementType := datasource_mysql_catalogs.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_mysql_catalogs.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, catalogs)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *mysql_catalogsDataSource) mapSingleMysqlCatalog(ctx context.Context, catalogMap map[string]interface{}) datasource_mysql_catalogs.ResultValue {
	attributeTypes := datasource_mysql_catalogs.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map catalog ID
	if catalogId, ok := catalogMap["catalogId"].(string); ok {
		attributes["catalog_id"] = types.StringValue(catalogId)
	} else {
		attributes["catalog_id"] = types.StringNull()
	}

	// Map other fields
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

	if cloudKind, ok := catalogMap["cloudKind"].(string); ok {
		attributes["cloud_kind"] = types.StringValue(cloudKind)
	} else {
		attributes["cloud_kind"] = types.StringNull()
	}

	if connectionType, ok := catalogMap["connectionType"].(string); ok {
		attributes["connection_type"] = types.StringValue(connectionType)
	} else {
		attributes["connection_type"] = types.StringNull()
	}

	if password, ok := catalogMap["password"].(string); ok {
		attributes["password"] = types.StringValue(password)
	} else {
		attributes["password"] = types.StringNull()
	}

	if port, ok := catalogMap["port"].(float64); ok {
		attributes["port"] = types.Int64Value(int64(port))
	} else {
		attributes["port"] = types.Int64Null()
	}

	if readOnly, ok := catalogMap["readOnly"].(bool); ok {
		attributes["read_only"] = types.BoolValue(readOnly)
	} else {
		attributes["read_only"] = types.BoolNull()
	}

	if username, ok := catalogMap["username"].(string); ok {
		attributes["username"] = types.StringValue(username)
	} else {
		attributes["username"] = types.StringNull()
	}

	// Handle validate field
	if validate, ok := catalogMap["validate"].(bool); ok {
		attributes["validate"] = types.BoolValue(validate)
	} else {
		attributes["validate"] = types.BoolNull()
	}
	// Create the ResultValue using the constructor
	catalog, diags := datasource_mysql_catalogs.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating catalog ResultValue: %v", diags))
		return datasource_mysql_catalogs.NewResultValueNull()
	}

	return catalog
}
