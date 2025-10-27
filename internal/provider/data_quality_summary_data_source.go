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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_data_quality_summary"
)

var _ datasource.DataSource = (*dataQualitySummaryDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*dataQualitySummaryDataSource)(nil)

func NewDataQualitySummaryDataSource() datasource.DataSource {
	return &dataQualitySummaryDataSource{}
}

type dataQualitySummaryDataSource struct {
	client *client.GalaxyClient
}

func (d *dataQualitySummaryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_quality_summary"
}

func (d *dataQualitySummaryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_data_quality_summary.DataQualitySummaryDataSourceSchema(ctx)
}

func (d *dataQualitySummaryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dataQualitySummaryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_data_quality_summary.DataQualitySummaryModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := config.CatalogId.ValueString()
	schemaID := config.SchemaId.ValueString()
	tflog.Debug(ctx, "Reading data quality summary", map[string]interface{}{"catalogId": catalogID, "schemaId": schemaID})

	response, err := d.client.GetDataQualitySummary(ctx, catalogID, schemaID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading data quality summary",
			"Could not read data quality summary for catalog "+catalogID+" schema "+schemaID+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *dataQualitySummaryDataSource) updateModelFromResponse(ctx context.Context, model *datasource_data_quality_summary.DataQualitySummaryModel, response map[string]interface{}) {
	// The catalogId and schemaId are already set from the configuration

	// Map category_counts
	if categoryCounts, ok := response["categoryCounts"].([]interface{}); ok {
		categoryList := make([]datasource_data_quality_summary.CategoryCountsValue, 0, len(categoryCounts))
		for _, item := range categoryCounts {
			if itemMap, ok := item.(map[string]interface{}); ok {
				categoryCount := datasource_data_quality_summary.CategoryCountsValue{
					Category:              types.StringValue(getStringFromMap(itemMap, "category")),
					FailedEvaluations:     types.Int64Value(getInt64FromMap(itemMap, "failedEvaluations")),
					SuccessfulEvaluations: types.Int64Value(getInt64FromMap(itemMap, "successfulEvaluations")),
					TotalEvaluations:      types.Int64Value(getInt64FromMap(itemMap, "totalEvaluations")),
				}
				categoryList = append(categoryList, categoryCount)
			}
		}
		categoryListValue, diags := types.ListValueFrom(ctx, datasource_data_quality_summary.CategoryCountsType{}.ValueType(ctx).Type(ctx), categoryList)
		if diags.HasError() {
			tflog.Error(ctx, "Error creating category counts list", map[string]interface{}{"errors": diags})
			model.CategoryCounts = types.ListNull(datasource_data_quality_summary.CategoryCountsType{}.ValueType(ctx).Type(ctx))
		} else {
			model.CategoryCounts = categoryListValue
		}
	} else {
		model.CategoryCounts = types.ListNull(datasource_data_quality_summary.CategoryCountsType{})
	}

	// Map daily_summaries
	if dailySummaries, ok := response["dailySummaries"].([]interface{}); ok {
		dailyList := make([]datasource_data_quality_summary.DailySummariesValue, 0, len(dailySummaries))
		for _, item := range dailySummaries {
			if itemMap, ok := item.(map[string]interface{}); ok {
				dailySummary := datasource_data_quality_summary.DailySummariesValue{
					Day:                   types.StringValue(getStringFromMap(itemMap, "day")),
					FailedEvaluations:     types.Int64Value(getInt64FromMap(itemMap, "failedEvaluations")),
					SuccessfulEvaluations: types.Int64Value(getInt64FromMap(itemMap, "successfulEvaluations")),
					TotalEvaluations:      types.Int64Value(getInt64FromMap(itemMap, "totalEvaluations")),
				}
				dailyList = append(dailyList, dailySummary)
			}
		}
		dailyListValue, diags := types.ListValueFrom(ctx, datasource_data_quality_summary.DailySummariesType{}.ValueType(ctx).Type(ctx), dailyList)
		if diags.HasError() {
			tflog.Error(ctx, "Error creating daily summaries list", map[string]interface{}{"errors": diags})
			model.DailySummaries = types.ListNull(datasource_data_quality_summary.DailySummariesType{}.ValueType(ctx).Type(ctx))
		} else {
			model.DailySummaries = dailyListValue
		}
	} else {
		model.DailySummaries = types.ListNull(datasource_data_quality_summary.DailySummariesType{})
	}

	// Map severity_counts
	if severityCounts, ok := response["severityCounts"].([]interface{}); ok {
		severityList := make([]datasource_data_quality_summary.SeverityCountsValue, 0, len(severityCounts))
		for _, item := range severityCounts {
			if itemMap, ok := item.(map[string]interface{}); ok {
				severityCount := datasource_data_quality_summary.SeverityCountsValue{
					Severity:              types.StringValue(getStringFromMap(itemMap, "severity")),
					FailedEvaluations:     types.Int64Value(getInt64FromMap(itemMap, "failedEvaluations")),
					SuccessfulEvaluations: types.Int64Value(getInt64FromMap(itemMap, "successfulEvaluations")),
					TotalEvaluations:      types.Int64Value(getInt64FromMap(itemMap, "totalEvaluations")),
				}
				severityList = append(severityList, severityCount)
			}
		}
		severityListValue, diags := types.ListValueFrom(ctx, datasource_data_quality_summary.SeverityCountsType{}.ValueType(ctx).Type(ctx), severityList)
		if diags.HasError() {
			tflog.Error(ctx, "Error creating severity counts list", map[string]interface{}{"errors": diags})
			model.SeverityCounts = types.ListNull(datasource_data_quality_summary.SeverityCountsType{}.ValueType(ctx).Type(ctx))
		} else {
			model.SeverityCounts = severityListValue
		}
	} else {
		model.SeverityCounts = types.ListNull(datasource_data_quality_summary.SeverityCountsType{})
	}

	// Map table_summaries
	if tableSummaries, ok := response["tableSummaries"].([]interface{}); ok {
		tableList := make([]datasource_data_quality_summary.TableSummariesValue, 0, len(tableSummaries))
		for _, item := range tableSummaries {
			if itemMap, ok := item.(map[string]interface{}); ok {
				tableSummary := datasource_data_quality_summary.TableSummariesValue{
					TableId:               types.StringValue(getStringFromMap(itemMap, "tableId")),
					EvaluatedChecks:       types.Int64Value(getInt64FromMap(itemMap, "evaluatedChecks")),
					FailedChecks:          types.Int64Value(getInt64FromMap(itemMap, "failedChecks")),
					NotYetEvaluatedChecks: types.Int64Value(getInt64FromMap(itemMap, "notYetEvaluatedChecks")),
					NumberOfChecks:        types.Int64Value(getInt64FromMap(itemMap, "numberOfChecks")),
					SuccessfulChecks:      types.Int64Value(getInt64FromMap(itemMap, "successfulChecks")),
				}
				tableList = append(tableList, tableSummary)
			}
		}
		tableListValue, diags := types.ListValueFrom(ctx, datasource_data_quality_summary.TableSummariesType{}.ValueType(ctx).Type(ctx), tableList)
		if diags.HasError() {
			tflog.Error(ctx, "Error creating table summaries list", map[string]interface{}{"errors": diags})
			model.TableSummaries = types.ListNull(datasource_data_quality_summary.TableSummariesType{}.ValueType(ctx).Type(ctx))
		} else {
			model.TableSummaries = tableListValue
		}
	} else {
		model.TableSummaries = types.ListNull(datasource_data_quality_summary.TableSummariesType{}.ValueType(ctx).Type(ctx))
	}
}

// Helper functions to safely extract values from maps
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getInt64FromMap(m map[string]interface{}, key string) int64 {
	if val, ok := m[key].(float64); ok {
		return int64(val)
	}
	if val, ok := m[key].(int64); ok {
		return val
	}
	if val, ok := m[key].(int); ok {
		return int64(val)
	}
	return 0
}
