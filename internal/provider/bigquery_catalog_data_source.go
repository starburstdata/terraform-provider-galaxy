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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_bigquery_catalog"
)

var _ datasource.DataSource = (*bigquery_catalogDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*bigquery_catalogDataSource)(nil)

func NewBigqueryCatalogDataSource() datasource.DataSource {
	return &bigquery_catalogDataSource{}
}

type bigquery_catalogDataSource struct {
	client *client.GalaxyClient
}

func (d *bigquery_catalogDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bigquery_catalog"
}

func (d *bigquery_catalogDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_bigquery_catalog.BigqueryCatalogDataSourceSchema(ctx)
}

func (d *bigquery_catalogDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bigquery_catalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_bigquery_catalog.BigqueryCatalogModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.CatalogId.ValueString()
	tflog.Debug(ctx, "Reading bigquery_catalog", map[string]interface{}{"id": id})

	response, err := d.client.GetCatalog(ctx, "bigquery", id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bigquery_catalog",
			"Could not read bigquery_catalog "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *bigquery_catalogDataSource) updateModelFromResponse(ctx context.Context, model *datasource_bigquery_catalog.BigqueryCatalogModel, response map[string]interface{}) {
	// Map response fields to model
	if catalogId, ok := response["catalogId"].(string); ok {
		model.CatalogId = types.StringValue(catalogId)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	}
	if readOnly, ok := response["readOnly"].(bool); ok {
		model.ReadOnly = types.BoolValue(readOnly)
	}

	// BigQuery-specific fields
	if credentialsKey, ok := response["credentialsKey"].(string); ok {
		model.CredentialsKey = types.StringValue(credentialsKey)
	}
}
