package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_column_mask"
)

var _ datasource.DataSource = (*column_maskDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*column_maskDataSource)(nil)

func NewColumnMaskDataSource() datasource.DataSource {
	return &column_maskDataSource{}
}

type column_maskDataSource struct {
	client *client.GalaxyClient
}

func (d *column_maskDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_column_mask"
}

func (d *column_maskDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_column_mask.ColumnMaskDataSourceSchema(ctx)
}

func (d *column_maskDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *column_maskDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_column_mask.ColumnMaskModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading column_mask", map[string]interface{}{"id": id})

	response, err := d.client.GetColumnMask(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading column_mask",
			"Could not read column_mask "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *column_maskDataSource) updateModelFromResponse(ctx context.Context, model *datasource_column_mask.ColumnMaskModel, response map[string]interface{}) {
	// Map response fields to model based on API response structure
	if columnMaskId, ok := response["columnMaskId"].(string); ok {
		model.Id = types.StringValue(columnMaskId)
		model.ColumnMaskId = types.StringValue(columnMaskId)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if expression, ok := response["expression"].(string); ok {
		model.Expression = types.StringValue(expression)
	}
	if columnMaskType, ok := response["columnMaskType"].(string); ok {
		model.ColumnMaskType = types.StringValue(columnMaskType)
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
