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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_file_ingest_sources"
)

var _ datasource.DataSource = (*fileIngestSourcesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*fileIngestSourcesDataSource)(nil)

func NewFileIngestSourcesDataSource() datasource.DataSource {
	return &fileIngestSourcesDataSource{}
}

type fileIngestSourcesDataSource struct {
	client *client.GalaxyClient
}

func (d *fileIngestSourcesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_ingest_sources"
}

func (d *fileIngestSourcesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_file_ingest_sources.FileIngestSourcesDataSourceSchema(ctx)
}

func (d *fileIngestSourcesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *fileIngestSourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_file_ingest_sources.FileIngestSourcesModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	results, err := d.client.ListFileIngestSources(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading file_ingest_sources",
			"Could not read file_ingest_sources: "+err.Error(),
		)
		return
	}

	elementType := datasource_file_ingest_sources.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_file_ingest_sources.ResultValue{}.AttributeTypes(ctx),
		},
	}
	items := make([]datasource_file_ingest_sources.ResultValue, 0, len(results))
	for _, raw := range results {
		m, ok := raw.(map[string]interface{})
		if !ok {
			tflog.Warn(ctx, fmt.Sprintf("unexpected entry type, skipping: %T", raw))
			continue
		}
		attrs := map[string]attr.Value{
			"bucket":                types.StringNull(),
			"description":           types.StringNull(),
			"file_ingest_source_id": types.StringNull(),
			"name":                  types.StringNull(),
			"prefix":                types.StringNull(),
		}
		if v, ok := m["bucket"].(string); ok {
			attrs["bucket"] = types.StringValue(v)
		}
		if v, ok := m["description"].(string); ok {
			attrs["description"] = types.StringValue(v)
		}
		if v, ok := m["fileIngestSourceId"].(string); ok {
			attrs["file_ingest_source_id"] = types.StringValue(v)
		}
		if v, ok := m["name"].(string); ok {
			attrs["name"] = types.StringValue(v)
		}
		if v, ok := m["prefix"].(string); ok {
			attrs["prefix"] = types.StringValue(v)
		}
		v, dd := datasource_file_ingest_sources.NewResultValue(datasource_file_ingest_sources.ResultValue{}.AttributeTypes(ctx), attrs)
		resp.Diagnostics.Append(dd...)
		items = append(items, v)
	}

	listValue, dd := types.ListValueFrom(ctx, elementType, items)
	resp.Diagnostics.Append(dd...)
	config.Result = listValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
