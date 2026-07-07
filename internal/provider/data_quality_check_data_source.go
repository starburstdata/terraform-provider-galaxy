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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_data_quality_check"
)

var _ datasource.DataSource = (*dataQualityCheckDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*dataQualityCheckDataSource)(nil)

func NewDataQualityCheckDataSource() datasource.DataSource {
	return &dataQualityCheckDataSource{}
}

type dataQualityCheckDataSource struct {
	client *client.GalaxyClient
}

func (d *dataQualityCheckDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_quality_check"
}

func (d *dataQualityCheckDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_data_quality_check.DataQualityCheckDataSourceSchema(ctx)
}

func (d *dataQualityCheckDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *dataQualityCheckDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_data_quality_check.DataQualityCheckModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.DataQualityCheckId.ValueString()
	tflog.Debug(ctx, "Reading data quality check", map[string]interface{}{"id": id})

	response, err := d.client.GetDataQualityCheck(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading data quality check",
			"Could not read data quality check "+id+": "+err.Error(),
		)
		return
	}

	d.updateModelFromResponse(&config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *dataQualityCheckDataSource) updateModelFromResponse(model *datasource_data_quality_check.DataQualityCheckModel, response map[string]interface{}) {
	if id, ok := response["dataQualityCheckId"].(string); ok {
		model.DataQualityCheckId = types.StringValue(id)
	}
	if catalogId, ok := response["catalogId"].(string); ok {
		model.CatalogId = types.StringValue(catalogId)
	}
	if category, ok := response["category"].(string); ok {
		model.Category = types.StringValue(category)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	}
	if kind, ok := response["kind"].(string); ok {
		model.Kind = types.StringValue(kind)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if query, ok := response["query"].(string); ok {
		model.Query = types.StringValue(query)
	}
	if schemaId, ok := response["schemaId"].(string); ok {
		model.SchemaId = types.StringValue(schemaId)
	}
	if severity, ok := response["severity"].(string); ok {
		model.Severity = types.StringValue(severity)
	}
	if tableId, ok := response["tableId"].(string); ok {
		model.TableId = types.StringValue(tableId)
	}
}
