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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_privatelink"
)

var _ datasource.DataSource = (*privatelinkDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*privatelinkDataSource)(nil)

func NewPrivatelinkDataSource() datasource.DataSource {
	return &privatelinkDataSource{}
}

type privatelinkDataSource struct {
	client *client.GalaxyClient
}

func (d *privatelinkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_privatelink"
}

func (d *privatelinkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_privatelink.PrivatelinkDataSourceSchema(ctx)
}

func (d *privatelinkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *privatelinkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_privatelink.PrivatelinkModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the privatelink ID from config
	id := config.PrivatelinkId.ValueString()

	tflog.Debug(ctx, "Reading privatelink data source", map[string]interface{}{"id": id})
	response, err := d.client.GetPrivatelink(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading privatelink",
			"Could not read privatelink "+id+": "+err.Error(),
		)
		return
	}

	d.updateModelFromResponse(ctx, &config, response)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *privatelinkDataSource) updateModelFromResponse(ctx context.Context, model *datasource_privatelink.PrivatelinkModel, response map[string]interface{}) {
	if privatelinkId, ok := response["privatelinkId"].(string); ok {
		model.PrivatelinkId = types.StringValue(privatelinkId)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if cloudRegionId, ok := response["cloudRegionId"].(string); ok {
		model.CloudRegionId = types.StringValue(cloudRegionId)
	}
}
