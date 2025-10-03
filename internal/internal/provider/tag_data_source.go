package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_tag"
)

var _ datasource.DataSource = (*tagDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*tagDataSource)(nil)

func NewTagDataSource() datasource.DataSource {
	return &tagDataSource{}
}

type tagDataSource struct {
	client *client.GalaxyClient
}

func (d *tagDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (d *tagDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_tag.TagDataSourceSchema(ctx)
}

func (d *tagDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_tag.TagModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading tag", map[string]interface{}{"id": id})

	response, err := d.client.GetTag(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading tag",
			"Could not read tag "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *tagDataSource) updateModelFromResponse(ctx context.Context, model *datasource_tag.TagModel, response map[string]interface{}) {
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	}
	if tagId, ok := response["id"].(string); ok {
		model.Id = types.StringValue(tagId)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	} else if key, ok := response["key"].(string); ok {
		// Some APIs might use "key" instead of "name"
		model.Name = types.StringValue(key)
	}
	if color, ok := response["color"].(string); ok {
		model.Color = types.StringValue(color)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	}
	// CreatedOn and ModifiedOn fields not present in generated model
}
