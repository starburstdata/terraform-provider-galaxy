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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_data_domains"
)

var _ datasource.DataSource = (*dataDomainsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*dataDomainsDataSource)(nil)

func NewDataDomainsDataSource() datasource.DataSource {
	return &dataDomainsDataSource{}
}

type dataDomainsDataSource struct {
	client *client.GalaxyClient
}

func (d *dataDomainsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_domains"
}

func (d *dataDomainsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_data_domains.DataDomainsDataSourceSchema(ctx)
}

func (d *dataDomainsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.GalaxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.GalaxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *dataDomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_data_domains.DataDomainsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataProductID := config.DataProductId.ValueString()
	tflog.Debug(ctx, "Reading data domains for data product", map[string]interface{}{"data_product_id": dataProductID})

	results, err := d.client.ListDataDomainsByDataProduct(ctx, dataProductID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading data domains",
			"Could not read data domains for data product "+dataProductID+": "+err.Error(),
		)
		return
	}

	config.DataDomainId = types.StringNull()
	config.Name = types.StringNull()
	for _, item := range results {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if id, ok := m["dataDomainId"].(string); ok {
			config.DataDomainId = types.StringValue(id)
		}
		if name, ok := m["name"].(string); ok {
			config.Name = types.StringValue(name)
		}
		break
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
