package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_cassandra_catalogs"
)

var _ datasource.DataSource = (*cassandra_catalogsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*cassandra_catalogsDataSource)(nil)

func NewCassandraCatalogsDataSource() datasource.DataSource {
	return &cassandra_catalogsDataSource{}
}

type cassandra_catalogsDataSource struct {
	client *client.GalaxyClient
}

func (d *cassandra_catalogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cassandra_catalogs"
}

func (d *cassandra_catalogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_cassandra_catalogs.CassandraCatalogsDataSourceSchema(ctx)
}

func (d *cassandra_catalogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cassandra_catalogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_cassandra_catalogs.CassandraCatalogsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading cassandra_catalogs with automatic pagination")

	// Use type-specific endpoint to get detailed Cassandra catalog information
	allCatalogs, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/catalogType/cassandra/catalog")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading cassandra_catalogs",
			"Could not read cassandra_catalogs: "+err.Error(),
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
		catalogs, err := d.mapCassandraCatalogsResult(ctx, catalogMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping cassandra_catalogs response",
				"Could not map cassandra_catalogs response: "+err.Error(),
			)
			return
		}
		config.Result = catalogs
	} else {
		elementType := datasource_cassandra_catalogs.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_cassandra_catalogs.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_cassandra_catalogs.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *cassandra_catalogsDataSource) mapCassandraCatalogsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	catalogs := make([]datasource_cassandra_catalogs.ResultValue, 0, len(result))

	for _, catalogMap := range result {
		catalog := d.mapSingleCassandraCatalog(ctx, catalogMap)
		catalogs = append(catalogs, catalog)
	}

	elementType := datasource_cassandra_catalogs.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_cassandra_catalogs.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, catalogs)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *cassandra_catalogsDataSource) mapSingleCassandraCatalog(ctx context.Context, catalogMap map[string]interface{}) datasource_cassandra_catalogs.ResultValue {
	attributeTypes := datasource_cassandra_catalogs.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Debug: Log the actual catalog map to see what fields are present
	tflog.Debug(ctx, "Mapping cassandra catalog", map[string]interface{}{
		"catalogMap": catalogMap,
	})

	// Map catalog ID
	if catalogId, ok := catalogMap["catalogId"].(string); ok {
		attributes["catalog_id"] = types.StringValue(catalogId)
	} else {
		attributes["catalog_id"] = types.StringNull()
	}

	// Map cloud_kind
	if cloudKind, ok := catalogMap["cloudKind"].(string); ok {
		attributes["cloud_kind"] = types.StringValue(cloudKind)
	} else {
		attributes["cloud_kind"] = types.StringNull()
	}

	// Map contact_points
	if contactPoints, ok := catalogMap["contactPoints"].(string); ok {
		attributes["contact_points"] = types.StringValue(contactPoints)
	} else {
		attributes["contact_points"] = types.StringNull()
	}

	// Map database_id - not available in Cassandra API, always null
	attributes["database_id"] = types.StringNull()

	// Map deployment_type
	if deploymentType, ok := catalogMap["deploymentType"].(string); ok {
		attributes["deployment_type"] = types.StringValue(deploymentType)
	} else {
		attributes["deployment_type"] = types.StringNull()
	}

	// Map description
	if description, ok := catalogMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	// Map local_datacenter
	if localDatacenter, ok := catalogMap["localDatacenter"].(string); ok {
		attributes["local_datacenter"] = types.StringValue(localDatacenter)
	} else {
		attributes["local_datacenter"] = types.StringNull()
	}

	// Map name
	if name, ok := catalogMap["name"].(string); ok {
		attributes["name"] = types.StringValue(name)
	} else {
		attributes["name"] = types.StringNull()
	}

	// Map password
	if password, ok := catalogMap["password"].(string); ok {
		attributes["password"] = types.StringValue(password)
	} else {
		attributes["password"] = types.StringNull()
	}

	// Map port
	if portFloat, ok := catalogMap["port"].(float64); ok {
		attributes["port"] = types.Int64Value(int64(portFloat))
	} else if portInt, ok := catalogMap["port"].(int64); ok {
		attributes["port"] = types.Int64Value(portInt)
	} else if portInt, ok := catalogMap["port"].(int); ok {
		attributes["port"] = types.Int64Value(int64(portInt))
	} else {
		attributes["port"] = types.Int64Null()
	}

	// Map read_only
	if readOnly, ok := catalogMap["readOnly"].(bool); ok {
		attributes["read_only"] = types.BoolValue(readOnly)
	} else {
		attributes["read_only"] = types.BoolNull()
	}

	// Map region - not available in Cassandra API, always null
	attributes["region"] = types.StringNull()

	// Map ssh_tunnel_id - not available in Cassandra API, always null
	attributes["ssh_tunnel_id"] = types.StringNull()

	// Map tls_enabled - not available in Cassandra API, always null
	attributes["tls_enabled"] = types.BoolNull()

	// Map token - not available in Cassandra API, always null
	attributes["token"] = types.StringNull()

	// Map username
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
	catalog, diags := datasource_cassandra_catalogs.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating catalog ResultValue: %v", diags))
		return datasource_cassandra_catalogs.NewResultValueNull()
	}

	return catalog
}
