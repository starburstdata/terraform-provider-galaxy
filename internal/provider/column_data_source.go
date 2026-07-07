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
	columnID := config.ColumnId.ValueString()
	tflog.Debug(ctx, "Reading column", map[string]interface{}{
		"catalog_id": catalogID,
		"schema_id":  schemaID,
		"table_id":   tableID,
		"column_id":  columnID,
	})

	response, err := d.client.ListColumns(ctx, catalogID, schemaID, tableID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading column",
			"Could not read column: "+err.Error(),
		)
		return
	}

	// Find the specific column by ID in the results
	var columnData map[string]interface{}
	if resultArray, ok := response["result"].([]interface{}); ok {
		for _, item := range resultArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if id, ok := itemMap["columnId"].(string); ok && id == columnID {
					columnData = itemMap
					break
				}
			}
		}
	}

	if columnData == nil {
		resp.Diagnostics.AddError(
			"Column not found",
			fmt.Sprintf("Could not find column with ID %s in catalog %s, schema %s, table %s", columnID, catalogID, schemaID, tableID),
		)
		return
	}

	d.updateModelFromResponse(ctx, &config, columnData)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *columnDataSource) updateModelFromResponse(ctx context.Context, model *datasource_column.ColumnModel, response map[string]interface{}) {
	if columnDefault, ok := response["columnDefault"].(string); ok {
		model.ColumnDefault = types.StringValue(columnDefault)
	} else {
		model.ColumnDefault = types.StringNull()
	}

	if dataType, ok := response["dataType"].(string); ok {
		model.DataType = types.StringValue(dataType)
	} else {
		model.DataType = types.StringNull()
	}

	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else {
		model.Description = types.StringNull()
	}

	if nullable, ok := response["nullable"].(bool); ok {
		model.Nullable = types.BoolValue(nullable)
	} else {
		model.Nullable = types.BoolNull()
	}

	// Map tags list
	tagElementType := datasource_column.TagsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_column.TagsValue{}.AttributeTypes(ctx),
		},
	}

	if tags, ok := response["tags"].([]interface{}); ok && len(tags) > 0 {
		tagList := make([]datasource_column.TagsValue, 0, len(tags))
		for _, tag := range tags {
			if tagMap, ok := tag.(map[string]interface{}); ok {
				tagAttrs := map[string]attr.Value{
					"tag_id": types.StringNull(),
					"name":   types.StringNull(),
				}
				if tagId, ok := tagMap["tagId"].(string); ok {
					tagAttrs["tag_id"] = types.StringValue(tagId)
				}
				if name, ok := tagMap["name"].(string); ok {
					tagAttrs["name"] = types.StringValue(name)
				}

				tagValue, diags := datasource_column.NewTagsValue(
					datasource_column.TagsValue{}.AttributeTypes(ctx),
					tagAttrs,
				)
				if diags.HasError() {
					tflog.Error(ctx, fmt.Sprintf("Error creating tag value: %v", diags))
					continue
				}
				tagList = append(tagList, tagValue)
			}
		}
		listValue, diags := types.ListValueFrom(ctx, tagElementType, tagList)
		if !diags.HasError() {
			model.Tags = listValue
		} else {
			model.Tags, _ = types.ListValueFrom(ctx, tagElementType, []datasource_column.TagsValue{})
		}
	} else {
		model.Tags, _ = types.ListValueFrom(ctx, tagElementType, []datasource_column.TagsValue{})
	}
}
