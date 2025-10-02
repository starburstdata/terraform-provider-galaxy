package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_row_filters"
)

var _ datasource.DataSource = (*row_filtersDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*row_filtersDataSource)(nil)

func NewRowFiltersDataSource() datasource.DataSource {
	return &row_filtersDataSource{}
}

type row_filtersDataSource struct {
	client *client.GalaxyClient
}

func (d *row_filtersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_row_filters"
}

func (d *row_filtersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_row_filters.RowFiltersDataSourceSchema(ctx)
}

func (d *row_filtersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *row_filtersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_row_filters.RowFiltersModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading row_filters with automatic pagination")

	// Use automatic pagination to get ALL row filters across all pages
	allRowFilters, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/rowFilter")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading row_filters",
			"Could not read row_filters: "+err.Error(),
		)
		return
	}

	// Convert []interface{} to []map[string]interface{} for mapping
	var rowFilterMaps []map[string]interface{}
	for _, rowFilterInterface := range allRowFilters {
		if rowFilterMap, ok := rowFilterInterface.(map[string]interface{}); ok {
			rowFilterMaps = append(rowFilterMaps, rowFilterMap)
		}
	}

	// Map API response to model
	if len(rowFilterMaps) > 0 {
		rowFilters, err := d.mapRowFiltersResult(ctx, rowFilterMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping row_filters response",
				"Could not map row_filters response: "+err.Error(),
			)
			return
		}
		config.Result = rowFilters
	} else {
		elementType := datasource_row_filters.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_row_filters.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_row_filters.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *row_filtersDataSource) mapRowFiltersResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	rowFilters := make([]datasource_row_filters.ResultValue, 0)

	for _, rowFilterMap := range result {
		rowFilter := d.mapSingleRowFilter(ctx, rowFilterMap)
		rowFilters = append(rowFilters, rowFilter)
	}

	elementType := datasource_row_filters.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_row_filters.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, rowFilters)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *row_filtersDataSource) mapSingleRowFilter(ctx context.Context, rowFilterMap map[string]interface{}) datasource_row_filters.ResultValue {
	attributeTypes := datasource_row_filters.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map row filter ID
	if rowFilterId, ok := rowFilterMap["rowFilterId"].(string); ok {
		attributes["row_filter_id"] = types.StringValue(rowFilterId)
	} else {
		attributes["row_filter_id"] = types.StringNull()
	}

	// Map name
	if name, ok := rowFilterMap["name"].(string); ok {
		attributes["name"] = types.StringValue(name)
	} else {
		attributes["name"] = types.StringNull()
	}

	// Map description
	if description, ok := rowFilterMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	// Map expression
	if expression, ok := rowFilterMap["expression"].(string); ok {
		attributes["expression"] = types.StringValue(expression)
	} else {
		attributes["expression"] = types.StringNull()
	}

	// Map created
	if created, ok := rowFilterMap["created"].(string); ok {
		attributes["created"] = types.StringValue(created)
	} else {
		attributes["created"] = types.StringNull()
	}

	// Map modified
	if modified, ok := rowFilterMap["modified"].(string); ok {
		attributes["modified"] = types.StringValue(modified)
	} else {
		attributes["modified"] = types.StringNull()
	}

	// Create the ResultValue using the constructor
	rowFilter, diags := datasource_row_filters.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating row filter ResultValue: %v", diags))
		return datasource_row_filters.NewResultValueNull()
	}

	return rowFilter
}
