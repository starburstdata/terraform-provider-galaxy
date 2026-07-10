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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_data_quality_schedule"
)

var _ datasource.DataSource = (*dataQualityScheduleDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*dataQualityScheduleDataSource)(nil)

func NewDataQualityScheduleDataSource() datasource.DataSource {
	return &dataQualityScheduleDataSource{}
}

type dataQualityScheduleDataSource struct {
	client *client.GalaxyClient
}

func (d *dataQualityScheduleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_quality_schedule"
}

func (d *dataQualityScheduleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_data_quality_schedule.DataQualityScheduleDataSourceSchema(ctx)
}

func (d *dataQualityScheduleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dataQualityScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Provider",
			"The provider has not been properly configured. Please ensure the provider credentials are set.",
		)
		return
	}

	var config datasource_data_quality_schedule.DataQualityScheduleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogId := config.CatalogId.ValueString()
	schemaId := config.SchemaId.ValueString()
	tableId := config.TableId.ValueString()
	tflog.Debug(ctx, "Reading data quality schedule", map[string]interface{}{
		"catalog_id": catalogId,
		"schema_id":  schemaId,
		"table_id":   tableId,
	})

	response, err := d.client.GetDataQualitySchedule(ctx, catalogId, schemaId, tableId)
	if err != nil {
		if client.IsNotFound(err) {
			// No schedule exists for this table - set computed fields to null/empty
			config.DataQualityScheduleId = types.StringNull()
			config.ClusterId = types.StringNull()
			config.CronExpression = types.StringNull()
			config.Enabled = types.BoolNull()
			config.NextExecution = types.StringNull()
			config.RoleId = types.StringNull()
			config.Timezone = types.StringNull()
			checksType := datasource_data_quality_schedule.DataQualityChecksType{
				ObjectType: types.ObjectType{
					AttrTypes: datasource_data_quality_schedule.DataQualityChecksValue{}.AttributeTypes(ctx),
				},
			}
			emptyList, _ := types.ListValueFrom(ctx, checksType, []datasource_data_quality_schedule.DataQualityChecksValue{})
			config.DataQualityChecks = emptyList
			resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading data quality schedule",
			"Could not read data quality schedule: "+err.Error(),
		)
		return
	}

	updateDiags := d.updateModelFromResponse(ctx, &config, response)
	resp.Diagnostics.Append(updateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *dataQualityScheduleDataSource) updateModelFromResponse(ctx context.Context, model *datasource_data_quality_schedule.DataQualityScheduleModel, response map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	if clusterId, ok := response["clusterId"].(string); ok {
		model.ClusterId = types.StringValue(clusterId)
	}
	if cronExpression, ok := response["cronExpression"].(string); ok {
		model.CronExpression = types.StringValue(cronExpression)
	}
	if dataQualityScheduleId, ok := response["dataQualityScheduleId"].(string); ok {
		model.DataQualityScheduleId = types.StringValue(dataQualityScheduleId)
	}
	if enabled, ok := response["enabled"].(bool); ok {
		model.Enabled = types.BoolValue(enabled)
	}
	if nextExecution, ok := response["nextExecution"].(string); ok {
		model.NextExecution = types.StringValue(nextExecution)
	}
	if roleId, ok := response["roleId"].(string); ok {
		model.RoleId = types.StringValue(roleId)
	}
	if timezone, ok := response["timezone"].(string); ok {
		model.Timezone = types.StringValue(timezone)
	}

	// Map data quality checks list
	checksType := datasource_data_quality_schedule.DataQualityChecksType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_data_quality_schedule.DataQualityChecksValue{}.AttributeTypes(ctx),
		},
	}
	if checks, ok := response["dataQualityChecks"].([]interface{}); ok {
		checksList := make([]datasource_data_quality_schedule.DataQualityChecksValue, 0, len(checks))
		for _, checkInterface := range checks {
			if checkMap, ok := checkInterface.(map[string]interface{}); ok {
				attributeTypes := datasource_data_quality_schedule.DataQualityChecksValue{}.AttributeTypes(ctx)
				attributes := map[string]attr.Value{}

				if dataQualityCheckId, ok := checkMap["dataQualityCheckId"].(string); ok {
					attributes["data_quality_check_id"] = types.StringValue(dataQualityCheckId)
				} else {
					attributes["data_quality_check_id"] = types.StringNull()
				}
				if name, ok := checkMap["name"].(string); ok {
					attributes["name"] = types.StringValue(name)
				} else {
					attributes["name"] = types.StringNull()
				}

				checkValue, d := datasource_data_quality_schedule.NewDataQualityChecksValue(attributeTypes, attributes)
				if d.HasError() {
					diags.Append(d...)
					continue
				}
				checksList = append(checksList, checkValue)
			}
		}
		listValue, d := types.ListValueFrom(ctx, checksType, checksList)
		if d.HasError() {
			diags.Append(d...)
			model.DataQualityChecks = types.ListNull(checksType)
		} else {
			model.DataQualityChecks = listValue
		}
	} else {
		emptyList, _ := types.ListValueFrom(ctx, checksType, []datasource_data_quality_schedule.DataQualityChecksValue{})
		model.DataQualityChecks = emptyList
	}

	return diags
}
