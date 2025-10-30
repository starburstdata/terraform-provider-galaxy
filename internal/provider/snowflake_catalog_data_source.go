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

	id := config.CatalogId.ValueString()
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

	if accountIdentifier, ok := response["accountIdentifier"].(string); ok {
		model.AccountIdentifier = types.StringValue(accountIdentifier)
	}

	if databaseName, ok := response["databaseName"].(string); ok {
		model.DatabaseName = types.StringValue(databaseName)
	}

	if warehouse, ok := response["warehouse"].(string); ok {
		model.Warehouse = types.StringValue(warehouse)
	}

	if authenticationType, ok := response["authenticationType"].(string); ok {
		model.AuthenticationType = types.StringValue(authenticationType)
	} else if model.AuthenticationType.IsUnknown() {
		model.AuthenticationType = types.StringNull()
	}

	if username, ok := response["username"].(string); ok {
		model.Username = types.StringValue(username)
	} else if model.Username.IsUnknown() {
		model.Username = types.StringNull()
	}

	// Password is write-only - the API returns "<Value is encrypted>"
	// We don't update the password field from the API response since it's not the actual value.

	// PrivateKey is write-only - the API returns "<Value is encrypted>"
	// We don't update the private_key field from the API response since it's not the actual value.

	// PrivateKeyPassphrase is write-only - the API returns "<Value is encrypted>"
	// We don't update the private_key_passphrase field from the API response since it's not the actual value.

	if role, ok := response["role"].(string); ok {
		model.Role = types.StringValue(role)
	} else if model.Role.IsUnknown() {
		model.Role = types.StringNull()
	}

	if cloudKind, ok := response["cloudKind"].(string); ok {
		model.CloudKind = types.StringValue(cloudKind)
	} else if model.CloudKind.IsUnknown() {
		model.CloudKind = types.StringNull()
	}

	if validate, ok := response["validate"].(bool); ok {
		model.Validate = types.BoolValue(validate)
	} else if model.Validate.IsUnknown() {
		model.Validate = types.BoolNull()
	}
}
