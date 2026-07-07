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
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_data_quality_checks"
)

var _ datasource.DataSource = (*dataQualityChecksDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*dataQualityChecksDataSource)(nil)

func NewDataQualityChecksDataSource() datasource.DataSource {
	return &dataQualityChecksDataSource{}
}

type dataQualityChecksDataSource struct {
	client *client.GalaxyClient
}

func (d *dataQualityChecksDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_quality_checks"
}

func (d *dataQualityChecksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_data_quality_checks.DataQualityChecksDataSourceSchema(ctx)
}

func (d *dataQualityChecksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dataQualityChecksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_data_quality_checks.DataQualityChecksModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading data quality checks")

	// Build query path with optional filters
	path := "/public/api/v1/dataQualityCheck"
	params := url.Values{}
	if !config.CatalogId.IsNull() && !config.CatalogId.IsUnknown() && config.CatalogId.ValueString() != "" {
		params.Add("catalogId", config.CatalogId.ValueString())
	}
	if !config.SchemaId.IsNull() && !config.SchemaId.IsUnknown() && config.SchemaId.ValueString() != "" {
		params.Add("schemaId", config.SchemaId.ValueString())
	}
	if !config.TableId.IsNull() && !config.TableId.IsUnknown() && config.TableId.ValueString() != "" {
		params.Add("tableId", config.TableId.ValueString())
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	allResults, err := d.client.GetAllPaginatedResults(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading data quality checks",
			"Could not read data quality checks: "+err.Error(),
		)
		return
	}

	var resultMaps []map[string]interface{}
	for _, resultInterface := range allResults {
		if resultMap, ok := resultInterface.(map[string]interface{}); ok {
			resultMaps = append(resultMaps, resultMap)
		}
	}

	if len(resultMaps) > 0 {
		results, err := d.mapResults(ctx, resultMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping data quality checks response",
				"Could not map data quality checks response: "+err.Error(),
			)
			return
		}
		config.Result = results
	} else {
		elementType := datasource_data_quality_checks.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_data_quality_checks.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_data_quality_checks.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *dataQualityChecksDataSource) mapResults(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	items := make([]datasource_data_quality_checks.ResultValue, 0)

	for _, itemMap := range result {
		item := d.mapSingleResult(ctx, itemMap)
		items = append(items, item)
	}

	elementType := datasource_data_quality_checks.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_data_quality_checks.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, items)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *dataQualityChecksDataSource) mapSingleResult(ctx context.Context, itemMap map[string]interface{}) datasource_data_quality_checks.ResultValue {
	attributeTypes := datasource_data_quality_checks.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	if catalogId, ok := itemMap["catalogId"].(string); ok {
		attributes["catalog_id"] = types.StringValue(catalogId)
	} else {
		attributes["catalog_id"] = types.StringNull()
	}

	if dataQualityCheckId, ok := itemMap["dataQualityCheckId"].(string); ok {
		attributes["data_quality_check_id"] = types.StringValue(dataQualityCheckId)
	} else {
		attributes["data_quality_check_id"] = types.StringNull()
	}

	if description, ok := itemMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	if name, ok := itemMap["name"].(string); ok {
		attributes["name"] = types.StringValue(name)
	} else {
		attributes["name"] = types.StringNull()
	}

	if schemaId, ok := itemMap["schemaId"].(string); ok {
		attributes["schema_id"] = types.StringValue(schemaId)
	} else {
		attributes["schema_id"] = types.StringNull()
	}

	if tableId, ok := itemMap["tableId"].(string); ok {
		attributes["table_id"] = types.StringValue(tableId)
	} else {
		attributes["table_id"] = types.StringNull()
	}

	item, diags := datasource_data_quality_checks.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating data quality check ResultValue: %v", diags))
		return datasource_data_quality_checks.NewResultValueNull()
	}

	return item
}
