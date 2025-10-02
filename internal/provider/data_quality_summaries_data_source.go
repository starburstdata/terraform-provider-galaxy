package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_data_quality_summaries"
)

var _ datasource.DataSource = (*dataQualitySummariesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*dataQualitySummariesDataSource)(nil)

func NewDataQualitySummariesDataSource() datasource.DataSource {
	return &dataQualitySummariesDataSource{}
}

type dataQualitySummariesDataSource struct {
	client *client.GalaxyClient
}

func (d *dataQualitySummariesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_quality_summaries"
}

func (d *dataQualitySummariesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_data_quality_summaries.DataQualitySummariesDataSourceSchema(ctx)
}

func (d *dataQualitySummariesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dataQualitySummariesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_data_quality_summaries.DataQualitySummariesModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading data quality summaries")

	response, err := d.client.ListDataQualitySummaries(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading data quality summaries",
			"Could not read data quality summaries: "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *dataQualitySummariesDataSource) updateModelFromResponse(ctx context.Context, model *datasource_data_quality_summaries.DataQualitySummariesModel, response map[string]interface{}) {
	// Map catalog_summaries
	if catalogSummaries, ok := response["catalogSummaries"].([]interface{}); ok {
		catalogList := make([]datasource_data_quality_summaries.CatalogSummariesValue, 0, len(catalogSummaries))
		for _, item := range catalogSummaries {
			if itemMap, ok := item.(map[string]interface{}); ok {
				catalogSummary := datasource_data_quality_summaries.CatalogSummariesValue{
					CatalogId:             types.StringValue(getStringFromMap(itemMap, "catalogId")),
					CatalogName:           types.StringValue(getStringFromMap(itemMap, "catalogName")),
					EvaluatedChecks:       types.Int64Value(getInt64FromMap(itemMap, "evaluatedChecks")),
					FailedChecks:          types.Int64Value(getInt64FromMap(itemMap, "failedChecks")),
					NotYetEvaluatedChecks: types.Int64Value(getInt64FromMap(itemMap, "notYetEvaluatedChecks")),
					NumberOfChecks:        types.Int64Value(getInt64FromMap(itemMap, "numberOfChecks")),
					SuccessfulChecks:      types.Int64Value(getInt64FromMap(itemMap, "successfulChecks")),
				}
				catalogList = append(catalogList, catalogSummary)
			}
		}
		catalogListValue, diags := types.ListValueFrom(ctx, datasource_data_quality_summaries.CatalogSummariesType{}.ValueType(ctx).Type(ctx), catalogList)
		if diags.HasError() {
			tflog.Error(ctx, "Error creating catalog summaries list", map[string]interface{}{"errors": diags})
			model.CatalogSummaries = types.ListNull(datasource_data_quality_summaries.CatalogSummariesType{}.ValueType(ctx).Type(ctx))
		} else {
			model.CatalogSummaries = catalogListValue
		}
	} else {
		model.CatalogSummaries = types.ListNull(datasource_data_quality_summaries.CatalogSummariesType{})
	}

	// Map category_counts
	if categoryCounts, ok := response["categoryCounts"].([]interface{}); ok {
		categoryList := make([]datasource_data_quality_summaries.CategoryCountsValue, 0, len(categoryCounts))
		for _, item := range categoryCounts {
			if itemMap, ok := item.(map[string]interface{}); ok {
				categoryCount := datasource_data_quality_summaries.CategoryCountsValue{
					Category:              types.StringValue(getStringFromMap(itemMap, "category")),
					FailedEvaluations:     types.Int64Value(getInt64FromMap(itemMap, "failedEvaluations")),
					SuccessfulEvaluations: types.Int64Value(getInt64FromMap(itemMap, "successfulEvaluations")),
					TotalEvaluations:      types.Int64Value(getInt64FromMap(itemMap, "totalEvaluations")),
				}
				categoryList = append(categoryList, categoryCount)
			}
		}
		categoryListValue, diags := types.ListValueFrom(ctx, datasource_data_quality_summaries.CategoryCountsType{}.ValueType(ctx).Type(ctx), categoryList)
		if diags.HasError() {
			tflog.Error(ctx, "Error creating category counts list", map[string]interface{}{"errors": diags})
			model.CategoryCounts = types.ListNull(datasource_data_quality_summaries.CategoryCountsType{}.ValueType(ctx).Type(ctx))
		} else {
			model.CategoryCounts = categoryListValue
		}
	} else {
		model.CategoryCounts = types.ListNull(datasource_data_quality_summaries.CategoryCountsType{})
	}

	// Map daily_summaries
	if dailySummaries, ok := response["dailySummaries"].([]interface{}); ok {
		dailyList := make([]datasource_data_quality_summaries.DailySummariesValue, 0, len(dailySummaries))
		for _, item := range dailySummaries {
			if itemMap, ok := item.(map[string]interface{}); ok {
				dailySummary := datasource_data_quality_summaries.DailySummariesValue{
					Day:                   types.StringValue(getStringFromMap(itemMap, "day")),
					FailedEvaluations:     types.Int64Value(getInt64FromMap(itemMap, "failedEvaluations")),
					SuccessfulEvaluations: types.Int64Value(getInt64FromMap(itemMap, "successfulEvaluations")),
					TotalEvaluations:      types.Int64Value(getInt64FromMap(itemMap, "totalEvaluations")),
				}
				dailyList = append(dailyList, dailySummary)
			}
		}
		dailyListValue, diags := types.ListValueFrom(ctx, datasource_data_quality_summaries.DailySummariesType{}.ValueType(ctx).Type(ctx), dailyList)
		if diags.HasError() {
			tflog.Error(ctx, "Error creating daily summaries list", map[string]interface{}{"errors": diags})
			model.DailySummaries = types.ListNull(datasource_data_quality_summaries.DailySummariesType{}.ValueType(ctx).Type(ctx))
		} else {
			model.DailySummaries = dailyListValue
		}
	} else {
		model.DailySummaries = types.ListNull(datasource_data_quality_summaries.DailySummariesType{})
	}

	// Map severity_counts
	if severityCounts, ok := response["severityCounts"].([]interface{}); ok {
		severityList := make([]datasource_data_quality_summaries.SeverityCountsValue, 0, len(severityCounts))
		for _, item := range severityCounts {
			if itemMap, ok := item.(map[string]interface{}); ok {
				severityCount := datasource_data_quality_summaries.SeverityCountsValue{
					Severity:              types.StringValue(getStringFromMap(itemMap, "severity")),
					FailedEvaluations:     types.Int64Value(getInt64FromMap(itemMap, "failedEvaluations")),
					SuccessfulEvaluations: types.Int64Value(getInt64FromMap(itemMap, "successfulEvaluations")),
					TotalEvaluations:      types.Int64Value(getInt64FromMap(itemMap, "totalEvaluations")),
				}
				severityList = append(severityList, severityCount)
			}
		}
		severityListValue, diags := types.ListValueFrom(ctx, datasource_data_quality_summaries.SeverityCountsType{}.ValueType(ctx).Type(ctx), severityList)
		if diags.HasError() {
			tflog.Error(ctx, "Error creating severity counts list", map[string]interface{}{"errors": diags})
			model.SeverityCounts = types.ListNull(datasource_data_quality_summaries.SeverityCountsType{}.ValueType(ctx).Type(ctx))
		} else {
			model.SeverityCounts = severityListValue
		}
	} else {
		model.SeverityCounts = types.ListNull(datasource_data_quality_summaries.SeverityCountsType{})
	}
}
