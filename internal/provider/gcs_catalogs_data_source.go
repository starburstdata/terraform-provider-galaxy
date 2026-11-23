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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_gcs_catalogs"
)

var _ datasource.DataSource = (*gcs_catalogsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*gcs_catalogsDataSource)(nil)

func NewGcsCatalogsDataSource() datasource.DataSource {
	return &gcs_catalogsDataSource{}
}

type gcs_catalogsDataSource struct {
	client *client.GalaxyClient
}

func (d *gcs_catalogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcs_catalogs"
}

func (d *gcs_catalogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_gcs_catalogs.GcsCatalogsDataSourceSchema(ctx)
}

func (d *gcs_catalogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *gcs_catalogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_gcs_catalogs.GcsCatalogsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading gcs_catalogs with automatic pagination")

	// Use automatic pagination to get ALL GCS catalogs across all pages
	allCatalogs, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/catalogType/gcs/catalog")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading gcs_catalogs",
			"Could not read gcs_catalogs: "+err.Error(),
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
		catalogs, err := d.mapGcsCatalogsResult(ctx, catalogMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping gcs_catalogs response",
				"Could not map gcs_catalogs response: "+err.Error(),
			)
			return
		}
		config.Result = catalogs
	} else {
		elementType := datasource_gcs_catalogs.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_gcs_catalogs.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_gcs_catalogs.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *gcs_catalogsDataSource) mapGcsCatalogsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	catalogs := make([]datasource_gcs_catalogs.ResultValue, 0)

	for _, catalogMap := range result {
		catalog := d.mapSingleGcsCatalog(ctx, catalogMap)
		catalogs = append(catalogs, catalog)
	}

	elementType := datasource_gcs_catalogs.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_gcs_catalogs.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, catalogs)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *gcs_catalogsDataSource) mapSingleGcsCatalog(ctx context.Context, catalogMap map[string]interface{}) datasource_gcs_catalogs.ResultValue {
	attributeTypes := datasource_gcs_catalogs.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	if catalogId, ok := catalogMap["catalogId"].(string); ok {
		attributes["catalog_id"] = types.StringValue(catalogId)
	} else {
		attributes["catalog_id"] = types.StringNull()
	}

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

	if metastoreType, ok := catalogMap["metastoreType"].(string); ok {
		attributes["metastore_type"] = types.StringValue(metastoreType)
	} else {
		attributes["metastore_type"] = types.StringNull()
	}

	// credentialsKey is write-only; the API returns "<Value is encrypted>" or omits it.
	// Surface the sentinel as null rather than leaking it into state.
	if credentialsKey, ok := catalogMap["credentialsKey"].(string); ok && credentialsKey != "" && credentialsKey != "<Value is encrypted>" {
		attributes["credentials_key"] = types.StringValue(credentialsKey)
	} else {
		attributes["credentials_key"] = types.StringNull()
	}

	if defaultBucket, ok := catalogMap["defaultBucket"].(string); ok {
		attributes["default_bucket"] = types.StringValue(defaultBucket)
	} else {
		attributes["default_bucket"] = types.StringNull()
	}

	if defaultDataLocation, ok := catalogMap["defaultDataLocation"].(string); ok {
		attributes["default_data_location"] = types.StringValue(defaultDataLocation)
	} else {
		attributes["default_data_location"] = types.StringNull()
	}

	if defaultTableFormat, ok := catalogMap["defaultTableFormat"].(string); ok {
		attributes["default_table_format"] = types.StringValue(defaultTableFormat)
	} else {
		attributes["default_table_format"] = types.StringNull()
	}

	if externalTableCreationEnabled, ok := catalogMap["externalTableCreationEnabled"].(bool); ok {
		attributes["external_table_creation_enabled"] = types.BoolValue(externalTableCreationEnabled)
	} else {
		attributes["external_table_creation_enabled"] = types.BoolNull()
	}

	if externalTableWritesEnabled, ok := catalogMap["externalTableWritesEnabled"].(bool); ok {
		attributes["external_table_writes_enabled"] = types.BoolValue(externalTableWritesEnabled)
	} else {
		attributes["external_table_writes_enabled"] = types.BoolNull()
	}

	if hiveMetastoreHost, ok := catalogMap["hiveMetastoreHost"].(string); ok {
		attributes["hive_metastore_host"] = types.StringValue(hiveMetastoreHost)
	} else {
		attributes["hive_metastore_host"] = types.StringNull()
	}

	if hiveMetastorePort, ok := catalogMap["hiveMetastorePort"].(float64); ok {
		attributes["hive_metastore_port"] = types.Int64Value(int64(hiveMetastorePort))
	} else if hiveMetastorePort, ok := catalogMap["hiveMetastorePort"].(int64); ok {
		attributes["hive_metastore_port"] = types.Int64Value(hiveMetastorePort)
	} else {
		attributes["hive_metastore_port"] = types.Int64Null()
	}

	if metadataMaxStaleness, ok := catalogMap["metadataMaxStalenessSeconds"].(float64); ok {
		attributes["metadata_max_staleness_seconds"] = types.Int64Value(int64(metadataMaxStaleness))
	} else {
		attributes["metadata_max_staleness_seconds"] = types.Int64Null()
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

	if validate, ok := catalogMap["validate"].(bool); ok {
		attributes["validate"] = types.BoolValue(validate)
	} else {
		attributes["validate"] = types.BoolNull()
	}

	catalog, diags := datasource_gcs_catalogs.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating GCS catalog ResultValue: %v", diags))
		return datasource_gcs_catalogs.NewResultValueNull()
	}

	return catalog
}
