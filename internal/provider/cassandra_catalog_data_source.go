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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_cassandra_catalog"
)

var _ datasource.DataSource = (*cassandra_catalogDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*cassandra_catalogDataSource)(nil)

func NewCassandraCatalogDataSource() datasource.DataSource {
	return &cassandra_catalogDataSource{}
}

type cassandra_catalogDataSource struct {
	client *client.GalaxyClient
}

func (d *cassandra_catalogDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cassandra_catalog"
}

func (d *cassandra_catalogDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_cassandra_catalog.CassandraCatalogDataSourceSchema(ctx)
}

func (d *cassandra_catalogDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *cassandra_catalogDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_cassandra_catalog.CassandraCatalogModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.CatalogId.ValueString()
	tflog.Debug(ctx, "Reading cassandra_catalog", map[string]interface{}{"id": id})

	response, err := d.client.GetCatalog(ctx, "cassandra", id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading cassandra_catalog",
			"Could not read cassandra_catalog "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *cassandra_catalogDataSource) updateModelFromResponse(ctx context.Context, model *datasource_cassandra_catalog.CassandraCatalogModel, response map[string]interface{}) {
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

	if deploymentType, ok := response["deploymentType"].(string); ok {
		model.DeploymentType = types.StringValue(deploymentType)
	} else if model.DeploymentType.IsUnknown() {
		model.DeploymentType = types.StringNull()
	}

	if contactPoints, ok := response["contactPoints"].(string); ok {
		model.ContactPoints = types.StringValue(contactPoints)
	} else if model.ContactPoints.IsUnknown() {
		model.ContactPoints = types.StringNull()
	}

	if localDatacenter, ok := response["localDatacenter"].(string); ok {
		model.LocalDatacenter = types.StringValue(localDatacenter)
	} else if model.LocalDatacenter.IsUnknown() {
		model.LocalDatacenter = types.StringNull()
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

	if cloudKind, ok := response["cloudKind"].(string); ok {
		model.CloudKind = types.StringValue(cloudKind)
	} else if model.CloudKind.IsUnknown() {
		model.CloudKind = types.StringNull()
	}

	if databaseId, ok := response["databaseId"].(string); ok {
		model.DatabaseId = types.StringValue(databaseId)
	} else if model.DatabaseId.IsUnknown() {
		model.DatabaseId = types.StringNull()
	}

	if region, ok := response["region"].(string); ok {
		model.Region = types.StringValue(region)
	} else if model.Region.IsUnknown() {
		model.Region = types.StringNull()
	}

	if sshTunnelId, ok := response["sshTunnelId"].(string); ok {
		model.SshTunnelId = types.StringValue(sshTunnelId)
	} else if model.SshTunnelId.IsUnknown() {
		model.SshTunnelId = types.StringNull()
	}

	if token, ok := response["token"].(string); ok {
		model.Token = types.StringValue(token)
	} else if model.Token.IsUnknown() {
		model.Token = types.StringNull()
	}

	if validate, ok := response["validate"].(bool); ok {
		model.Validate = types.BoolValue(validate)
	} else if model.Validate.IsUnknown() {
		model.Validate = types.BoolNull()
	}
}
