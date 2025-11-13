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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_privatelinks"
)

var _ datasource.DataSource = (*privatelinksDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*privatelinksDataSource)(nil)

func NewPrivatelinksDataSource() datasource.DataSource {
	return &privatelinksDataSource{}
}

type privatelinksDataSource struct {
	client *client.GalaxyClient
}

func (d *privatelinksDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_privatelinks"
}

func (d *privatelinksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_privatelinks.PrivatelinksDataSourceSchema(ctx)
}

func (d *privatelinksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *privatelinksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_privatelinks.PrivatelinksModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading privatelinks data source")
	response, err := d.client.ListPrivatelinks(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading privatelinks",
			"Could not read privatelinks: "+err.Error(),
		)
		return
	}

	d.updateModelFromResponse(ctx, resp, &config, response)

	// Check for any errors before updating state
	if resp.Diagnostics.HasError() {
		return
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *privatelinksDataSource) updateModelFromResponse(ctx context.Context, resp *datasource.ReadResponse, model *datasource_privatelinks.PrivatelinksModel, response map[string]interface{}) {
	// Extract the result array from the API response
	resultInterface, ok := response["result"]
	if !ok {
		// If no result field, set to empty list
		elementType := datasource_privatelinks.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_privatelinks.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, diags := types.ListValueFrom(ctx, elementType, []datasource_privatelinks.ResultValue{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		model.Result = emptyList
		return
	}

	resultArray, ok := resultInterface.([]interface{})
	if !ok {
		// If result is not an array, set to empty list
		elementType := datasource_privatelinks.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_privatelinks.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, diags := types.ListValueFrom(ctx, elementType, []datasource_privatelinks.ResultValue{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		model.Result = emptyList
		return
	}

	// Convert to list of ResultValue objects using the proper constructor
	attributeTypes := datasource_privatelinks.ResultValue{}.AttributeTypes(ctx)
	// Use make() to create empty slice, not nil - nil slice converts to null list
	resultList := make([]datasource_privatelinks.ResultValue, 0, len(resultArray))

	for _, item := range resultArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Create attributes map
		attributes := make(map[string]attr.Value)

		if cloudRegionId, ok := itemMap["cloudRegionId"].(string); ok {
			attributes["cloud_region_id"] = types.StringValue(cloudRegionId)
		} else {
			attributes["cloud_region_id"] = types.StringNull()
		}

		if name, ok := itemMap["name"].(string); ok {
			attributes["name"] = types.StringValue(name)
		} else {
			attributes["name"] = types.StringNull()
		}

		if privatelinkId, ok := itemMap["privatelinkId"].(string); ok {
			attributes["privatelink_id"] = types.StringValue(privatelinkId)
		} else {
			attributes["privatelink_id"] = types.StringNull()
		}

		// Use the proper constructor to create ResultValue
		resultValue, diags := datasource_privatelinks.NewResultValue(attributeTypes, attributes)
		if diags.HasError() {
			tflog.Error(ctx, "Error creating privatelink ResultValue", map[string]interface{}{"errors": diags})
			continue
		}

		resultList = append(resultList, resultValue)
	}

	// Convert to Terraform list
	elementType := datasource_privatelinks.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: attributeTypes,
		},
	}
	resultListValue, diags := types.ListValueFrom(ctx, elementType, resultList)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	model.Result = resultListValue
}
