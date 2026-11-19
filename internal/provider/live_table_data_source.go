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
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_live_table"
)

var _ datasource.DataSource = (*liveTableDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*liveTableDataSource)(nil)

func NewLiveTableDataSource() datasource.DataSource {
	return &liveTableDataSource{}
}

type liveTableDataSource struct {
	client *client.GalaxyClient
}

func (d *liveTableDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_live_table"
}

func (d *liveTableDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_live_table.LiveTableDataSourceSchema(ctx)
}

func (d *liveTableDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *liveTableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Pull the live_table_id from configuration
	var liveTableID types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("live_table_id"), &liveTableID)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := liveTableID.ValueString()
	response, err := d.client.GetLiveTable(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading live_table",
			"Could not read live_table "+id+": "+err.Error(),
		)
		return
	}

	s := datasource_live_table.LiveTableDataSourceSchema(ctx)
	for attrName, attrSchema := range s.Attributes {
		attrFwType := attrSchema.GetType()
		attrType := attrFwType.TerraformType(ctx)
		key := snakeToCamel(attrName)
		raw, has := response[key]
		if !has {
			raw = nil
		}
		tfval, convDiags := mapToTftypesValue(ctx, raw, attrType)
		resp.Diagnostics.Append(convDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		attrVal, err := attrFwType.ValueFromTerraform(ctx, tfval)
		if err != nil {
			resp.Diagnostics.AddError("conversion error", fmt.Sprintf("%s: %v", attrName, err))
			return
		}
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(attrName), attrVal)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}
