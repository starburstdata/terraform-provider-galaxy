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
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_live_tables"
)

var _ datasource.DataSource = (*liveTablesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*liveTablesDataSource)(nil)

func NewLiveTablesDataSource() datasource.DataSource {
	return &liveTablesDataSource{}
}

type liveTablesDataSource struct {
	client *client.GalaxyClient
}

func (d *liveTablesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_live_tables"
}

func (d *liveTablesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_live_tables.LiveTablesDataSourceSchema(ctx)
}

func (d *liveTablesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *liveTablesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	results, err := d.client.ListLiveTables(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading live_tables",
			"Could not read live_tables: "+err.Error(),
		)
		return
	}

	s := datasource_live_tables.LiveTablesDataSourceSchema(ctx)
	resultSchemaAttr := s.Attributes["result"]
	resultFwType := resultSchemaAttr.GetType()
	resultTfType := resultFwType.TerraformType(ctx)

	tfval, convDiags := mapToTftypesValue(ctx, results, resultTfType)
	resp.Diagnostics.Append(convDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	attrVal, err := resultFwType.ValueFromTerraform(ctx, tfval)
	if err != nil {
		resp.Diagnostics.AddError("conversion error", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("result"), attrVal)...)
}
