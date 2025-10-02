package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_cross_account_iam_role_metadatas"
)

var _ datasource.DataSource = (*crossAccountIamRoleMetadatasDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*crossAccountIamRoleMetadatasDataSource)(nil)

func NewCrossAccountIamRoleMetadatasDataSource() datasource.DataSource {
	return &crossAccountIamRoleMetadatasDataSource{}
}

type crossAccountIamRoleMetadatasDataSource struct {
	client *client.GalaxyClient
}

func (d *crossAccountIamRoleMetadatasDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cross_account_iam_role_metadatas"
}

func (d *crossAccountIamRoleMetadatasDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_cross_account_iam_role_metadatas.CrossAccountIamRoleMetadatasDataSourceSchema(ctx)
}

func (d *crossAccountIamRoleMetadatasDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.GalaxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.GalaxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *crossAccountIamRoleMetadatasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_cross_account_iam_role_metadatas.CrossAccountIamRoleMetadatasModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading cross_account_iam_role_metadatas data source")
	response, err := d.client.ListCrossAccountIamRoleMetadatas(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading cross account IAM role metadatas",
			"Could not read cross account IAM role metadatas: "+err.Error(),
		)
		return
	}

	d.updateModelFromResponse(ctx, &config, response)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *crossAccountIamRoleMetadatasDataSource) updateModelFromResponse(ctx context.Context, model *datasource_cross_account_iam_role_metadatas.CrossAccountIamRoleMetadatasModel, response map[string]interface{}) {
	if externalId, ok := response["externalId"].(string); ok {
		model.ExternalId = types.StringValue(externalId)
	}
	if starburstAwsAccountId, ok := response["starburstAwsAccountId"].(string); ok {
		model.StarburstAwsAccountId = types.StringValue(starburstAwsAccountId)
	}
}
