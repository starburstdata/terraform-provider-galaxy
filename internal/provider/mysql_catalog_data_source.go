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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_mysql_catalog"
)

var _ datasource.DataSource = (*mysql_catalogDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*mysql_catalogDataSource)(nil)

func NewMysqlCatalogDataSource() datasource.DataSource {
	return &mysql_catalogDataSource{}
}

type mysql_catalogDataSource struct {
	client *client.GalaxyClient
}

func (d *mysql_catalogDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_catalog"
}

func (d *mysql_catalogDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_mysql_catalog.MysqlCatalogDataSourceSchema(ctx)
}

func (d *mysql_catalogDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *mysql_catalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_mysql_catalog.MysqlCatalogModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.CatalogId.ValueString()
	tflog.Debug(ctx, "Reading mysql_catalog", map[string]interface{}{"id": id})

	response, err := d.client.GetCatalog(ctx, "mysql", id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading mysql_catalog",
			"Could not read mysql_catalog "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *mysql_catalogDataSource) updateModelFromResponse(ctx context.Context, model *datasource_mysql_catalog.MysqlCatalogModel, response map[string]interface{}) {
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

	if host, ok := response["host"].(string); ok {
		model.Host = types.StringValue(host)
	}

	if port, ok := response["port"].(float64); ok {
		model.Port = types.Int64Value(int64(port))
	}

	if username, ok := response["username"].(string); ok {
		model.Username = types.StringValue(username)
	} else if model.Username.IsUnknown() {
		model.Username = types.StringNull()
	}

	// Password is write-only - the API returns "<Value is encrypted>"
	// We don't update the password field from the API response since it's not the actual value.

	if connectionType, ok := response["connectionType"].(string); ok {
		model.ConnectionType = types.StringValue(connectionType)
	} else if model.ConnectionType.IsUnknown() {
		model.ConnectionType = types.StringNull()
	}

	if cloudKind, ok := response["cloudKind"].(string); ok {
		model.CloudKind = types.StringValue(cloudKind)
	} else if model.CloudKind.IsUnknown() {
		model.CloudKind = types.StringNull()
	}

	if privateLinkId, ok := response["privateLinkId"].(string); ok {
		model.PrivateLinkId = types.StringValue(privateLinkId)
	} else if model.PrivateLinkId.IsUnknown() {
		model.PrivateLinkId = types.StringNull()
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
