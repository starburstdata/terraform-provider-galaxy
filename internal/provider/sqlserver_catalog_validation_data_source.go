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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_sqlserver_catalog_validation"
)

var _ datasource.DataSource = (*sqlserver_catalog_validationDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*sqlserver_catalog_validationDataSource)(nil)

func NewSqlserverCatalogValidationDataSource() datasource.DataSource {
	return &sqlserver_catalog_validationDataSource{}
}

type sqlserver_catalog_validationDataSource struct {
	client *client.GalaxyClient
}

func (d *sqlserver_catalog_validationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sqlserver_catalog_validation"
}

func (d *sqlserver_catalog_validationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_sqlserver_catalog_validation.SqlserverCatalogValidationDataSourceSchema(ctx)
}

func (d *sqlserver_catalog_validationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sqlserver_catalog_validationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_sqlserver_catalog_validation.SqlserverCatalogValidationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading sqlserver_catalog_validation", map[string]interface{}{"id": id})

	response, err := d.client.GetCatalog(ctx, "sqlserver", id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading sqlserver_catalog_validation",
			"Could not read sqlserver_catalog_validation "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *sqlserver_catalog_validationDataSource) updateModelFromResponse(ctx context.Context, model *datasource_sqlserver_catalog_validation.SqlserverCatalogValidationModel, response map[string]interface{}) {
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	}

	// Map other fields based on actual response structure
}
