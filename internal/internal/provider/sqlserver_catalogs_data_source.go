package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_sqlserver_catalogs"
)

var _ datasource.DataSource = (*sqlserver_catalogsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*sqlserver_catalogsDataSource)(nil)

func NewSqlserverCatalogsDataSource() datasource.DataSource {
	return &sqlserver_catalogsDataSource{}
}

type sqlserver_catalogsDataSource struct {
	client *client.GalaxyClient
}

func (d *sqlserver_catalogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sqlserver_catalogs"
}

func (d *sqlserver_catalogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_sqlserver_catalogs.SqlserverCatalogsDataSourceSchema(ctx)
}

func (d *sqlserver_catalogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sqlserver_catalogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_sqlserver_catalogs.SqlserverCatalogsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading sqlserver_catalogs with automatic pagination")

	// Use automatic pagination to get ALL sqlserver catalogs across all pages
	allCatalogs, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/catalog?catalogType=SQLSERVER")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading sqlserver_catalogs",
			"Could not read sqlserver_catalogs: "+err.Error(),
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
		catalogs, err := d.mapSqlserverCatalogsResult(ctx, catalogMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping sqlserver_catalogs response",
				"Could not map sqlserver_catalogs response: "+err.Error(),
			)
			return
		}
		config.Result = catalogs
	} else {
		elementType := datasource_sqlserver_catalogs.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_sqlserver_catalogs.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_sqlserver_catalogs.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *sqlserver_catalogsDataSource) mapSqlserverCatalogsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	catalogs := make([]datasource_sqlserver_catalogs.ResultValue, 0, len(result))

	for _, catalogMap := range result {
		catalog := d.mapSingleSqlserverCatalog(ctx, catalogMap)
		catalogs = append(catalogs, catalog)
	}

	elementType := datasource_sqlserver_catalogs.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_sqlserver_catalogs.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, catalogs)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *sqlserver_catalogsDataSource) mapSingleSqlserverCatalog(ctx context.Context, catalogMap map[string]interface{}) datasource_sqlserver_catalogs.ResultValue {
	attributeTypes := datasource_sqlserver_catalogs.ResultValue{}.AttributeTypes(ctx)
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

	if endpoint, ok := catalogMap["endpoint"].(string); ok {
		attributes["endpoint"] = types.StringValue(endpoint)
	} else {
		attributes["endpoint"] = types.StringNull()
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

	if sshTunnelId, ok := catalogMap["sshTunnelId"].(string); ok {
		attributes["ssh_tunnel_id"] = types.StringValue(sshTunnelId)
	} else {
		attributes["ssh_tunnel_id"] = types.StringNull()
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
	catalog, diags := datasource_sqlserver_catalogs.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating catalog ResultValue: %v", diags))
		return datasource_sqlserver_catalogs.NewResultValueNull()
	}

	return catalog
}
