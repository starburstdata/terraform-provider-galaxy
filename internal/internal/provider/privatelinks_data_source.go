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

	d.updateModelFromResponse(ctx, &config, response)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *privatelinksDataSource) updateModelFromResponse(ctx context.Context, model *datasource_privatelinks.PrivatelinksModel, response map[string]interface{}) {
	// Extract the result array from the API response
	resultInterface, ok := response["result"]
	if !ok {
		return
	}

	resultArray, ok := resultInterface.([]interface{})
	if !ok {
		return
	}

	// Convert to list of Terraform objects
	var resultList []attr.Value
	for _, item := range resultArray {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Create a ResultValue for each privatelink
		resultValue := datasource_privatelinks.ResultValue{
			CloudRegionId: types.StringValue(""),
			Name:          types.StringValue(""),
			PrivatelinkId: types.StringValue(""),
		}

		if cloudRegionId, ok := itemMap["cloudRegionId"].(string); ok {
			resultValue.CloudRegionId = types.StringValue(cloudRegionId)
		}
		if name, ok := itemMap["name"].(string); ok {
			resultValue.Name = types.StringValue(name)
		}
		if privatelinkId, ok := itemMap["privatelinkId"].(string); ok {
			resultValue.PrivatelinkId = types.StringValue(privatelinkId)
		}

		objectValue, diags := resultValue.ToObjectValue(ctx)
		if diags.HasError() {
			tflog.Error(ctx, "Error converting privatelink result to object value", map[string]interface{}{"errors": diags})
			continue
		}

		resultList = append(resultList, objectValue)
	}

	// Convert to Terraform list
	resultListValue, diags := types.ListValue(datasource_privatelinks.ResultType{}.ValueType(ctx).Type(ctx), resultList)
	if diags.HasError() {
		tflog.Error(ctx, "Error creating privatelinks result list", map[string]interface{}{"errors": diags})
		model.Result = types.ListNull(datasource_privatelinks.ResultType{}.ValueType(ctx).Type(ctx))
	} else {
		model.Result = resultListValue
	}
}
