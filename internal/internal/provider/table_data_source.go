package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_table"
)

var _ datasource.DataSource = (*tableDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*tableDataSource)(nil)

func NewTableDataSource() datasource.DataSource {
	return &tableDataSource{}
}

type tableDataSource struct {
	client *client.GalaxyClient
}

func (d *tableDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_table"
}

func (d *tableDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_table.TableDataSourceSchema(ctx)
}

func (d *tableDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tableDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_table.TableModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := config.CatalogId.ValueString()
	schemaID := config.SchemaId.ValueString()
	tflog.Debug(ctx, "Reading tables", map[string]interface{}{"catalogId": catalogID, "schemaId": schemaID})

	response, err := d.client.ListTables(ctx, catalogID, schemaID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading tables",
			"Could not read tables for catalog "+catalogID+" schema "+schemaID+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *tableDataSource) updateModelFromResponse(ctx context.Context, model *datasource_table.TableModel, response map[string]interface{}) {
	// The catalogId and schemaId are already set from the configuration

	// Map the result array
	if resultArray, ok := response["result"].([]interface{}); ok {
		resultList := make([]datasource_table.ResultValue, 0, len(resultArray))
		for _, item := range resultArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				resultItem := datasource_table.ResultValue{
					TableId:     types.StringValue(getStringFromMap(itemMap, "tableId")),
					TableType:   types.StringValue(getStringFromMap(itemMap, "tableType")),
					Description: types.StringValue(getStringFromMap(itemMap, "description")),
				}

				// Map contacts
				if contacts, ok := itemMap["contacts"].([]interface{}); ok {
					contactList := make([]datasource_table.ContactsValue, 0, len(contacts))
					for _, contact := range contacts {
						if contactMap, ok := contact.(map[string]interface{}); ok {
							contactItem := datasource_table.ContactsValue{
								UserId: types.StringValue(getStringFromMap(contactMap, "userId")),
								Email:  types.StringValue(getStringFromMap(contactMap, "email")),
							}
							contactList = append(contactList, contactItem)
						}
					}
					contactListValue, _ := types.ListValueFrom(ctx, datasource_table.ContactsType{}.ValueType(ctx).Type(ctx), contactList)
					resultItem.Contacts = contactListValue
				} else {
					resultItem.Contacts = types.ListNull(datasource_table.ContactsType{}.ValueType(ctx).Type(ctx))
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
					tagList := make([]datasource_table.TagsValue, 0, len(tags))
					for _, tag := range tags {
						if tagMap, ok := tag.(map[string]interface{}); ok {
							tagItem := datasource_table.TagsValue{
								TagId: types.StringValue(getStringFromMap(tagMap, "tagId")),
								Name:  types.StringValue(getStringFromMap(tagMap, "name")),
							}
							tagList = append(tagList, tagItem)
						}
					}
					tagListValue, _ := types.ListValueFrom(ctx, datasource_table.TagsType{}.ValueType(ctx).Type(ctx), tagList)
					resultItem.Tags = tagListValue
				} else {
					resultItem.Tags = types.ListNull(datasource_table.TagsType{}.ValueType(ctx).Type(ctx))
				}

				resultList = append(resultList, resultItem)
			}
		}
		resultListValue, _ := types.ListValueFrom(ctx, datasource_table.ResultType{}.ValueType(ctx).Type(ctx), resultList)
		model.Result = resultListValue
	} else {
		model.Result = types.ListNull(datasource_table.ResultType{}.ValueType(ctx).Type(ctx))
	}
}
