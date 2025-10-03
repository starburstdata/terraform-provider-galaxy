package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_catalog_metadata"
)

var _ datasource.DataSource = (*catalog_metadataDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*catalog_metadataDataSource)(nil)

func NewCatalogMetadataDataSource() datasource.DataSource {
	return &catalog_metadataDataSource{}
}

type catalog_metadataDataSource struct {
	client *client.GalaxyClient
}

func (d *catalog_metadataDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_catalog_metadata"
}

func (d *catalog_metadataDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_catalog_metadata.CatalogMetadataDataSourceSchema(ctx)
}

func (d *catalog_metadataDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *catalog_metadataDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_catalog_metadata.CatalogMetadataModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading catalog_metadata", map[string]interface{}{"id": id})

	response, err := d.client.GetCatalogMetadata(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading catalog_metadata",
			"Could not read catalog_metadata "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *catalog_metadataDataSource) updateModelFromResponse(ctx context.Context, model *datasource_catalog_metadata.CatalogMetadataModel, response map[string]interface{}) {
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	}
	if catalogId, ok := response["catalogId"].(string); ok {
		model.CatalogId = types.StringValue(catalogId)
	}
	if catalogName, ok := response["catalogName"].(string); ok {
		model.CatalogName = types.StringValue(catalogName)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	}

	// Map tags to the Tags field in the model
	if tags, ok := response["tags"].([]interface{}); ok {
		if len(tags) > 0 {
			// Convert tags to the expected structure
			var tagElements []attr.Value
			for _, tag := range tags {
				if tagMap, ok := tag.(map[string]interface{}); ok {
					tagValue := datasource_catalog_metadata.TagsValue{
						Name:  types.StringValue(""),
						TagId: types.StringValue(""),
					}
					if name, ok := tagMap["name"].(string); ok {
						tagValue.Name = types.StringValue(name)
					}
					if tagId, ok := tagMap["tag_id"].(string); ok {
						tagValue.TagId = types.StringValue(tagId)
					}
					tagElements = append(tagElements, tagValue)
				}
			}
			elementType := datasource_catalog_metadata.TagsType{
				ObjectType: types.ObjectType{
					AttrTypes: datasource_catalog_metadata.TagsValue{}.AttributeTypes(ctx),
				},
			}
			model.Tags, _ = types.ListValue(elementType, tagElements)
		} else {
			// Empty tags list with proper type
			elementType := datasource_catalog_metadata.TagsType{
				ObjectType: types.ObjectType{
					AttrTypes: datasource_catalog_metadata.TagsValue{}.AttributeTypes(ctx),
				},
			}
			model.Tags, _ = types.ListValue(elementType, []attr.Value{})
		}
	} else {
		elementType := datasource_catalog_metadata.TagsType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_catalog_metadata.TagsValue{}.AttributeTypes(ctx),
			},
		}
		model.Tags, _ = types.ListValue(elementType, []attr.Value{})
	}

	// Map contacts to the Contacts field in the model
	if contacts, ok := response["contacts"].([]interface{}); ok {
		if len(contacts) > 0 {
			// Convert contacts to the expected structure
			var contactElements []attr.Value
			for _, contact := range contacts {
				if contactMap, ok := contact.(map[string]interface{}); ok {
					contactValue := datasource_catalog_metadata.ContactsValue{
						Email:  types.StringValue(""),
						UserId: types.StringValue(""),
					}
					if email, ok := contactMap["email"].(string); ok {
						contactValue.Email = types.StringValue(email)
					}
					if userId, ok := contactMap["user_id"].(string); ok {
						contactValue.UserId = types.StringValue(userId)
					}
					contactElements = append(contactElements, contactValue)
				}
			}
			elementType := datasource_catalog_metadata.ContactsType{
				ObjectType: types.ObjectType{
					AttrTypes: datasource_catalog_metadata.ContactsValue{}.AttributeTypes(ctx),
				},
			}
			model.Contacts, _ = types.ListValue(elementType, contactElements)
		} else {
			// Empty contacts list with proper type
			elementType := datasource_catalog_metadata.ContactsType{
				ObjectType: types.ObjectType{
					AttrTypes: datasource_catalog_metadata.ContactsValue{}.AttributeTypes(ctx),
				},
			}
			model.Contacts, _ = types.ListValue(elementType, []attr.Value{})
		}
	} else {
		elementType := datasource_catalog_metadata.ContactsType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_catalog_metadata.ContactsValue{}.AttributeTypes(ctx),
			},
		}
		model.Contacts, _ = types.ListValue(elementType, []attr.Value{})
	}

	// Map owner role
	if ownerData, ok := response["owner"].(map[string]interface{}); ok {
		ownerValue := datasource_catalog_metadata.OwnerValue{
			RoleId:   types.StringValue(""),
			RoleName: types.StringValue(""),
		}
		if roleId, ok := ownerData["roleId"].(string); ok {
			ownerValue.RoleId = types.StringValue(roleId)
		}
		if roleName, ok := ownerData["roleName"].(string); ok {
			ownerValue.RoleName = types.StringValue(roleName)
		}
		var diags diag.Diagnostics
		model.Owner, diags = datasource_catalog_metadata.NewOwnerValue(
			ownerValue.AttributeTypes(ctx),
			map[string]attr.Value{
				"role_id":   ownerValue.RoleId,
				"role_name": ownerValue.RoleName,
			},
		)
		_ = diags // Ignore diagnostics for now
	} else {
		model.Owner = datasource_catalog_metadata.NewOwnerValueNull()
	}
}
