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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_gcs_catalog"
)

var _ datasource.DataSource = (*gcs_catalogDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*gcs_catalogDataSource)(nil)

func NewGcsCatalogDataSource() datasource.DataSource {
	return &gcs_catalogDataSource{}
}

type gcs_catalogDataSource struct {
	client *client.GalaxyClient
}

func (d *gcs_catalogDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcs_catalog"
}

func (d *gcs_catalogDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_gcs_catalog.GcsCatalogDataSourceSchema(ctx)
}

func (d *gcs_catalogDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *gcs_catalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_gcs_catalog.GcsCatalogModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.CatalogId.ValueString()
	tflog.Debug(ctx, "Reading gcs_catalog", map[string]interface{}{"id": id})

	response, err := d.client.GetCatalog(ctx, "gcs", id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading gcs_catalog",
			"Could not read gcs_catalog "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *gcs_catalogDataSource) updateModelFromResponse(ctx context.Context, model *datasource_gcs_catalog.GcsCatalogModel, response map[string]interface{}) {
	// Map response fields to model
	if catalogId, ok := response["catalogId"].(string); ok {
		model.CatalogId = types.StringValue(catalogId)
	}

	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}

	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else if model.Description.IsUnknown() {
		model.Description = types.StringNull()
	}

	if readOnly, ok := response["readOnly"].(bool); ok {
		model.ReadOnly = types.BoolValue(readOnly)
	}

	if metastoreType, ok := response["metastoreType"].(string); ok {
		model.MetastoreType = types.StringValue(metastoreType)
	} else if model.MetastoreType.IsUnknown() {
		model.MetastoreType = types.StringNull()
	}

	// CredentialsKey is write-only - the API returns "<Value is encrypted>"
	// We don't update the credentials_key field from the API response since it's not the actual value.

	if defaultBucket, ok := response["defaultBucket"].(string); ok {
		model.DefaultBucket = types.StringValue(defaultBucket)
	} else if model.DefaultBucket.IsUnknown() {
		model.DefaultBucket = types.StringNull()
	}

	if defaultDataLocation, ok := response["defaultDataLocation"].(string); ok {
		model.DefaultDataLocation = types.StringValue(defaultDataLocation)
	} else if model.DefaultDataLocation.IsUnknown() {
		model.DefaultDataLocation = types.StringNull()
	}

	if defaultTableFormat, ok := response["defaultTableFormat"].(string); ok {
		model.DefaultTableFormat = types.StringValue(defaultTableFormat)
	} else if model.DefaultTableFormat.IsUnknown() {
		model.DefaultTableFormat = types.StringNull()
	}

	if externalTableCreationEnabled, ok := response["externalTableCreationEnabled"].(bool); ok {
		model.ExternalTableCreationEnabled = types.BoolValue(externalTableCreationEnabled)
	}

	if externalTableWritesEnabled, ok := response["externalTableWritesEnabled"].(bool); ok {
		model.ExternalTableWritesEnabled = types.BoolValue(externalTableWritesEnabled)
	}

	if hiveMetastoreHost, ok := response["hiveMetastoreHost"].(string); ok {
		model.HiveMetastoreHost = types.StringValue(hiveMetastoreHost)
	} else if model.HiveMetastoreHost.IsUnknown() {
		model.HiveMetastoreHost = types.StringNull()
	}

	if hiveMetastorePort, ok := response["hiveMetastorePort"].(float64); ok {
		model.HiveMetastorePort = types.Int64Value(int64(hiveMetastorePort))
	}

	if sshTunnelId, ok := response["sshTunnelId"].(string); ok {
		model.SshTunnelId = types.StringValue(sshTunnelId)
	} else if model.SshTunnelId.IsUnknown() {
		model.SshTunnelId = types.StringNull()
	}

	if validate, ok := response["validate"].(bool); ok {
		model.Validate = types.BoolValue(validate)
	} else if model.Validate.IsUnknown() {
		model.Validate = types.BoolNull()
	}
}
