// Copyright Starburst Data, Inc. All rights reserved.
//
// The source code is the proprietary and confidential information of Starburst Data, Inc. and
// may be used only for reference purposes in connection with the Terraform Registry. All rights,
// title, interest and ownership of the code and any derivatives, updates, upgrades, enhancements
// and modifications thereof remain with Starburst Data, Inc. You are not permitted to distribute,
// disclose, sell, lease, transfer, assign, modify, create derivative works of, or sublicense the
// code, or use the code to create or develop any products or services.

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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_schemas"
)

var _ datasource.DataSource = (*schemasDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*schemasDataSource)(nil)

func NewSchemasDataSource() datasource.DataSource {
	return &schemasDataSource{}
}

type schemasDataSource struct {
	client *client.GalaxyClient
}

func (d *schemasDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schemas"
}

func (d *schemasDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_schemas.SchemasDataSourceSchema(ctx)
}

func (d *schemasDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *schemasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_schemas.SchemasModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := config.CatalogId.ValueString()
	tflog.Debug(ctx, "Reading schemas", map[string]interface{}{"catalog_id": catalogID})

	response, err := d.client.ListSchemas(ctx, catalogID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading schemas",
			"Could not read schemas for catalog "+catalogID+": "+err.Error(),
		)
		return
	}

	// Map results
	elementType := datasource_schemas.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_schemas.ResultValue{}.AttributeTypes(ctx),
		},
	}

	var results []interface{}
	if items, ok := response["items"].([]interface{}); ok {
		results = items
	} else if arr, ok := response["result"].([]interface{}); ok {
		results = arr
	}

	if len(results) > 0 {
		var resultElements []attr.Value
		for _, itemRaw := range results {
			if item, ok := itemRaw.(map[string]interface{}); ok {
				resultValue, diags := d.mapSingleResult(ctx, item)
				resp.Diagnostics.Append(diags...)
				if diags.HasError() {
					continue
				}
				resultElements = append(resultElements, resultValue)
			}
		}
		var ld diag.Diagnostics
		config.Result, ld = types.ListValue(elementType, resultElements)
		resp.Diagnostics.Append(ld...)
	} else {
		var ld diag.Diagnostics
		config.Result, ld = types.ListValue(elementType, []attr.Value{})
		resp.Diagnostics.Append(ld...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *schemasDataSource) mapSingleResult(ctx context.Context, item map[string]interface{}) (datasource_schemas.ResultValue, diag.Diagnostics) {
	attributeTypes := datasource_schemas.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	if v, ok := item["schemaId"].(string); ok {
		attributes["schema_id"] = types.StringValue(v)
	} else {
		attributes["schema_id"] = types.StringNull()
	}
	if v, ok := item["description"].(string); ok {
		attributes["description"] = types.StringValue(v)
	} else {
		attributes["description"] = types.StringNull()
	}

	// Map owner
	ownerAttributeTypes := datasource_schemas.OwnerValue{}.AttributeTypes(ctx)
	if ownerData, ok := item["owner"].(map[string]interface{}); ok {
		ownerAttributes := map[string]attr.Value{}
		if v, ok := ownerData["roleId"].(string); ok {
			ownerAttributes["role_id"] = types.StringValue(v)
		} else {
			ownerAttributes["role_id"] = types.StringNull()
		}
		if v, ok := ownerData["roleName"].(string); ok {
			ownerAttributes["role_name"] = types.StringValue(v)
		} else {
			ownerAttributes["role_name"] = types.StringNull()
		}
		ownerTyped, diags := datasource_schemas.NewOwnerValue(ownerAttributeTypes, ownerAttributes)
		if diags.HasError() {
			return datasource_schemas.ResultValue{}, diags
		}
		ownerObj, diags := ownerTyped.ToObjectValue(ctx)
		if diags.HasError() {
			return datasource_schemas.ResultValue{}, diags
		}
		attributes["owner"] = ownerObj
	} else {
		attributes["owner"] = types.ObjectNull(ownerAttributeTypes)
	}

	// Map contacts
	contactsElementType := datasource_schemas.ContactsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_schemas.ContactsValue{}.AttributeTypes(ctx),
		},
	}
	contactsAttributeTypes := datasource_schemas.ContactsValue{}.AttributeTypes(ctx)
	if contacts, ok := item["contacts"].([]interface{}); ok && len(contacts) > 0 {
		var contactElements []attr.Value
		for _, contactRaw := range contacts {
			if contactMap, ok := contactRaw.(map[string]interface{}); ok {
				contactAttributes := map[string]attr.Value{}
				if v, ok := contactMap["email"].(string); ok {
					contactAttributes["email"] = types.StringValue(v)
				} else {
					contactAttributes["email"] = types.StringNull()
				}
				if v, ok := contactMap["userId"].(string); ok {
					contactAttributes["user_id"] = types.StringValue(v)
				} else {
					contactAttributes["user_id"] = types.StringNull()
				}
				contactValue, diags := datasource_schemas.NewContactsValue(contactsAttributeTypes, contactAttributes)
				if diags.HasError() {
					return datasource_schemas.ResultValue{}, diags
				}
				contactElements = append(contactElements, contactValue)
			}
		}
		attributes["contacts"], _ = types.ListValue(contactsElementType, contactElements)
	} else {
		attributes["contacts"], _ = types.ListValue(contactsElementType, []attr.Value{})
	}

	// Map links
	linksElementType := datasource_schemas.LinksType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_schemas.LinksValue{}.AttributeTypes(ctx),
		},
	}
	linksAttributeTypes := datasource_schemas.LinksValue{}.AttributeTypes(ctx)
	if links, ok := item["links"].([]interface{}); ok && len(links) > 0 {
		var linkElements []attr.Value
		for _, linkRaw := range links {
			if linkMap, ok := linkRaw.(map[string]interface{}); ok {
				linkAttributes := map[string]attr.Value{}
				if v, ok := linkMap["name"].(string); ok {
					linkAttributes["name"] = types.StringValue(v)
				} else {
					linkAttributes["name"] = types.StringNull()
				}
				if v, ok := linkMap["uri"].(string); ok {
					linkAttributes["uri"] = types.StringValue(v)
				} else {
					linkAttributes["uri"] = types.StringNull()
				}
				linkValue, diags := datasource_schemas.NewLinksValue(linksAttributeTypes, linkAttributes)
				if diags.HasError() {
					return datasource_schemas.ResultValue{}, diags
				}
				linkElements = append(linkElements, linkValue)
			}
		}
		attributes["links"], _ = types.ListValue(linksElementType, linkElements)
	} else {
		attributes["links"], _ = types.ListValue(linksElementType, []attr.Value{})
	}

	// Map tags
	tagsElementType := datasource_schemas.TagsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_schemas.TagsValue{}.AttributeTypes(ctx),
		},
	}
	tagsAttributeTypes := datasource_schemas.TagsValue{}.AttributeTypes(ctx)
	if tags, ok := item["tags"].([]interface{}); ok && len(tags) > 0 {
		var tagElements []attr.Value
		for _, tagRaw := range tags {
			if tagMap, ok := tagRaw.(map[string]interface{}); ok {
				tagAttributes := map[string]attr.Value{}
				if v, ok := tagMap["name"].(string); ok {
					tagAttributes["name"] = types.StringValue(v)
				} else {
					tagAttributes["name"] = types.StringNull()
				}
				if v, ok := tagMap["tagId"].(string); ok {
					tagAttributes["tag_id"] = types.StringValue(v)
				} else {
					tagAttributes["tag_id"] = types.StringNull()
				}
				tagValue, diags := datasource_schemas.NewTagsValue(tagsAttributeTypes, tagAttributes)
				if diags.HasError() {
					return datasource_schemas.ResultValue{}, diags
				}
				tagElements = append(tagElements, tagValue)
			}
		}
		attributes["tags"], _ = types.ListValue(tagsElementType, tagElements)
	} else {
		attributes["tags"], _ = types.ListValue(tagsElementType, []attr.Value{})
	}

	return datasource_schemas.NewResultValue(attributeTypes, attributes)
}
