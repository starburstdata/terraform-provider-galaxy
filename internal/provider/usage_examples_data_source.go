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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_usage_examples"
)

var _ datasource.DataSource = (*usageExamplesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*usageExamplesDataSource)(nil)

func NewUsageExamplesDataSource() datasource.DataSource {
	return &usageExamplesDataSource{}
}

type usageExamplesDataSource struct {
	client *client.GalaxyClient
}

func (d *usageExamplesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_usage_examples"
}

func (d *usageExamplesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_usage_examples.UsageExamplesDataSourceSchema(ctx)
}

func (d *usageExamplesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *usageExamplesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_usage_examples.UsageExamplesModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dataProductID := config.DataProductId.ValueString()
	tflog.Debug(ctx, "Reading usage examples", map[string]interface{}{"data_product_id": dataProductID})

	response, err := d.client.ListUsageExamples(ctx, dataProductID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading usage examples",
			"Could not read usage examples for data product "+dataProductID+": "+err.Error(),
		)
		return
	}

	// Map results
	elementType := datasource_usage_examples.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_usage_examples.ResultValue{}.AttributeTypes(ctx),
		},
	}

	var results []interface{}
	if items, ok := response["items"].([]interface{}); ok {
		results = items
	} else if arr, ok := response["result"].([]interface{}); ok {
		results = arr
	}

	if len(results) > 0 {
		var resultElements []attr.Value
		for _, itemRaw := range results {
			if item, ok := itemRaw.(map[string]interface{}); ok {
				resultValue := d.mapSingleResult(item)
				resultElements = append(resultElements, resultValue)
			}
		}
		config.Result, _ = types.ListValue(elementType, resultElements)
	} else {
		config.Result, _ = types.ListValue(elementType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *usageExamplesDataSource) mapSingleResult(item map[string]interface{}) datasource_usage_examples.ResultValue {
	result := datasource_usage_examples.ResultValue{}

	if v, ok := item["usageExampleId"].(string); ok {
		result.UsageExampleId = types.StringValue(v)
	} else {
		result.UsageExampleId = types.StringNull()
	}
	if v, ok := item["dataProductId"].(string); ok {
		result.DataProductId = types.StringValue(v)
	} else {
		result.DataProductId = types.StringNull()
	}
	if v, ok := item["name"].(string); ok {
		result.Name = types.StringValue(v)
	} else {
		result.Name = types.StringNull()
	}
	if v, ok := item["code"].(string); ok {
		result.Code = types.StringValue(v)
	} else {
		result.Code = types.StringNull()
	}
	if v, ok := item["description"].(string); ok {
		result.Description = types.StringValue(v)
	} else {
		result.Description = types.StringNull()
	}
	if v, ok := item["suggestedPrompt"].(string); ok {
		result.SuggestedPrompt = types.StringValue(v)
	} else {
		result.SuggestedPrompt = types.StringNull()
	}
	if v, ok := item["createdOn"].(string); ok {
		result.CreatedOn = types.StringValue(v)
	} else {
		result.CreatedOn = types.StringNull()
	}
	if v, ok := item["modifiedOn"].(string); ok {
		result.ModifiedOn = types.StringValue(v)
	} else {
		result.ModifiedOn = types.StringNull()
	}

	return result
}
