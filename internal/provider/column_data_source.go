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

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *columnDataSource) updateModelFromResponse(ctx context.Context, model *datasource_column.ColumnModel, response map[string]interface{}) {
	// The catalogId, schemaId, and tableId are already set from the configuration
	// Ensure they remain in known state (they should already be set from config)

	// Map the result array
	if resultArray, ok := response["result"].([]interface{}); ok {
		resultList := make([]attr.Value, 0, len(resultArray))
		for _, item := range resultArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				resultItem := datasource_column.ResultValue{
					ColumnId:      types.StringValue(getStringFromMap(itemMap, "columnId")),
					ColumnDefault: types.StringValue(getStringFromMap(itemMap, "columnDefault")),
					DataType:      types.StringValue(getStringFromMap(itemMap, "dataType")),
					Description:   types.StringValue(getStringFromMap(itemMap, "description")),
					Nullable:      types.BoolValue(getBoolFromMap(itemMap, "nullable")),
				}

				// Map tags with proper error handling
				if tags, ok := itemMap["tags"].([]interface{}); ok && len(tags) > 0 {
					tagList := make([]attr.Value, 0, len(tags))
					for _, tag := range tags {
						if tagMap, ok := tag.(map[string]interface{}); ok {
							tagItem := datasource_column.TagsValue{
								TagId: types.StringValue(getStringFromMap(tagMap, "tagId")),
								Name:  types.StringValue(getStringFromMap(tagMap, "name")),
							}
							objValue, diags := tagItem.ToObjectValue(ctx)
							if diags.HasError() {
								tflog.Error(ctx, "Error converting tag to object value", map[string]interface{}{"errors": diags})
								continue
							}
							tagList = append(tagList, objValue)
						}
					}
					tagListValue, diags := types.ListValue(
						datasource_column.TagsType{}.ValueType(ctx).Type(ctx),
						tagList,
					)
					if diags.HasError() {
						tflog.Error(ctx, "Error creating tags list", map[string]interface{}{"errors": diags})
						resultItem.Tags = types.ListNull(
							datasource_column.TagsType{}.ValueType(ctx).Type(ctx),
						)
					} else {
						resultItem.Tags = tagListValue
					}
				} else {
					resultItem.Tags = types.ListNull(
						datasource_column.TagsType{}.ValueType(ctx).Type(ctx),
					)
				}

				objValue, diags := resultItem.ToObjectValue(ctx)
				if diags.HasError() {
					tflog.Error(ctx, "Error converting column result to object value", map[string]interface{}{"errors": diags})
					continue
				}
				resultList = append(resultList, objValue)
			}
		}
		resultListValue, diags := types.ListValue(
			datasource_column.ResultType{}.ValueType(ctx).Type(ctx),
			resultList,
		)
		if diags.HasError() {
			tflog.Error(ctx, "Error creating result list", map[string]interface{}{"errors": diags})
			model.Result = types.ListNull(
				datasource_column.ResultType{}.ValueType(ctx).Type(ctx),
			)
		} else {
			model.Result = resultListValue
		}
	} else {
		// No results or result is not an array - set to empty list instead of null
		emptyList, diags := types.ListValue(
			datasource_column.ResultType{}.ValueType(ctx).Type(ctx),
			[]attr.Value{},
		)
		if diags.HasError() {
			tflog.Error(ctx, "Error creating empty result list", map[string]interface{}{"errors": diags})
			model.Result = types.ListNull(
				datasource_column.ResultType{}.ValueType(ctx).Type(ctx),
			)
		} else {
			model.Result = emptyList
		}
	}
}
