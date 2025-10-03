package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_catalogs"
)

var _ datasource.DataSource = (*catalogsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*catalogsDataSource)(nil)

func NewCatalogsDataSource() datasource.DataSource {
	return &catalogsDataSource{}
}

type catalogsDataSource struct {
	client *client.GalaxyClient
}

func (d *catalogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_catalogs"
}

func (d *catalogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_catalogs.CatalogsDataSourceSchema(ctx)
}

func (d *catalogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *catalogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_catalogs.CatalogsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading catalogs with automatic pagination")

	// Use automatic pagination to get ALL catalogs across all pages
	allCatalogs, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/catalog")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading catalogs",
			"Could not read catalogs: "+err.Error(),
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
		catalogs, err := d.mapCatalogsResult(ctx, catalogMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping catalogs response",
				"Could not map catalogs response: "+err.Error(),
			)
			return
		}
		config.Result = catalogs
	} else {
		elementType := datasource_catalogs.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_catalogs.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_catalogs.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *catalogsDataSource) mapCatalogsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	catalogs := make([]datasource_catalogs.ResultValue, 0)

	for _, catalogMap := range result {
		catalog := d.mapSingleCatalog(ctx, catalogMap)
		catalogs = append(catalogs, catalog)
	}

	elementType := datasource_catalogs.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_catalogs.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, catalogs)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *catalogsDataSource) mapSingleCatalog(ctx context.Context, catalogMap map[string]interface{}) datasource_catalogs.ResultValue {
	attributeTypes := datasource_catalogs.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map catalog ID
	if catalogId, ok := catalogMap["catalogId"].(string); ok {
		attributes["catalog_id"] = types.StringValue(catalogId)
	} else {
		attributes["catalog_id"] = types.StringNull()
	}

	// Map catalog name
	if catalogName, ok := catalogMap["catalogName"].(string); ok {
		attributes["catalog_name"] = types.StringValue(catalogName)
	} else if catalogName, ok := catalogMap["name"].(string); ok {
		attributes["catalog_name"] = types.StringValue(catalogName)
	} else {
		attributes["catalog_name"] = types.StringNull()
	}

	// Create the ResultValue using the constructor
	catalog, diags := datasource_catalogs.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating catalog ResultValue: %v", diags))
		return datasource_catalogs.NewResultValueNull()
	}

	return catalog
}
