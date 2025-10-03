package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_schema"
)

var _ datasource.DataSource = (*schemaDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*schemaDataSource)(nil)

func NewSchemaDataSource() datasource.DataSource {
	return &schemaDataSource{}
}

type schemaDataSource struct {
	client *client.GalaxyClient
}

func (d *schemaDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schema"
}

func (d *schemaDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_schema.SchemaDataSourceSchema(ctx)
}

func (d *schemaDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *schemaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_schema.SchemaModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := config.Id.ValueString()
	tflog.Debug(ctx, "Reading schemas", map[string]interface{}{"catalogId": catalogID})

	response, err := d.client.ListSchemas(ctx, catalogID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading schemas",
			"Could not read schemas for catalog "+catalogID+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *schemaDataSource) updateModelFromResponse(ctx context.Context, model *datasource_schema.SchemaModel, response map[string]interface{}) {
	// The id (catalog ID) is already set from the configuration

	// Map the result array
	if resultArray, ok := response["result"].([]interface{}); ok {
		resultList := make([]datasource_schema.ResultValue, 0, len(resultArray))
		for _, item := range resultArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				resultItem := datasource_schema.ResultValue{
					SchemaId:    types.StringValue(getStringFromMap(itemMap, "schemaId")),
					Description: types.StringValue(getStringFromMap(itemMap, "description")),
				}

				// Map contacts
				if contacts, ok := itemMap["contacts"].([]interface{}); ok {
					contactList := make([]datasource_schema.ContactsValue, 0, len(contacts))
					for _, contact := range contacts {
						if contactMap, ok := contact.(map[string]interface{}); ok {
							contactItem := datasource_schema.ContactsValue{
								UserId: types.StringValue(getStringFromMap(contactMap, "userId")),
								Email:  types.StringValue(getStringFromMap(contactMap, "email")),
							}
							contactList = append(contactList, contactItem)
						}
					}
					contactListValue, _ := types.ListValueFrom(ctx, datasource_schema.ContactsType{}.ValueType(ctx).Type(ctx), contactList)
					resultItem.Contacts = contactListValue
				} else {
					resultItem.Contacts = types.ListNull(datasource_schema.ContactsType{}.ValueType(ctx).Type(ctx))
				}

				// Map owner as ObjectValue (not OwnerValue struct)
				if owner, ok := itemMap["owner"].(map[string]interface{}); ok {
					ownerAttrs := map[string]attr.Value{
						"role_id":   types.StringValue(getStringFromMap(owner, "roleId")),
						"role_name": types.StringValue(getStringFromMap(owner, "roleName")),
					}
					ownerObj, _ := types.ObjectValue(map[string]attr.Type{
						"role_id":   types.StringType,
						"role_name": types.StringType,
					}, ownerAttrs)
					resultItem.Owner = ownerObj
				} else {
					resultItem.Owner = types.ObjectNull(map[string]attr.Type{
						"role_id":   types.StringType,
						"role_name": types.StringType,
					})
				}

				// Map tags
				if tags, ok := itemMap["tags"].([]interface{}); ok {
					tagList := make([]datasource_schema.TagsValue, 0, len(tags))
					for _, tag := range tags {
						if tagMap, ok := tag.(map[string]interface{}); ok {
							tagItem := datasource_schema.TagsValue{
								TagId: types.StringValue(getStringFromMap(tagMap, "tagId")),
								Name:  types.StringValue(getStringFromMap(tagMap, "name")),
							}
							tagList = append(tagList, tagItem)
						}
					}
					tagListValue, _ := types.ListValueFrom(ctx, datasource_schema.TagsType{}.ValueType(ctx).Type(ctx), tagList)
					resultItem.Tags = tagListValue
				} else {
					resultItem.Tags = types.ListNull(datasource_schema.TagsType{}.ValueType(ctx).Type(ctx))
				}

				// Map links - initially set to null list
				resultItem.Links = types.ListNull(types.StringType)

				resultList = append(resultList, resultItem)
			}
		}
		resultListValue, _ := types.ListValueFrom(ctx, datasource_schema.ResultType{}.ValueType(ctx).Type(ctx), resultList)
		model.Result = resultListValue
	} else {
		model.Result = types.ListNull(datasource_schema.ResultType{}.ValueType(ctx).Type(ctx))
	}
}
