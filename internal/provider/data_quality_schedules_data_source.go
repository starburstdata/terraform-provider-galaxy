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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_data_quality_schedules"
)

var _ datasource.DataSource = (*dataQualitySchedulesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*dataQualitySchedulesDataSource)(nil)

func NewDataQualitySchedulesDataSource() datasource.DataSource {
	return &dataQualitySchedulesDataSource{}
}

type dataQualitySchedulesDataSource struct {
	client *client.GalaxyClient
}

func (d *dataQualitySchedulesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_quality_schedules"
}

func (d *dataQualitySchedulesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_data_quality_schedules.DataQualitySchedulesDataSourceSchema(ctx)
}

func (d *dataQualitySchedulesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dataQualitySchedulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_data_quality_schedules.DataQualitySchedulesModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := config.CatalogId.ValueString()
	schemaID := config.SchemaId.ValueString()
	tableID := config.TableId.ValueString()

	tflog.Debug(ctx, "Reading data_quality_schedules", map[string]interface{}{
		"catalog_id": catalogID,
		"schema_id":  schemaID,
		"table_id":   tableID,
	})

	response, err := d.client.GetDataQualitySchedule(ctx, catalogID, schemaID, tableID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading data_quality_schedules",
			"Could not read data quality schedule for catalog "+catalogID+", schema "+schemaID+", table "+tableID+": "+err.Error(),
		)
		return
	}

	// Map scalar fields
	if v, ok := response["dataQualityScheduleId"].(string); ok {
		config.DataQualityScheduleId = types.StringValue(v)
	} else {
		config.DataQualityScheduleId = types.StringNull()
	}
	if v, ok := response["clusterId"].(string); ok {
		config.ClusterId = types.StringValue(v)
	} else {
		config.ClusterId = types.StringNull()
	}
	if v, ok := response["cronExpression"].(string); ok {
		config.CronExpression = types.StringValue(v)
	} else {
		config.CronExpression = types.StringNull()
	}
	if v, ok := response["enabled"].(bool); ok {
		config.Enabled = types.BoolValue(v)
	} else {
		config.Enabled = types.BoolNull()
	}
	if v, ok := response["nextExecution"].(string); ok {
		config.NextExecution = types.StringValue(v)
	} else {
		config.NextExecution = types.StringNull()
	}
	if v, ok := response["roleId"].(string); ok {
		config.RoleId = types.StringValue(v)
	} else {
		config.RoleId = types.StringNull()
	}
	if v, ok := response["timezone"].(string); ok {
		config.Timezone = types.StringValue(v)
	} else {
		config.Timezone = types.StringNull()
	}

	// Map data_quality_checks list
	checksElementType := datasource_data_quality_schedules.DataQualityChecksType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_data_quality_schedules.DataQualityChecksValue{}.AttributeTypes(ctx),
		},
	}

	var checkItems []interface{}
	if items, ok := response["dataQualityChecks"].([]interface{}); ok {
		checkItems = items
	} else if items, ok := response["items"].([]interface{}); ok {
		checkItems = items
	} else if items, ok := response["result"].([]interface{}); ok {
		checkItems = items
	}

	checksAttributeTypes := datasource_data_quality_schedules.DataQualityChecksValue{}.AttributeTypes(ctx)
	if len(checkItems) > 0 {
		var checkElements []attr.Value
		for _, itemRaw := range checkItems {
			if item, ok := itemRaw.(map[string]interface{}); ok {
				checkAttributes := map[string]attr.Value{}
				if v, ok := item["dataQualityCheckId"].(string); ok {
					checkAttributes["data_quality_check_id"] = types.StringValue(v)
				} else {
					checkAttributes["data_quality_check_id"] = types.StringNull()
				}
				if v, ok := item["name"].(string); ok {
					checkAttributes["name"] = types.StringValue(v)
				} else {
					checkAttributes["name"] = types.StringNull()
				}
				checkValue, diags := datasource_data_quality_schedules.NewDataQualityChecksValue(checksAttributeTypes, checkAttributes)
				resp.Diagnostics.Append(diags...)
				if diags.HasError() {
					continue
				}
				checkElements = append(checkElements, checkValue)
			}
		}
		config.DataQualityChecks, _ = types.ListValue(checksElementType, checkElements)
	} else {
		config.DataQualityChecks, _ = types.ListValue(checksElementType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
