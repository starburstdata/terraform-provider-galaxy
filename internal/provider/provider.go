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
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
)

var _ provider.Provider = (*galaxyProvider)(nil)

func New() func() provider.Provider {
	return func() provider.Provider {
		return &galaxyProvider{}
	}
}

type galaxyProvider struct{}

type galaxyProviderModel struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Domain       types.String `tfsdk:"domain"`
}

func (p *galaxyProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Galaxy OAuth2 Client ID. Can also be set via GALAXY_CLIENT_ID environment variable.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Galaxy OAuth2 Client Secret. Can also be set via GALAXY_CLIENT_SECRET environment variable.",
			},
			"domain": schema.StringAttribute{
				Optional:    true,
				Description: "Galaxy Domain URL. Can also be set via GALAXY_DOMAIN environment variable.",
			},
		},
	}
}

func (p *galaxyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config galaxyProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default values from environment variables
	clientID := os.Getenv("GALAXY_CLIENT_ID")
	if !config.ClientID.IsNull() {
		clientID = config.ClientID.ValueString()
	}

	clientSecret := os.Getenv("GALAXY_CLIENT_SECRET")
	if !config.ClientSecret.IsNull() {
		clientSecret = config.ClientSecret.ValueString()
	}

	domain := os.Getenv("GALAXY_DOMAIN")
	if !config.Domain.IsNull() {
		domain = config.Domain.ValueString()
	}

	// Validate required configuration
	if clientID == "" {
		resp.Diagnostics.AddError(
			"Missing Client ID",
			"The provider cannot create the Galaxy client as there is a missing or empty value for the Galaxy client ID. "+
				"Set the client_id value in the configuration or use the GALAXY_CLIENT_ID environment variable.",
		)
	}

	if clientSecret == "" {
		resp.Diagnostics.AddError(
			"Missing Client Secret",
			"The provider cannot create the Galaxy client as there is a missing or empty value for the Galaxy client secret. "+
				"Set the client_secret value in the configuration or use the GALAXY_CLIENT_SECRET environment variable.",
		)
	}

	if domain == "" {
		resp.Diagnostics.AddError(
			"Missing Domain",
			"The provider cannot create the Galaxy client as there is a missing or empty value for the Galaxy domain. "+
				"Set the domain value in the configuration or use the GALAXY_DOMAIN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the Galaxy client
	client := client.NewGalaxyClient(domain, clientID, clientSecret)

	// Log the successful configuration
	tflog.Info(ctx, "Configured Galaxy client", map[string]interface{}{
		"domain": domain,
	})

	// Store the client in the provider data for use by resources and data sources
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *galaxyProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "galaxy"
}

func (p *galaxyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Single-item data sources
		NewClusterDataSource,
		NewUserDataSource,
		NewRoleDataSource,
		NewServiceAccountDataSource,
		NewColumnMaskDataSource,
		NewRowFilterDataSource,
		NewPolicyDataSource,
		NewDataProductDataSource,
		NewTagDataSource,
		NewCatalogMetadataDataSource,

		// SQL Job data sources
		NewSqlJobDataSource,
		NewSqlJobStatusDataSource,
		NewSqlJobHistoryDataSource,

		// New data sources from OpenAPI changes - implemented
		NewPrivatelinkDataSource,
		NewCrossAccountIamRoleMetadatasDataSource,
		// Newly implemented data sources
		NewDataQualitySummaryDataSource,
		NewTableDataSource,
		NewRolegrantDataSource,
		NewSchemaDataSource,
		NewColumnDataSource,

		// List data sources
		NewClustersDataSource,
		NewUsersDataSource,
		NewRolesDataSource,
		NewServiceAccountsDataSource,
		NewColumnMasksDataSource,
		NewRowFiltersDataSource,
		NewPoliciesDataSource,
		NewDataProductsDataSource,
		NewTagsDataSource,
		NewCatalogsDataSource,
		NewCrossAccountIamRolesDataSource,

		// New list data sources from OpenAPI changes - implemented
		NewSqlJobsDataSource,
		NewPrivatelinksDataSource,
		// Newly implemented list data source
		NewDataQualitySummariesDataSource,

		// Catalog-specific data sources
		NewS3CatalogDataSource,
		NewS3CatalogsDataSource,
		NewRedshiftCatalogDataSource,
		NewRedshiftCatalogsDataSource,
		NewPostgresqlCatalogDataSource,
		NewPostgresqlCatalogsDataSource,
		NewMongodbCatalogDataSource,
		NewMongodbCatalogsDataSource,
		NewCassandraCatalogDataSource,
		NewCassandraCatalogsDataSource,
		NewMysqlCatalogDataSource,
		NewMysqlCatalogsDataSource,
		NewOpensearchCatalogDataSource,
		NewOpensearchCatalogsDataSource,
		NewBigqueryCatalogDataSource,
		NewBigqueryCatalogsDataSource,
		NewSqlserverCatalogDataSource,
		NewSqlserverCatalogsDataSource,
		NewGcsCatalogDataSource,
		NewGcsCatalogsDataSource,
		NewSnowflakeCatalogDataSource,
		NewSnowflakeCatalogsDataSource,

		// Validation data sources
		NewS3CatalogValidationDataSource,
		NewRedshiftCatalogValidationDataSource,
		NewPostgresqlCatalogValidationDataSource,
		NewMongodbCatalogValidationDataSource,
		NewCassandraCatalogValidationDataSource,
		NewMysqlCatalogValidationDataSource,
		NewOpensearchCatalogValidationDataSource,
		NewBigqueryCatalogValidationDataSource,
		NewSqlserverCatalogValidationDataSource,
		NewGcsCatalogValidationDataSource,
		NewSnowflakeCatalogValidationDataSource,
	}
}

func (p *galaxyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Core resources
		NewClusterResource,
		NewRoleResource,
		NewServiceAccountResource,
		NewServiceAccountPasswordResource,

		// Security resources
		NewColumnMaskResource,
		NewRowFilterResource,
		NewPolicyResource,
		NewRolePrivilegeGrantResource,

		// Data resources
		NewDataProductResource,
		NewTagResource,
		NewCrossAccountIamRoleResource,

		// SQL Job resources
		NewSqlJobResource,

		// Catalog resources
		NewS3CatalogResource,
		NewRedshiftCatalogResource,
		NewPostgresqlCatalogResource,
		NewMongodbCatalogResource,
		NewCassandraCatalogResource,
		NewMysqlCatalogResource,
		NewOpensearchCatalogResource,
		NewBigqueryCatalogResource,
		NewSqlserverCatalogResource,
		NewGcsCatalogResource,
		NewSnowflakeCatalogResource,
	}
}
