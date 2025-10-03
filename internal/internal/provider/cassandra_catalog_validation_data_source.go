package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_cassandra_catalog_validation"
)

var _ datasource.DataSource = (*cassandra_catalog_validationDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*cassandra_catalog_validationDataSource)(nil)

func NewCassandraCatalogValidationDataSource() datasource.DataSource {
	return &cassandra_catalog_validationDataSource{}
}

type cassandra_catalog_validationDataSource struct {
	client *client.GalaxyClient
}

func (d *cassandra_catalog_validationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cassandra_catalog_validation"
}

func (d *cassandra_catalog_validationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_cassandra_catalog_validation.CassandraCatalogValidationDataSourceSchema(ctx)
}

func (d *cassandra_catalog_validationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cassandra_catalog_validationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_cassandra_catalog_validation.CassandraCatalogValidationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading cassandra_catalog_validation", map[string]interface{}{"id": id})

	response, err := d.client.GetCatalog(ctx, "cassandra", id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading cassandra_catalog_validation",
			"Could not read cassandra_catalog_validation "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *cassandra_catalog_validationDataSource) updateModelFromResponse(ctx context.Context, model *datasource_cassandra_catalog_validation.CassandraCatalogValidationModel, response map[string]interface{}) {
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	}

	// Map other fields based on actual response structure
}
