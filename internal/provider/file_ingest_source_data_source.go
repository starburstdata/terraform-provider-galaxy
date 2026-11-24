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

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_file_ingest_source"
)

var _ datasource.DataSource = (*fileIngestSourceDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*fileIngestSourceDataSource)(nil)

func NewFileIngestSourceDataSource() datasource.DataSource {
	return &fileIngestSourceDataSource{}
}

type fileIngestSourceDataSource struct {
	client *client.GalaxyClient
}

func (d *fileIngestSourceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_ingest_source"
}

func (d *fileIngestSourceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_file_ingest_source.FileIngestSourceDataSourceSchema(ctx)
}

func (d *fileIngestSourceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *fileIngestSourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_file_ingest_source.FileIngestSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.FileIngestSourceId.ValueString()
	response, err := d.client.GetFileIngestSource(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading file_ingest_source",
			"Could not read file_ingest_source "+id+": "+err.Error(),
		)
		return
	}

	if name, ok := response["name"].(string); ok {
		config.Name = types.StringValue(name)
	} else {
		config.Name = types.StringNull()
	}
	if description, ok := response["description"].(string); ok {
		config.Description = types.StringValue(description)
	} else {
		config.Description = types.StringNull()
	}
	if bucket, ok := response["bucket"].(string); ok {
		config.Bucket = types.StringValue(bucket)
	} else {
		config.Bucket = types.StringNull()
	}
	if prefix, ok := response["prefix"].(string); ok {
		config.Prefix = types.StringValue(prefix)
	} else {
		config.Prefix = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
