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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_mongodb_catalogs"
)

var _ datasource.DataSource = (*mongodb_catalogsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*mongodb_catalogsDataSource)(nil)

func NewMongodbCatalogsDataSource() datasource.DataSource {
	return &mongodb_catalogsDataSource{}
}

type mongodb_catalogsDataSource struct {
	client *client.GalaxyClient
}

func (d *mongodb_catalogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mongodb_catalogs"
}

func (d *mongodb_catalogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_mongodb_catalogs.MongodbCatalogsDataSourceSchema(ctx)
}

func (d *mongodb_catalogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *mongodb_catalogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_mongodb_catalogs.MongodbCatalogsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading mongodb_catalogs with automatic pagination")

	// Use automatic pagination to get ALL mongodb catalogs across all pages
	allCatalogs, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/catalog?catalogType=MONGODB")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading mongodb_catalogs",
			"Could not read mongodb_catalogs: "+err.Error(),
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
		catalogs, err := d.mapMongodbCatalogsResult(ctx, catalogMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping mongodb_catalogs response",
				"Could not map mongodb_catalogs response: "+err.Error(),
			)
			return
		}
		config.Result = catalogs
	} else {
		elementType := datasource_mongodb_catalogs.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_mongodb_catalogs.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_mongodb_catalogs.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *mongodb_catalogsDataSource) mapMongodbCatalogsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	catalogs := make([]datasource_mongodb_catalogs.ResultValue, 0, len(result))

	for _, catalogMap := range result {
		catalog := d.mapSingleMongodbCatalog(ctx, catalogMap)
		catalogs = append(catalogs, catalog)
	}

	elementType := datasource_mongodb_catalogs.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_mongodb_catalogs.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, catalogs)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *mongodb_catalogsDataSource) mapSingleMongodbCatalog(ctx context.Context, catalogMap map[string]interface{}) datasource_mongodb_catalogs.ResultValue {
	attributeTypes := datasource_mongodb_catalogs.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map catalog ID
	if catalogId, ok := catalogMap["catalogId"].(string); ok {
		attributes["catalog_id"] = types.StringValue(catalogId)
	} else {
		attributes["catalog_id"] = types.StringNull()
	}

	// Map other required fields
	if name, ok := catalogMap["catalogName"].(string); ok {
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

	if dnsSeedListEnabled, ok := catalogMap["dnsSeedListEnabled"].(bool); ok {
		attributes["dns_seed_list_enabled"] = types.BoolValue(dnsSeedListEnabled)
	} else {
		attributes["dns_seed_list_enabled"] = types.BoolNull()
	}

	if federatedDatabaseEnabled, ok := catalogMap["federatedDatabaseEnabled"].(bool); ok {
		attributes["federated_database_enabled"] = types.BoolValue(federatedDatabaseEnabled)
	} else {
		attributes["federated_database_enabled"] = types.BoolNull()
	}

	if hosts, ok := catalogMap["hosts"].(string); ok {
		attributes["hosts"] = types.StringValue(hosts)
	} else {
		attributes["hosts"] = types.StringNull()
	}

	if password, ok := catalogMap["password"].(string); ok {
		attributes["password"] = types.StringValue(password)
	} else {
		attributes["password"] = types.StringNull()
	}

	if privateLinkId, ok := catalogMap["privateLinkId"].(string); ok {
		attributes["private_link_id"] = types.StringValue(privateLinkId)
	} else {
		attributes["private_link_id"] = types.StringNull()
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

	if tlsEnabled, ok := catalogMap["tlsEnabled"].(bool); ok {
		attributes["tls_enabled"] = types.BoolValue(tlsEnabled)
	} else {
		attributes["tls_enabled"] = types.BoolNull()
	}

	if username, ok := catalogMap["username"].(string); ok {
		attributes["username"] = types.StringValue(username)
	} else {
		attributes["username"] = types.StringNull()
	}

	// Handle regions list
	if regions, ok := catalogMap["regions"].([]interface{}); ok {
		regionStrings := make([]string, 0, len(regions))
		for _, region := range regions {
			if regionStr, ok := region.(string); ok {
				regionStrings = append(regionStrings, regionStr)
			}
		}
		regionListValue, _ := types.ListValueFrom(ctx, types.StringType, regionStrings)
		attributes["regions"] = regionListValue
	} else {
		attributes["regions"] = types.ListNull(types.StringType)
	}

	// Handle validate field
	if validate, ok := catalogMap["validate"].(bool); ok {
		attributes["validate"] = types.BoolValue(validate)
	} else {
		attributes["validate"] = types.BoolNull()
	}

	// Create the ResultValue using the constructor
	catalog, diags := datasource_mongodb_catalogs.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating catalog ResultValue: %v", diags))
		return datasource_mongodb_catalogs.NewResultValueNull()
	}

	return catalog
}
