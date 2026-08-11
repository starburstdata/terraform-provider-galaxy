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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_columns"
)

var _ datasource.DataSource = (*columnsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*columnsDataSource)(nil)

func NewColumnsDataSource() datasource.DataSource {
	return &columnsDataSource{}
}

type columnsDataSource struct {
	client *client.GalaxyClient
}

func (d *columnsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_columns"
}

func (d *columnsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_columns.ColumnsDataSourceSchema(ctx)
}

func (d *columnsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *columnsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_columns.ColumnsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := config.CatalogId.ValueString()
	schemaID := config.SchemaId.ValueString()
	tableID := config.TableId.ValueString()

	tflog.Debug(ctx, "Reading columns", map[string]interface{}{"catalog_id": catalogID, "schema_id": schemaID, "table_id": tableID})

	response, err := d.client.ListColumns(ctx, catalogID, schemaID, tableID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading columns",
			"Could not read columns: "+err.Error(),
		)
		return
	}

	// Map results
	elementType := datasource_columns.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_columns.ResultValue{}.AttributeTypes(ctx),
		},
	}

	var results []interface{}
	if items, ok := response["items"].([]interface{}); ok {
		results = items
	} else if arr, ok := response["result"].([]interface{}); ok {
		results = arr
	}

	if len(results) > 0 {
		var resultElements []attr.Value
		for _, itemRaw := range results {
			if item, ok := itemRaw.(map[string]interface{}); ok {
				resultValue, diags := d.mapSingleResult(ctx, item)
				resp.Diagnostics.Append(diags...)
				if diags.HasError() {
					continue
				}
				resultElements = append(resultElements, resultValue)
			}
		}
		config.Result, _ = types.ListValue(elementType, resultElements)
	} else {
		config.Result, _ = types.ListValue(elementType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *columnsDataSource) mapSingleResult(ctx context.Context, item map[string]interface{}) (datasource_columns.ResultValue, diag.Diagnostics) {
	attributeTypes := datasource_columns.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	if v, ok := item["columnDefault"].(string); ok {
		attributes["column_default"] = types.StringValue(v)
	} else {
		attributes["column_default"] = types.StringNull()
	}
	if v, ok := item["columnId"].(string); ok {
		attributes["column_id"] = types.StringValue(v)
	} else {
		attributes["column_id"] = types.StringNull()
	}
	if v, ok := item["dataType"].(string); ok {
		attributes["data_type"] = types.StringValue(v)
	} else {
		attributes["data_type"] = types.StringNull()
	}
	if v, ok := item["description"].(string); ok {
		attributes["description"] = types.StringValue(v)
	} else {
		attributes["description"] = types.StringNull()
	}
	if v, ok := item["nullable"].(bool); ok {
		attributes["nullable"] = types.BoolValue(v)
	} else {
		attributes["nullable"] = types.BoolNull()
	}

	// Map tags list
	tagsElementType := datasource_columns.TagsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_columns.TagsValue{}.AttributeTypes(ctx),
		},
	}
	tagsAttributeTypes := datasource_columns.TagsValue{}.AttributeTypes(ctx)
	if tags, ok := item["tags"].([]interface{}); ok && len(tags) > 0 {
		var tagElements []attr.Value
		for _, tagRaw := range tags {
			if tagMap, ok := tagRaw.(map[string]interface{}); ok {
				tagAttributes := map[string]attr.Value{}
				if v, ok := tagMap["name"].(string); ok {
					tagAttributes["name"] = types.StringValue(v)
				} else {
					tagAttributes["name"] = types.StringNull()
				}
				if v, ok := tagMap["tagId"].(string); ok {
					tagAttributes["tag_id"] = types.StringValue(v)
				} else {
					tagAttributes["tag_id"] = types.StringNull()
				}
				tagValue, diags := datasource_columns.NewTagsValue(tagsAttributeTypes, tagAttributes)
				if diags.HasError() {
					return datasource_columns.ResultValue{}, diags
				}
				tagElements = append(tagElements, tagValue)
			}
		}
		attributes["tags"], _ = types.ListValue(tagsElementType, tagElements)
	} else {
		attributes["tags"], _ = types.ListValue(tagsElementType, []attr.Value{})
	}

	return datasource_columns.NewResultValue(attributeTypes, attributes)
}
