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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_bigquery_catalog_validation"
)

var _ datasource.DataSource = (*bigquery_catalog_validationDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*bigquery_catalog_validationDataSource)(nil)

func NewBigqueryCatalogValidationDataSource() datasource.DataSource {
	return &bigquery_catalog_validationDataSource{}
}

type bigquery_catalog_validationDataSource struct {
	client *client.GalaxyClient
}

func (d *bigquery_catalog_validationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bigquery_catalog_validation"
}

func (d *bigquery_catalog_validationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_bigquery_catalog_validation.BigqueryCatalogValidationDataSourceSchema(ctx)
}

func (d *bigquery_catalog_validationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bigquery_catalog_validationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_bigquery_catalog_validation.BigqueryCatalogValidationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading bigquery_catalog_validation", map[string]interface{}{"id": id})

	response, err := d.client.GetCatalog(ctx, "bigquery", id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bigquery_catalog_validation",
			"Could not read bigquery_catalog_validation "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *bigquery_catalog_validationDataSource) updateModelFromResponse(ctx context.Context, model *datasource_bigquery_catalog_validation.BigqueryCatalogValidationModel, response map[string]interface{}) {
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	}

	if validationSuccessful, ok := response["validationSuccessful"].(bool); ok {
		model.ValidationSuccessful = types.BoolValue(validationSuccessful)
	}

	// Handle error messages list
	if errorMessages, ok := response["errorMessages"].([]interface{}); ok {
		var errorValues []types.String
		for _, msg := range errorMessages {
			if msgStr, ok := msg.(string); ok {
				errorValues = append(errorValues, types.StringValue(msgStr))
			}
		}
		if len(errorValues) > 0 {
			model.ErrorMessages, _ = types.ListValueFrom(ctx, types.StringType, errorValues)
		} else {
			model.ErrorMessages = types.ListNull(types.StringType)
		}
	} else {
		model.ErrorMessages = types.ListNull(types.StringType)
	}

	// Handle warning messages list
	if warningMessages, ok := response["warningMessages"].([]interface{}); ok {
		var warningValues []types.String
		for _, msg := range warningMessages {
			if msgStr, ok := msg.(string); ok {
				warningValues = append(warningValues, types.StringValue(msgStr))
			}
		}
		if len(warningValues) > 0 {
			model.WarningMessages, _ = types.ListValueFrom(ctx, types.StringType, warningValues)
		} else {
			model.WarningMessages = types.ListNull(types.StringType)
		}
	} else {
		model.WarningMessages = types.ListNull(types.StringType)
	}

	// Handle info messages list
	if infoMessages, ok := response["infoMessages"].([]interface{}); ok {
		var infoValues []types.String
		for _, msg := range infoMessages {
			if msgStr, ok := msg.(string); ok {
				infoValues = append(infoValues, types.StringValue(msgStr))
			}
		}
		if len(infoValues) > 0 {
			model.InfoMessages, _ = types.ListValueFrom(ctx, types.StringType, infoValues)
		} else {
			model.InfoMessages = types.ListNull(types.StringType)
		}
	} else {
		model.InfoMessages = types.ListNull(types.StringType)
	}
}
