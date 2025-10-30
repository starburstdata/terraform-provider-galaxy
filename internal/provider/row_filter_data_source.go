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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_row_filter"
)

var _ datasource.DataSource = (*row_filterDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*row_filterDataSource)(nil)

func NewRowFilterDataSource() datasource.DataSource {
	return &row_filterDataSource{}
}

type row_filterDataSource struct {
	client *client.GalaxyClient
}

func (d *row_filterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_row_filter"
}

func (d *row_filterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_row_filter.RowFilterDataSourceSchema(ctx)
}

func (d *row_filterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *row_filterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_row_filter.RowFilterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.RowFilterId.ValueString()
	tflog.Debug(ctx, "Reading row_filter", map[string]interface{}{"id": id})

	response, err := d.client.GetRowFilter(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading row_filter",
			"Could not read row_filter "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *row_filterDataSource) updateModelFromResponse(ctx context.Context, model *datasource_row_filter.RowFilterModel, response map[string]interface{}) {
	// Map response fields to model based on API response structure
	if rowFilterId, ok := response["rowFilterId"].(string); ok {
		model.RowFilterId = types.StringValue(rowFilterId)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if expression, ok := response["expression"].(string); ok {
		model.Expression = types.StringValue(expression)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	}
	if created, ok := response["created"].(string); ok {
		model.Created = types.StringValue(created)
	}
	if modified, ok := response["modified"].(string); ok {
		model.Modified = types.StringValue(modified)
	}
}
