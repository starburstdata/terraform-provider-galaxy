package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_snowflake_catalog"
)

var _ datasource.DataSource = (*snowflake_catalogDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*snowflake_catalogDataSource)(nil)

func NewSnowflakeCatalogDataSource() datasource.DataSource {
	return &snowflake_catalogDataSource{}
}

type snowflake_catalogDataSource struct {
	client *client.GalaxyClient
}

func (d *snowflake_catalogDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snowflake_catalog"
}

func (d *snowflake_catalogDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_snowflake_catalog.SnowflakeCatalogDataSourceSchema(ctx)
}

func (d *snowflake_catalogDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *snowflake_catalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_snowflake_catalog.SnowflakeCatalogModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading snowflake_catalog", map[string]interface{}{"id": id})

	response, err := d.client.GetCatalog(ctx, "snowflake", id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading snowflake_catalog",
			"Could not read snowflake_catalog "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *snowflake_catalogDataSource) updateModelFromResponse(ctx context.Context, model *datasource_snowflake_catalog.SnowflakeCatalogModel, response map[string]interface{}) {
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	}

	// Map other fields based on actual response structure
}
