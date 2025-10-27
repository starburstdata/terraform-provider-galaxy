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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_postgresql_catalogs"
)

var _ datasource.DataSource = (*postgresql_catalogsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*postgresql_catalogsDataSource)(nil)

func NewPostgresqlCatalogsDataSource() datasource.DataSource {
	return &postgresql_catalogsDataSource{}
}

type postgresql_catalogsDataSource struct {
	client *client.GalaxyClient
}

func (d *postgresql_catalogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgresql_catalogs"
}

func (d *postgresql_catalogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_postgresql_catalogs.PostgresqlCatalogsDataSourceSchema(ctx)
}

func (d *postgresql_catalogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *postgresql_catalogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_postgresql_catalogs.PostgresqlCatalogsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading postgresql_catalogs with automatic pagination")

	// Use automatic pagination to get ALL PostgreSQL catalogs across all pages
	allCatalogs, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/catalog?catalogType=POSTGRESQL")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading postgresql_catalogs",
			"Could not read postgresql_catalogs: "+err.Error(),
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
		catalogs, err := d.mapPostgresqlCatalogsResult(ctx, catalogMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping postgresql_catalogs response",
				"Could not map postgresql_catalogs response: "+err.Error(),
			)
			return
		}
		config.Result = catalogs
	} else {
		elementType := datasource_postgresql_catalogs.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_postgresql_catalogs.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_postgresql_catalogs.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *postgresql_catalogsDataSource) mapPostgresqlCatalogsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	catalogs := make([]datasource_postgresql_catalogs.ResultValue, 0)

	for _, catalogMap := range result {
		catalog := d.mapSinglePostgresqlCatalog(ctx, catalogMap)
		catalogs = append(catalogs, catalog)
	}

	elementType := datasource_postgresql_catalogs.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_postgresql_catalogs.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, catalogs)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *postgresql_catalogsDataSource) mapSinglePostgresqlCatalog(ctx context.Context, catalogMap map[string]interface{}) datasource_postgresql_catalogs.ResultValue {
	attributeTypes := datasource_postgresql_catalogs.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map catalog ID
	if catalogId, ok := catalogMap["catalogId"].(string); ok {
		attributes["catalog_id"] = types.StringValue(catalogId)
	} else {
		attributes["catalog_id"] = types.StringNull()
	}

	// Map cloud kind
	if cloudKind, ok := catalogMap["cloudKind"].(string); ok {
		attributes["cloud_kind"] = types.StringValue(cloudKind)
	} else {
		attributes["cloud_kind"] = types.StringNull()
	}

	// Map database name
	if databaseName, ok := catalogMap["databaseName"].(string); ok {
		attributes["database_name"] = types.StringValue(databaseName)
	} else {
		attributes["database_name"] = types.StringNull()
	}

	// Map description
	if description, ok := catalogMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	// Map endpoint
	if endpoint, ok := catalogMap["endpoint"].(string); ok {
		attributes["endpoint"] = types.StringValue(endpoint)
	} else {
		attributes["endpoint"] = types.StringNull()
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
	if port, ok := catalogMap["port"].(float64); ok {
		attributes["port"] = types.Int64Value(int64(port))
	} else {
		attributes["port"] = types.Int64Null()
	}

	// Map read only
	if readOnly, ok := catalogMap["readOnly"].(bool); ok {
		attributes["read_only"] = types.BoolValue(readOnly)
	} else {
		attributes["read_only"] = types.BoolNull()
	}

	// Map SSH tunnel ID
	if sshTunnelId, ok := catalogMap["sshTunnelId"].(string); ok {
		attributes["ssh_tunnel_id"] = types.StringValue(sshTunnelId)
	} else {
		attributes["ssh_tunnel_id"] = types.StringNull()
	}

	// Map TLS enabled
	if tlsEnabled, ok := catalogMap["tlsEnabled"].(bool); ok {
		attributes["tls_enabled"] = types.BoolValue(tlsEnabled)
	} else {
		attributes["tls_enabled"] = types.BoolNull()
	}

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
	catalog, diags := datasource_postgresql_catalogs.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating PostgreSQL catalog ResultValue: %v", diags))
		return datasource_postgresql_catalogs.NewResultValueNull()
	}

	return catalog
}
