package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_bigquery_catalogs"
)

var _ datasource.DataSource = (*bigquery_catalogsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*bigquery_catalogsDataSource)(nil)

func NewBigqueryCatalogsDataSource() datasource.DataSource {
	return &bigquery_catalogsDataSource{}
}

type bigquery_catalogsDataSource struct {
	client *client.GalaxyClient
}

func (d *bigquery_catalogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bigquery_catalogs"
}

func (d *bigquery_catalogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_bigquery_catalogs.BigqueryCatalogsDataSourceSchema(ctx)
}

func (d *bigquery_catalogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *bigquery_catalogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_bigquery_catalogs.BigqueryCatalogsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading bigquery_catalogs with automatic pagination")

	// Use automatic pagination to get ALL bigquery catalogs across all pages
	allCatalogs, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/catalog?catalogType=BIGQUERY")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bigquery_catalogs",
			"Could not read bigquery_catalogs: "+err.Error(),
		)
		return
	}

	// Convert []interface{} to []map[string]interface{} for mapping
	var catalogMaps []map[string]interface{}
	for _, catalogInterface := range allCatalogs {
		if catalogMap, ok := catalogInterface.(map[string]interface{}); ok {
			catalogMaps = append(catalogMaps, catalogMap)
		}
	}

	// Map API response to model
	if len(catalogMaps) > 0 {
		catalogs, err := d.mapBigqueryCatalogsResult(ctx, catalogMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping bigquery_catalogs response",
				"Could not map bigquery_catalogs response: "+err.Error(),
			)
			return
		}
		config.Result = catalogs
	} else {
		elementType := datasource_bigquery_catalogs.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_bigquery_catalogs.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_bigquery_catalogs.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *bigquery_catalogsDataSource) mapBigqueryCatalogsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	catalogs := make([]datasource_bigquery_catalogs.ResultValue, 0, len(result))

	for _, catalogMap := range result {
		catalog := d.mapSingleBigqueryCatalog(ctx, catalogMap)
		catalogs = append(catalogs, catalog)
	}

	elementType := datasource_bigquery_catalogs.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_bigquery_catalogs.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, catalogs)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *bigquery_catalogsDataSource) mapSingleBigqueryCatalog(ctx context.Context, catalogMap map[string]interface{}) datasource_bigquery_catalogs.ResultValue {
	attributeTypes := datasource_bigquery_catalogs.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map catalog ID
	if catalogId, ok := catalogMap["catalogId"].(string); ok {
		attributes["catalog_id"] = types.StringValue(catalogId)
	} else {
		attributes["catalog_id"] = types.StringNull()
	}

	// Map name
	if name, ok := catalogMap["name"].(string); ok {
		attributes["name"] = types.StringValue(name)
	} else {
		attributes["name"] = types.StringNull()
	}

	// Map description
	if description, ok := catalogMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	// Handle validate field
	if validate, ok := catalogMap["validate"].(bool); ok {
		attributes["validate"] = types.BoolValue(validate)
	} else {
		attributes["validate"] = types.BoolNull()
	}
	// Create the ResultValue using the constructor
	catalog, diags := datasource_bigquery_catalogs.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating catalog ResultValue: %v", diags))
		return datasource_bigquery_catalogs.NewResultValueNull()
	}

	return catalog
}
