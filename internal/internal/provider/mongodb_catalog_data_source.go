package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_mongodb_catalog"
)

var _ datasource.DataSource = (*mongodb_catalogDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*mongodb_catalogDataSource)(nil)

func NewMongodbCatalogDataSource() datasource.DataSource {
	return &mongodb_catalogDataSource{}
}

type mongodb_catalogDataSource struct {
	client *client.GalaxyClient
}

func (d *mongodb_catalogDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mongodb_catalog"
}

func (d *mongodb_catalogDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_mongodb_catalog.MongodbCatalogDataSourceSchema(ctx)
}

func (d *mongodb_catalogDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *mongodb_catalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_mongodb_catalog.MongodbCatalogModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading mongodb_catalog", map[string]interface{}{"id": id})

	response, err := d.client.GetCatalog(ctx, "mongodb", id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading mongodb_catalog",
			"Could not read mongodb_catalog "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *mongodb_catalogDataSource) updateModelFromResponse(ctx context.Context, model *datasource_mongodb_catalog.MongodbCatalogModel, response map[string]interface{}) {
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	}

	// Map other fields based on actual response structure
}
