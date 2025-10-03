package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_redshift_catalogs"
)

var _ datasource.DataSource = (*redshift_catalogsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*redshift_catalogsDataSource)(nil)

func NewRedshiftCatalogsDataSource() datasource.DataSource {
	return &redshift_catalogsDataSource{}
}

type redshift_catalogsDataSource struct {
	client *client.GalaxyClient
}

func (d *redshift_catalogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redshift_catalogs"
}

func (d *redshift_catalogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_redshift_catalogs.RedshiftCatalogsDataSourceSchema(ctx)
}

func (d *redshift_catalogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *redshift_catalogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_redshift_catalogs.RedshiftCatalogsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading redshift_catalogs with automatic pagination")

	// Use automatic pagination to get ALL Redshift catalogs across all pages
	allCatalogs, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/catalog?catalogType=REDSHIFT")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading redshift_catalogs",
			"Could not read redshift_catalogs: "+err.Error(),
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
		catalogs, err := d.mapRedshiftCatalogsResult(ctx, catalogMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping redshift_catalogs response",
				"Could not map redshift_catalogs response: "+err.Error(),
			)
			return
		}
		config.Result = catalogs
	} else {
		elementType := datasource_redshift_catalogs.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_redshift_catalogs.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_redshift_catalogs.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *redshift_catalogsDataSource) mapRedshiftCatalogsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	catalogs := make([]datasource_redshift_catalogs.ResultValue, 0)

	for _, catalogMap := range result {
		catalog := d.mapSingleRedshiftCatalog(ctx, catalogMap)
		catalogs = append(catalogs, catalog)
	}

	elementType := datasource_redshift_catalogs.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_redshift_catalogs.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, catalogs)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *redshift_catalogsDataSource) mapSingleRedshiftCatalog(ctx context.Context, catalogMap map[string]interface{}) datasource_redshift_catalogs.ResultValue {
	attributeTypes := datasource_redshift_catalogs.ResultValue{}.AttributeTypes(ctx)
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

	// Map endpoint
	if endpoint, ok := catalogMap["endpoint"].(string); ok {
		attributes["endpoint"] = types.StringValue(endpoint)
	} else if host, ok := catalogMap["host"].(string); ok {
		// Construct endpoint from host and port if available
		if port, ok := catalogMap["port"].(float64); ok {
			attributes["endpoint"] = types.StringValue(fmt.Sprintf("%s:%d", host, int(port)))
		} else {
			attributes["endpoint"] = types.StringValue(host)
		}
	} else {
		attributes["endpoint"] = types.StringNull()
	}

	// Handle validate field
	if validate, ok := catalogMap["validate"].(bool); ok {
		attributes["validate"] = types.BoolValue(validate)
	} else {
		attributes["validate"] = types.BoolNull()
	}
	// Create the ResultValue using the constructor
	catalog, diags := datasource_redshift_catalogs.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating Redshift catalog ResultValue: %v", diags))
		return datasource_redshift_catalogs.NewResultValueNull()
	}

	return catalog
}
