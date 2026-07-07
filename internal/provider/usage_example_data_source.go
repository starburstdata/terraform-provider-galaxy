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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_usage_example"
)

var _ datasource.DataSource = (*usageExampleDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*usageExampleDataSource)(nil)

func NewUsageExampleDataSource() datasource.DataSource {
	return &usageExampleDataSource{}
}

type usageExampleDataSource struct {
	client *client.GalaxyClient
}

func (d *usageExampleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_usage_example"
}

func (d *usageExampleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_usage_example.UsageExampleDataSourceSchema(ctx)
}

func (d *usageExampleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *usageExampleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_usage_example.UsageExampleModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Client",
			"Cannot perform data source operation: client is not configured. Please ensure the provider configuration is complete.",
		)
		return
	}

	dataProductId := config.DataProductId.ValueString()
	tflog.Debug(ctx, "Reading usage examples", map[string]interface{}{"data_product_id": dataProductId})

	path := fmt.Sprintf("/public/api/v1/dataProduct/%s/usageExample", dataProductId)
	allResults, err := d.client.GetAllPaginatedResults(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading usage examples",
			"Could not read usage examples: "+err.Error(),
		)
		return
	}

	var resultMaps []map[string]interface{}
	for _, resultInterface := range allResults {
		if resultMap, ok := resultInterface.(map[string]interface{}); ok {
			resultMaps = append(resultMaps, resultMap)
		} else {
			tflog.Warn(ctx, fmt.Sprintf("unexpected entry type in usage examples list, skipping: %T", resultInterface))
		}
	}

	if len(resultMaps) > 0 {
		results, err := d.mapResults(ctx, resultMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping usage examples response",
				"Could not map usage examples response: "+err.Error(),
			)
			return
		}
		config.Result = results
	} else {
		elementType := datasource_usage_example.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_usage_example.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_usage_example.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *usageExampleDataSource) mapResults(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	items := make([]datasource_usage_example.ResultValue, 0)

	for _, itemMap := range result {
		item := d.mapSingleResult(ctx, itemMap)
		items = append(items, item)
	}

	elementType := datasource_usage_example.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_usage_example.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, items)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *usageExampleDataSource) mapSingleResult(ctx context.Context, itemMap map[string]interface{}) datasource_usage_example.ResultValue {
	attributeTypes := datasource_usage_example.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	if code, ok := itemMap["code"].(string); ok {
		attributes["code"] = types.StringValue(code)
	} else {
		attributes["code"] = types.StringNull()
	}

	if createdOn, ok := itemMap["createdOn"].(string); ok {
		attributes["created_on"] = types.StringValue(createdOn)
	} else {
		attributes["created_on"] = types.StringNull()
	}

	if dataProductId, ok := itemMap["dataProductId"].(string); ok {
		attributes["data_product_id"] = types.StringValue(dataProductId)
	} else {
		attributes["data_product_id"] = types.StringNull()
	}

	if description, ok := itemMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	if modifiedOn, ok := itemMap["modifiedOn"].(string); ok {
		attributes["modified_on"] = types.StringValue(modifiedOn)
	} else {
		attributes["modified_on"] = types.StringNull()
	}

	if name, ok := itemMap["name"].(string); ok {
		attributes["name"] = types.StringValue(name)
	} else {
		attributes["name"] = types.StringNull()
	}

	if suggestedPrompt, ok := itemMap["suggestedPrompt"].(string); ok {
		attributes["suggested_prompt"] = types.StringValue(suggestedPrompt)
	} else {
		attributes["suggested_prompt"] = types.StringNull()
	}

	if usageExampleId, ok := itemMap["usageExampleId"].(string); ok {
		attributes["usage_example_id"] = types.StringValue(usageExampleId)
	} else {
		attributes["usage_example_id"] = types.StringNull()
	}

	item, diags := datasource_usage_example.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating usage example ResultValue: %v", diags))
		return datasource_usage_example.NewResultValueNull()
	}

	return item
}
