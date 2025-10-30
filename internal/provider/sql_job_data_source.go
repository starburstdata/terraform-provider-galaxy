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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_sql_job"
)

var _ datasource.DataSource = (*sqlJobDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*sqlJobDataSource)(nil)

func NewSqlJobDataSource() datasource.DataSource {
	return &sqlJobDataSource{}
}

type sqlJobDataSource struct {
	client *client.GalaxyClient
}

func (d *sqlJobDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sql_job"
}

func (d *sqlJobDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_sql_job.SqlJobDataSourceSchema(ctx)
}

func (d *sqlJobDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sqlJobDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_sql_job.SqlJobModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the SQL job ID from config
	id := config.SqlJobId.ValueString()

	tflog.Debug(ctx, "Reading sql_job data source", map[string]interface{}{"id": id})
	response, err := d.client.GetSqlJob(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading sql_job",
			"Could not read sql_job "+id+": "+err.Error(),
		)
		return
	}

	d.updateModelFromResponse(ctx, &config, response)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *sqlJobDataSource) updateModelFromResponse(ctx context.Context, model *datasource_sql_job.SqlJobModel, response map[string]interface{}) {
	if sqlJobId, ok := response["sqlJobId"].(string); ok {
		model.SqlJobId = types.StringValue(sqlJobId)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	}
	if clusterId, ok := response["clusterId"].(string); ok {
		model.ClusterId = types.StringValue(clusterId)
	}
	if roleId, ok := response["roleId"].(string); ok {
		model.RoleId = types.StringValue(roleId)
	}
	if query, ok := response["query"].(string); ok {
		model.Query = types.StringValue(query)
	}
	if cronExpression, ok := response["cronExpression"].(string); ok {
		model.CronExpression = types.StringValue(cronExpression)
	}
	if timezone, ok := response["timezone"].(string); ok {
		model.Timezone = types.StringValue(timezone)
	}
	if nextExecution, ok := response["nextExecution"].(string); ok {
		model.NextExecution = types.StringValue(nextExecution)
	}
}
