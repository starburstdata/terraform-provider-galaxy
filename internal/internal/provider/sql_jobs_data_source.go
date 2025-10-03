package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_sql_jobs"
)

var _ datasource.DataSource = (*sqlJobsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*sqlJobsDataSource)(nil)

func NewSqlJobsDataSource() datasource.DataSource {
	return &sqlJobsDataSource{}
}

type sqlJobsDataSource struct {
	client *client.GalaxyClient
}

func (d *sqlJobsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_jobs"
}

func (d *sqlJobsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_sql_jobs.SqlJobsDataSourceSchema(ctx)
}

func (d *sqlJobsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sqlJobsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_sql_jobs.SqlJobsModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading sql_jobs data source")
	response, err := d.client.ListSqlJobs(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading sql jobs",
			"Could not read sql jobs: "+err.Error(),
		)
		return
	}

	d.updateModelFromResponse(ctx, &config, response)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *sqlJobsDataSource) updateModelFromResponse(ctx context.Context, model *datasource_sql_jobs.SqlJobsModel, response map[string]interface{}) {
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

		// Create a ResultValue for each SQL job
		resultValue := datasource_sql_jobs.ResultValue{
			CronExpression: types.StringValue(""),
			Description:    types.StringValue(""),
			Name:           types.StringValue(""),
			RoleId:         types.StringValue(""),
			SqlJobId:       types.StringValue(""),
			Timezone:       types.StringValue(""),
		}

		if cronExpression, ok := itemMap["cronExpression"].(string); ok {
			resultValue.CronExpression = types.StringValue(cronExpression)
		}
		if description, ok := itemMap["description"].(string); ok {
			resultValue.Description = types.StringValue(description)
		}
		if name, ok := itemMap["name"].(string); ok {
			resultValue.Name = types.StringValue(name)
		}
		if roleId, ok := itemMap["roleId"].(string); ok {
			resultValue.RoleId = types.StringValue(roleId)
		}
		if sqlJobId, ok := itemMap["sqlJobId"].(string); ok {
			resultValue.SqlJobId = types.StringValue(sqlJobId)
		}
		if timezone, ok := itemMap["timezone"].(string); ok {
			resultValue.Timezone = types.StringValue(timezone)
		}

		objectValue, diags := resultValue.ToObjectValue(ctx)
		if diags.HasError() {
			tflog.Error(ctx, "Error converting sql job result to object value", map[string]interface{}{"errors": diags})
			continue
		}

		resultList = append(resultList, objectValue)
	}

	// Convert to Terraform list
	resultListValue, diags := types.ListValue(datasource_sql_jobs.ResultType{}.ValueType(ctx).Type(ctx), resultList)
	if diags.HasError() {
		tflog.Error(ctx, "Error creating sql jobs result list", map[string]interface{}{"errors": diags})
		model.Result = types.ListNull(datasource_sql_jobs.ResultType{}.ValueType(ctx).Type(ctx))
	} else {
		model.Result = resultListValue
	}
}
