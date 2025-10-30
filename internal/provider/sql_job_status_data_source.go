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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_sql_job_status"
)

var _ datasource.DataSource = (*sqlJobStatusDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*sqlJobStatusDataSource)(nil)

func NewSqlJobStatusDataSource() datasource.DataSource {
	return &sqlJobStatusDataSource{}
}

type sqlJobStatusDataSource struct {
	client *client.GalaxyClient
}

func (d *sqlJobStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_job_status"
}

func (d *sqlJobStatusDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_sql_job_status.SqlJobStatusDataSourceSchema(ctx)
}

func (d *sqlJobStatusDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sqlJobStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_sql_job_status.SqlJobStatusModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the SQL job ID from config
	id := config.SqlJobId.ValueString()

	tflog.Debug(ctx, "Reading sql_job_status data source", map[string]interface{}{"id": id})
	response, err := d.client.GetSqlJobStatus(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading sql job status",
			"Could not read sql job status for "+id+": "+err.Error(),
		)
		return
	}

	d.updateModelFromResponse(ctx, &config, response)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *sqlJobStatusDataSource) updateModelFromResponse(ctx context.Context, model *datasource_sql_job_status.SqlJobStatusModel, response map[string]interface{}) {
	if sqlJobId, ok := response["sqlJobId"].(string); ok {
		model.SqlJobId = types.StringValue(sqlJobId)
	}
	if status, ok := response["status"].(string); ok {
		model.Status = types.StringValue(status)
	}
	if queryId, ok := response["queryId"].(string); ok {
		model.QueryId = types.StringValue(queryId)
	}
	if errorMessage, ok := response["errorMessage"].(string); ok {
		model.ErrorMessage = types.StringValue(errorMessage)
	}
	if updatedAt, ok := response["updatedAt"].(string); ok {
		model.UpdatedAt = types.StringValue(updatedAt)
	}
	if progressPercentage, ok := response["progressPercentage"].(float64); ok {
		model.ProgressPercentage = types.Float64Value(progressPercentage)
	}
}
