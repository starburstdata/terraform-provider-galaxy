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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_column"
)

var _ datasource.DataSource = (*columnDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*columnDataSource)(nil)

func NewColumnDataSource() datasource.DataSource {
	return &columnDataSource{}
}

type columnDataSource struct {
	client *client.GalaxyClient
}

func (d *columnDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_column"
}

func (d *columnDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_column.ColumnDataSourceSchema(ctx)
}

func (d *columnDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *columnDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_column.ColumnModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := config.CatalogId.ValueString()
	schemaID := config.SchemaId.ValueString()
	tableID := config.TableId.ValueString()
	tflog.Debug(ctx, "Reading columns", map[string]interface{}{"catalogId": catalogID, "schemaId": schemaID, "tableId": tableID})

	response, err := d.client.ListColumns(ctx, catalogID, schemaID, tableID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading columns",
			"Could not read columns for catalog "+catalogID+" schema "+schemaID+" table "+tableID+": "+err.Error(),
		)
		return
	}

	// Debug: Log the raw response to help diagnose issues
	tflog.Info(ctx, "ListColumns API response", map[string]interface{}{
		"catalogId": catalogID,
		"schemaId":  schemaID,
		"tableId":   tableID,
		"response":  fmt.Sprintf("%+v", response),
	})

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *columnDataSource) updateModelFromResponse(ctx context.Context, model *datasource_column.ColumnModel, response map[string]interface{}) {
	// The catalogId, schemaId, and tableId are already set from the configuration
	// Ensure they remain in known state (they should already be set from config)

	// Define element types for custom types
	tagElementType := datasource_column.TagsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_column.TagsValue{}.AttributeTypes(ctx),
		},
	}

	resultElementType := datasource_column.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_column.ResultValue{}.AttributeTypes(ctx),
		},
	}

	// Map the result array
	if resultArray, ok := response["result"].([]interface{}); ok {
		resultList := make([]datasource_column.ResultValue, 0, len(resultArray))
		for _, item := range resultArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				// Build tags list first
				var tagListValue types.List
				if tags, ok := itemMap["tags"].([]interface{}); ok && len(tags) > 0 {
					tagList := make([]datasource_column.TagsValue, 0, len(tags))
					for _, tag := range tags {
						if tagMap, ok := tag.(map[string]interface{}); ok {
							tagAttrs := map[string]attr.Value{
								"tag_id": types.StringValue(getStringFromMap(tagMap, "tagId")),
								"name":   types.StringValue(getStringFromMap(tagMap, "name")),
							}
							tagItem, tagDiags := datasource_column.NewTagsValue(
								datasource_column.TagsValue{}.AttributeTypes(ctx),
								tagAttrs,
							)
							if tagDiags.HasError() {
								tflog.Error(ctx, "Error creating tag value", map[string]interface{}{"errors": tagDiags})
								continue
							}
							tagList = append(tagList, tagItem)
						}
					}
					var diags diag.Diagnostics
					tagListValue, diags = types.ListValueFrom(ctx, tagElementType, tagList)
					if diags.HasError() {
						tflog.Error(ctx, "Error creating tags list", map[string]interface{}{"errors": diags})
						tagListValue = types.ListNull(tagElementType)
					}
				} else {
					// Set to empty list instead of null
					var diags diag.Diagnostics
					tagListValue, diags = types.ListValueFrom(ctx, tagElementType, []datasource_column.TagsValue{})
					if diags.HasError() {
						tagListValue = types.ListNull(tagElementType)
					}
				}

				// Create ResultValue using the proper constructor
				resultAttrs := map[string]attr.Value{
					"column_id":      types.StringValue(getStringFromMap(itemMap, "columnId")),
					"column_default": types.StringValue(getStringFromMap(itemMap, "columnDefault")),
					"data_type":      types.StringValue(getStringFromMap(itemMap, "dataType")),
					"description":    types.StringValue(getStringFromMap(itemMap, "description")),
					"nullable":       types.BoolValue(getBoolFromMap(itemMap, "nullable")),
					"tags":           tagListValue,
				}

				resultItem, resultDiags := datasource_column.NewResultValue(
					datasource_column.ResultValue{}.AttributeTypes(ctx),
					resultAttrs,
				)
				if resultDiags.HasError() {
					tflog.Error(ctx, "Error creating result value", map[string]interface{}{"errors": resultDiags})
					continue
				}
				resultList = append(resultList, resultItem)
			}
		}
		resultListValue, diags := types.ListValueFrom(ctx, resultElementType, resultList)
		if diags.HasError() {
			tflog.Error(ctx, "Error creating result list", map[string]interface{}{"errors": diags})
			model.Result = types.ListNull(resultElementType)
		} else {
			model.Result = resultListValue
		}
	} else {
		// No results or result is not an array - set to empty list instead of null
		emptyList, diags := types.ListValueFrom(ctx, resultElementType, []datasource_column.ResultValue{})
		if diags.HasError() {
			tflog.Error(ctx, "Error creating empty result list", map[string]interface{}{"errors": diags})
			model.Result = types.ListNull(resultElementType)
		} else {
			model.Result = emptyList
		}
	}
}
