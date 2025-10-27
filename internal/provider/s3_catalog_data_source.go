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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_s3_catalog"
)

var _ datasource.DataSource = (*s3_catalogDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*s3_catalogDataSource)(nil)

func NewS3CatalogDataSource() datasource.DataSource {
	return &s3_catalogDataSource{}
}

type s3_catalogDataSource struct {
	client *client.GalaxyClient
}

func (d *s3_catalogDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_catalog"
}

func (d *s3_catalogDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_s3_catalog.S3CatalogDataSourceSchema(ctx)
}

func (d *s3_catalogDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *s3_catalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_s3_catalog.S3CatalogModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading s3_catalog", map[string]interface{}{"id": id})

	response, err := d.client.GetCatalog(ctx, "s3", id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading s3_catalog",
			"Could not read s3_catalog "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *s3_catalogDataSource) updateModelFromResponse(ctx context.Context, model *datasource_s3_catalog.S3CatalogModel, response map[string]interface{}) {
	// Map response fields to model - this is now a regular catalog data source with full catalog fields
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	}

	if catalogId, ok := response["catalogId"].(string); ok {
		model.CatalogId = types.StringValue(catalogId)
	}

	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}

	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	}

	if readOnly, ok := response["readOnly"].(bool); ok {
		model.ReadOnly = types.BoolValue(readOnly)
	}

	if accessKey, ok := response["accessKey"].(string); ok {
		model.AccessKey = types.StringValue(accessKey)
	}

	if secretKey, ok := response["secretKey"].(string); ok {
		model.SecretKey = types.StringValue(secretKey)
	}

	if roleArn, ok := response["roleArn"].(string); ok {
		model.RoleArn = types.StringValue(roleArn)
	}

	if metastoreType, ok := response["metastoreType"].(string); ok {
		model.MetastoreType = types.StringValue(metastoreType)
	}

	if defaultTableFormat, ok := response["defaultTableFormat"].(string); ok {
		model.DefaultTableFormat = types.StringValue(defaultTableFormat)
	}

	if externalTableCreationEnabled, ok := response["externalTableCreationEnabled"].(bool); ok {
		model.ExternalTableCreationEnabled = types.BoolValue(externalTableCreationEnabled)
	}

	if externalTableWritesEnabled, ok := response["externalTableWritesEnabled"].(bool); ok {
		model.ExternalTableWritesEnabled = types.BoolValue(externalTableWritesEnabled)
	}

}
