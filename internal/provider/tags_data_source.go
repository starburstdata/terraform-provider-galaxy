package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_tags"
)

var _ datasource.DataSource = (*tagsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*tagsDataSource)(nil)

func NewTagsDataSource() datasource.DataSource {
	return &tagsDataSource{}
}

type tagsDataSource struct {
	client *client.GalaxyClient
}

func (d *tagsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tags"
}

func (d *tagsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_tags.TagsDataSourceSchema(ctx)
}

func (d *tagsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_tags.TagsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading tags with automatic pagination")

	// Use automatic pagination to get ALL tags across all pages
	allTags, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/tag")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading tags",
			"Could not read tags: "+err.Error(),
		)
		return
	}

	// Convert []interface{} to []map[string]interface{} for mapping
	var tagMaps []map[string]interface{}
	for _, tagInterface := range allTags {
		if tagMap, ok := tagInterface.(map[string]interface{}); ok {
			tagMaps = append(tagMaps, tagMap)
		}
	}

	// Map API response to model
	if len(tagMaps) > 0 {
		tags, err := d.mapTagsResult(ctx, tagMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping tags response",
				"Could not map tags response: "+err.Error(),
			)
			return
		}
		config.Result = tags
	} else {
		elementType := datasource_tags.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_tags.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_tags.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *tagsDataSource) mapTagsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	tags := make([]datasource_tags.ResultValue, 0)

	for _, tagMap := range result {
		tag := d.mapSingleTag(ctx, tagMap)
		tags = append(tags, tag)
	}

	elementType := datasource_tags.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_tags.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, tags)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *tagsDataSource) mapSingleTag(ctx context.Context, tagMap map[string]interface{}) datasource_tags.ResultValue {
	attributeTypes := datasource_tags.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map tag ID
	if tagId, ok := tagMap["tagId"].(string); ok {
		attributes["tag_id"] = types.StringValue(tagId)
	} else {
		attributes["tag_id"] = types.StringNull()
	}

	// Map name
	if name, ok := tagMap["name"].(string); ok {
		attributes["name"] = types.StringValue(name)
	} else {
		attributes["name"] = types.StringNull()
	}

	// Map description
	if description, ok := tagMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	// Map color
	if color, ok := tagMap["color"].(string); ok {
		attributes["color"] = types.StringValue(color)
	} else {
		attributes["color"] = types.StringNull()
	}

	// Create the ResultValue using the constructor
	tag, diags := datasource_tags.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating tag ResultValue: %v", diags))
		return datasource_tags.NewResultValueNull()
	}

	return tag
}
