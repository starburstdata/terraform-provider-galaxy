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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_opensearch_catalogs"
)

var _ datasource.DataSource = (*opensearch_catalogsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*opensearch_catalogsDataSource)(nil)

func NewOpensearchCatalogsDataSource() datasource.DataSource {
	return &opensearch_catalogsDataSource{}
}

type opensearch_catalogsDataSource struct {
	client *client.GalaxyClient
}

func (d *opensearch_catalogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_opensearch_catalogs"
}

func (d *opensearch_catalogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_opensearch_catalogs.OpensearchCatalogsDataSourceSchema(ctx)
}

func (d *opensearch_catalogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *opensearch_catalogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_opensearch_catalogs.OpensearchCatalogsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading opensearch_catalogs with automatic pagination")

	// Use automatic pagination to get ALL opensearch catalogs across all pages
	allCatalogs, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/catalog?catalogType=OPENSEARCH")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading opensearch_catalogs",
			"Could not read opensearch_catalogs: "+err.Error(),
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
		catalogs, err := d.mapOpensearchCatalogsResult(ctx, catalogMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping opensearch_catalogs response",
				"Could not map opensearch_catalogs response: "+err.Error(),
			)
			return
		}
		config.Result = catalogs
	} else {
		elementType := datasource_opensearch_catalogs.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_opensearch_catalogs.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_opensearch_catalogs.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *opensearch_catalogsDataSource) mapOpensearchCatalogsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	catalogs := make([]datasource_opensearch_catalogs.ResultValue, 0, len(result))

	for _, catalogMap := range result {
		catalog := d.mapSingleOpensearchCatalog(ctx, catalogMap)
		catalogs = append(catalogs, catalog)
	}

	elementType := datasource_opensearch_catalogs.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_opensearch_catalogs.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, catalogs)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *opensearch_catalogsDataSource) mapSingleOpensearchCatalog(ctx context.Context, catalogMap map[string]interface{}) datasource_opensearch_catalogs.ResultValue {
	attributeTypes := datasource_opensearch_catalogs.ResultValue{}.AttributeTypes(ctx)
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

	if authType, ok := catalogMap["authType"].(string); ok {
		attributes["auth_type"] = types.StringValue(authType)
	} else {
		attributes["auth_type"] = types.StringNull()
	}

	if endpoint, ok := catalogMap["endpoint"].(string); ok {
		attributes["endpoint"] = types.StringValue(endpoint)
	} else {
		attributes["endpoint"] = types.StringNull()
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

	// Handle validate field
	if validate, ok := catalogMap["validate"].(bool); ok {
		attributes["validate"] = types.BoolValue(validate)
	} else {
		attributes["validate"] = types.BoolNull()
	}
	// Create the ResultValue using the constructor
	catalog, diags := datasource_opensearch_catalogs.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating catalog ResultValue: %v", diags))
		return datasource_opensearch_catalogs.NewResultValueNull()
	}

	return catalog
}
