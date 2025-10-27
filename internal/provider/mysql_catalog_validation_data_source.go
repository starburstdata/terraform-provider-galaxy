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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_mysql_catalog_validation"
)

var _ datasource.DataSource = (*mysql_catalog_validationDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*mysql_catalog_validationDataSource)(nil)

func NewMysqlCatalogValidationDataSource() datasource.DataSource {
	return &mysql_catalog_validationDataSource{}
}

type mysql_catalog_validationDataSource struct {
	client *client.GalaxyClient
}

func (d *mysql_catalog_validationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql_catalog_validation"
}

func (d *mysql_catalog_validationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_mysql_catalog_validation.MysqlCatalogValidationDataSourceSchema(ctx)
}

func (d *mysql_catalog_validationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *mysql_catalog_validationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_mysql_catalog_validation.MysqlCatalogValidationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading mysql_catalog_validation", map[string]interface{}{"id": id})

	response, err := d.client.ValidateCatalog(ctx, "mysql", id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading mysql_catalog_validation",
			"Could not read mysql_catalog_validation "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *mysql_catalog_validationDataSource) updateModelFromResponse(ctx context.Context, model *datasource_mysql_catalog_validation.MysqlCatalogValidationModel, response map[string]interface{}) {
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	}

	// Map validation_successful
	if validationSuccessful, ok := response["validationSuccessful"].(bool); ok {
		model.ValidationSuccessful = types.BoolValue(validationSuccessful)
	} else {
		model.ValidationSuccessful = types.BoolNull()
	}

	// Map error_messages
	if errorMessages, ok := response["errorMessages"].([]interface{}); ok && len(errorMessages) > 0 {
		stringErrors := make([]string, len(errorMessages))
		for i, msg := range errorMessages {
			if strMsg, ok := msg.(string); ok {
				stringErrors[i] = strMsg
			}
		}
		if len(stringErrors) > 0 {
			errorList, _ := types.ListValueFrom(ctx, types.StringType, stringErrors)
			model.ErrorMessages = errorList
		} else {
			model.ErrorMessages = types.ListNull(types.StringType)
		}
	} else {
		model.ErrorMessages = types.ListNull(types.StringType)
	}

	// Map info_messages
	if infoMessages, ok := response["infoMessages"].([]interface{}); ok && len(infoMessages) > 0 {
		stringInfos := make([]string, len(infoMessages))
		for i, msg := range infoMessages {
			if strMsg, ok := msg.(string); ok {
				stringInfos[i] = strMsg
			}
		}
		if len(stringInfos) > 0 {
			infoList, _ := types.ListValueFrom(ctx, types.StringType, stringInfos)
			model.InfoMessages = infoList
		} else {
			model.InfoMessages = types.ListNull(types.StringType)
		}
	} else {
		model.InfoMessages = types.ListNull(types.StringType)
	}

	// Map warning_messages
	if warningMessages, ok := response["warningMessages"].([]interface{}); ok && len(warningMessages) > 0 {
		stringWarnings := make([]string, len(warningMessages))
		for i, msg := range warningMessages {
			if strMsg, ok := msg.(string); ok {
				stringWarnings[i] = strMsg
			}
		}
		if len(stringWarnings) > 0 {
			warningList, _ := types.ListValueFrom(ctx, types.StringType, stringWarnings)
			model.WarningMessages = warningList
		} else {
			model.WarningMessages = types.ListNull(types.StringType)
		}
	} else {
		model.WarningMessages = types.ListNull(types.StringType)
	}
}
