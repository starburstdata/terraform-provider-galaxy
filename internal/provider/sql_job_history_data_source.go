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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_sql_job_history"
)

var _ datasource.DataSource = (*sqlJobHistoryDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*sqlJobHistoryDataSource)(nil)

func NewSqlJobHistoryDataSource() datasource.DataSource {
	return &sqlJobHistoryDataSource{}
}

type sqlJobHistoryDataSource struct {
	client *client.GalaxyClient
}

func (d *sqlJobHistoryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_job_history"
}

func (d *sqlJobHistoryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_sql_job_history.SqlJobHistoryDataSourceSchema(ctx)
}

func (d *sqlJobHistoryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.GalaxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.GalaxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *sqlJobHistoryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_sql_job_history.SqlJobHistoryModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the SQL job ID from config
	id := config.Id.ValueString()

	tflog.Debug(ctx, "Reading sql_job_history data source", map[string]interface{}{"id": id})
	response, err := d.client.GetSqlJobHistory(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading sql job history",
			"Could not read sql job history for "+id+": "+err.Error(),
		)
		return
	}

	d.updateModelFromResponse(ctx, &config, response)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *sqlJobHistoryDataSource) updateModelFromResponse(ctx context.Context, model *datasource_sql_job_history.SqlJobHistoryModel, response map[string]interface{}) {
	// Set the ID field
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	}

	// Extract the result array from the API response
	resultInterface, ok := response["result"]
	if !ok {
		return
	}

	resultArray, ok := resultInterface.([]interface{})
	if !ok {
		return
	}

	// Convert to list of Terraform objects
	var resultList []attr.Value
	for _, item := range resultArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Create a ResultValue for each history entry
		resultValue := datasource_sql_job_history.ResultValue{
			ErrorMessage:       types.StringValue(""),
			ProgressPercentage: types.Float64Value(0),
			Query:              types.StringValue(""),
			QueryId:            types.StringValue(""),
			StartedAt:          types.StringValue(""),
			Status:             types.StringValue(""),
			UpdatedAt:          types.StringValue(""),
		}

		if errorMessage, ok := itemMap["errorMessage"].(string); ok {
			resultValue.ErrorMessage = types.StringValue(errorMessage)
		}
		if progressPercentage, ok := itemMap["progressPercentage"].(float64); ok {
			resultValue.ProgressPercentage = types.Float64Value(progressPercentage)
		}
		if query, ok := itemMap["query"].(string); ok {
			resultValue.Query = types.StringValue(query)
		}
		if queryId, ok := itemMap["queryId"].(string); ok {
			resultValue.QueryId = types.StringValue(queryId)
		}
		if startedAt, ok := itemMap["startedAt"].(string); ok {
			resultValue.StartedAt = types.StringValue(startedAt)
		}
		if status, ok := itemMap["status"].(string); ok {
			resultValue.Status = types.StringValue(status)
		}
		if updatedAt, ok := itemMap["updatedAt"].(string); ok {
			resultValue.UpdatedAt = types.StringValue(updatedAt)
		}

		objectValue, diags := resultValue.ToObjectValue(ctx)
		if diags.HasError() {
			tflog.Error(ctx, "Error converting sql job history result to object value", map[string]interface{}{"errors": diags})
			continue
		}

		resultList = append(resultList, objectValue)
	}

	// Convert to Terraform list
	resultListValue, diags := types.ListValue(datasource_sql_job_history.ResultType{}.ValueType(ctx).Type(ctx), resultList)
	if diags.HasError() {
		tflog.Error(ctx, "Error creating sql job history result list", map[string]interface{}{"errors": diags})
		model.Result = types.ListNull(datasource_sql_job_history.ResultType{}.ValueType(ctx).Type(ctx))
	} else {
		model.Result = resultListValue
	}
}
