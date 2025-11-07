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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_s3_catalog_validation"
)

var _ datasource.DataSource = (*s3_catalog_validationDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*s3_catalog_validationDataSource)(nil)

func NewS3CatalogValidationDataSource() datasource.DataSource {
	return &s3_catalog_validationDataSource{}
}

type s3_catalog_validationDataSource struct {
	client *client.GalaxyClient
}

func (d *s3_catalog_validationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_catalog_validation"
}

func (d *s3_catalog_validationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_s3_catalog_validation.S3CatalogValidationDataSourceSchema(ctx)
}

func (d *s3_catalog_validationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *s3_catalog_validationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_s3_catalog_validation.S3CatalogValidationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogId := config.CatalogId.ValueString()
	tflog.Debug(ctx, "Reading s3_catalog_validation", map[string]interface{}{"catalog_id": catalogId})

	response, err := d.client.ValidateCatalog(ctx, "s3", catalogId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading s3_catalog_validation",
			"Could not read s3_catalog_validation "+catalogId+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *s3_catalog_validationDataSource) updateModelFromResponse(ctx context.Context, model *datasource_s3_catalog_validation.S3CatalogValidationModel, response map[string]interface{}) {
	// Keep the ID from the request
	// model.Id is already set from the request

	// Map validation response fields
	if validationSuccessful, ok := response["validationSuccessful"].(bool); ok {
		model.ValidationSuccessful = types.BoolValue(validationSuccessful)
	}

	// Map error messages
	if errorMessages, ok := response["errorMessages"].([]interface{}); ok {
		errorList := make([]types.String, 0, len(errorMessages))
		for _, msg := range errorMessages {
			if strMsg, ok := msg.(string); ok {
				errorList = append(errorList, types.StringValue(strMsg))
			}
		}
		model.ErrorMessages, _ = types.ListValueFrom(ctx, types.StringType, errorList)
	} else {
		model.ErrorMessages = types.ListNull(types.StringType)
	}

	// Map warning messages
	if warningMessages, ok := response["warningMessages"].([]interface{}); ok {
		warningList := make([]types.String, 0, len(warningMessages))
		for _, msg := range warningMessages {
			if strMsg, ok := msg.(string); ok {
				warningList = append(warningList, types.StringValue(strMsg))
			}
		}
		model.WarningMessages, _ = types.ListValueFrom(ctx, types.StringType, warningList)
	} else {
		model.WarningMessages = types.ListNull(types.StringType)
	}

	// Map info messages
	if infoMessages, ok := response["infoMessages"].([]interface{}); ok {
		infoList := make([]types.String, 0, len(infoMessages))
		for _, msg := range infoMessages {
			if strMsg, ok := msg.(string); ok {
				infoList = append(infoList, types.StringValue(strMsg))
			}
		}
		model.InfoMessages, _ = types.ListValueFrom(ctx, types.StringType, infoList)
	} else {
		model.InfoMessages = types.ListNull(types.StringType)
	}
}
