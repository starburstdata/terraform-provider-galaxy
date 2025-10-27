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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_column_masks"
)

var _ datasource.DataSource = (*column_masksDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*column_masksDataSource)(nil)

func NewColumnMasksDataSource() datasource.DataSource {
	return &column_masksDataSource{}
}

type column_masksDataSource struct {
	client *client.GalaxyClient
}

func (d *column_masksDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_column_masks"
}

func (d *column_masksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_column_masks.ColumnMasksDataSourceSchema(ctx)
}

func (d *column_masksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *column_masksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_column_masks.ColumnMasksModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading column_masks with automatic pagination")

	// Use automatic pagination to get ALL column masks across all pages
	allColumnMasks, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/columnMask")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading column_masks",
			"Could not read column_masks: "+err.Error(),
		)
		return
	}

	// Convert []interface{} to []map[string]interface{} for mapping
	var columnMaskMaps []map[string]interface{}
	for _, columnMaskInterface := range allColumnMasks {
		if columnMaskMap, ok := columnMaskInterface.(map[string]interface{}); ok {
			columnMaskMaps = append(columnMaskMaps, columnMaskMap)
		}
	}

	// Map API response to model
	if len(columnMaskMaps) > 0 {
		columnMasks, err := d.mapColumnMasksResult(ctx, columnMaskMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping column_masks response",
				"Could not map column_masks response: "+err.Error(),
			)
			return
		}
		config.Result = columnMasks
	} else {
		elementType := datasource_column_masks.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_column_masks.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_column_masks.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *column_masksDataSource) mapColumnMasksResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	columnMasks := make([]datasource_column_masks.ResultValue, 0)

	for _, columnMaskMap := range result {
		columnMask := d.mapSingleColumnMask(ctx, columnMaskMap)
		columnMasks = append(columnMasks, columnMask)
	}

	elementType := datasource_column_masks.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_column_masks.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, columnMasks)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *column_masksDataSource) mapSingleColumnMask(ctx context.Context, columnMaskMap map[string]interface{}) datasource_column_masks.ResultValue {
	attributeTypes := datasource_column_masks.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map column mask ID
	if columnMaskId, ok := columnMaskMap["columnMaskId"].(string); ok {
		attributes["column_mask_id"] = types.StringValue(columnMaskId)
	} else {
		attributes["column_mask_id"] = types.StringNull()
	}

	// Map name
	if name, ok := columnMaskMap["name"].(string); ok {
		attributes["name"] = types.StringValue(name)
	} else {
		attributes["name"] = types.StringNull()
	}

	// Map description
	if description, ok := columnMaskMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	// Map column mask type
	if columnMaskType, ok := columnMaskMap["columnMaskType"].(string); ok {
		attributes["column_mask_type"] = types.StringValue(columnMaskType)
	} else if columnMaskType, ok := columnMaskMap["type"].(string); ok {
		attributes["column_mask_type"] = types.StringValue(columnMaskType)
	} else {
		attributes["column_mask_type"] = types.StringNull()
	}

	// Map expression
	if expression, ok := columnMaskMap["expression"].(string); ok {
		attributes["expression"] = types.StringValue(expression)
	} else {
		attributes["expression"] = types.StringNull()
	}

	// Map created
	if created, ok := columnMaskMap["created"].(string); ok {
		attributes["created"] = types.StringValue(created)
	} else {
		attributes["created"] = types.StringNull()
	}

	// Map modified
	if modified, ok := columnMaskMap["modified"].(string); ok {
		attributes["modified"] = types.StringValue(modified)
	} else {
		attributes["modified"] = types.StringNull()
	}

	// Create the ResultValue using the constructor
	columnMask, diags := datasource_column_masks.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating column mask ResultValue: %v", diags))
		return datasource_column_masks.NewResultValueNull()
	}

	return columnMask
}
