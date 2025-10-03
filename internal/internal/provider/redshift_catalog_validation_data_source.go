package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_redshift_catalog_validation"
)

var _ datasource.DataSource = (*redshift_catalog_validationDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*redshift_catalog_validationDataSource)(nil)

func NewRedshiftCatalogValidationDataSource() datasource.DataSource {
	return &redshift_catalog_validationDataSource{}
}

type redshift_catalog_validationDataSource struct {
	client *client.GalaxyClient
}

func (d *redshift_catalog_validationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redshift_catalog_validation"
}

func (d *redshift_catalog_validationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_redshift_catalog_validation.RedshiftCatalogValidationDataSourceSchema(ctx)
}

func (d *redshift_catalog_validationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *redshift_catalog_validationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_redshift_catalog_validation.RedshiftCatalogValidationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading redshift_catalog_validation", map[string]interface{}{"id": id})

	response, err := d.client.GetCatalog(ctx, "redshift", id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading redshift_catalog_validation",
			"Could not read redshift_catalog_validation "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *redshift_catalog_validationDataSource) updateModelFromResponse(ctx context.Context, model *datasource_redshift_catalog_validation.RedshiftCatalogValidationModel, response map[string]interface{}) {
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	}

	// Map other fields based on actual response structure
}
