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
				resultValue := d.mapSingleResult(ctx, item)
				resultElements = append(resultElements, resultValue)
			}
		}
		config.Result, _ = types.ListValue(elementType, resultElements)
	} else {
		config.Result, _ = types.ListValue(elementType, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *schemasDataSource) mapSingleResult(ctx context.Context, item map[string]interface{}) datasource_schemas.ResultValue {
	result := datasource_schemas.ResultValue{}

	if v, ok := item["schemaId"].(string); ok {
		result.SchemaId = types.StringValue(v)
	} else {
		result.SchemaId = types.StringNull()
	}
	if v, ok := item["description"].(string); ok {
		result.Description = types.StringValue(v)
	} else {
		result.Description = types.StringNull()
	}

	// Map owner
	if ownerData, ok := item["owner"].(map[string]interface{}); ok {
		ownerValue := datasource_schemas.OwnerValue{
			RoleId:   types.StringNull(),
			RoleName: types.StringNull(),
		}
		if v, ok := ownerData["roleId"].(string); ok {
			ownerValue.RoleId = types.StringValue(v)
		}
		if v, ok := ownerData["roleName"].(string); ok {
			ownerValue.RoleName = types.StringValue(v)
		}
		ownerTyped, d2 := datasource_schemas.NewOwnerValue(ownerValue.AttributeTypes(ctx), map[string]attr.Value{
			"role_id":   ownerValue.RoleId,
			"role_name": ownerValue.RoleName,
		})
		if !d2.HasError() {
			ownerObj, d3 := ownerTyped.ToObjectValue(ctx)
			if !d3.HasError() {
				result.Owner = ownerObj
			}
		}
	} else {
		ownerNull := datasource_schemas.NewOwnerValueNull()
		ownerObj, _ := ownerNull.ToObjectValue(ctx)
		result.Owner = ownerObj
	}

	// Map contacts
	contactsElementType := datasource_schemas.ContactsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_schemas.ContactsValue{}.AttributeTypes(ctx),
		},
	}
	if contacts, ok := item["contacts"].([]interface{}); ok && len(contacts) > 0 {
		var contactElements []attr.Value
		for _, contactRaw := range contacts {
			if contactMap, ok := contactRaw.(map[string]interface{}); ok {
				contactValue := datasource_schemas.ContactsValue{}
				if v, ok := contactMap["email"].(string); ok {
					contactValue.Email = types.StringValue(v)
				} else {
					contactValue.Email = types.StringNull()
				}
				if v, ok := contactMap["userId"].(string); ok {
					contactValue.UserId = types.StringValue(v)
				} else {
					contactValue.UserId = types.StringNull()
				}
				contactElements = append(contactElements, contactValue)
			}
		}
		result.Contacts, _ = types.ListValue(contactsElementType, contactElements)
	} else {
		result.Contacts, _ = types.ListValue(contactsElementType, []attr.Value{})
	}

	// Map links
	linksElementType := datasource_schemas.LinksType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_schemas.LinksValue{}.AttributeTypes(ctx),
		},
	}
	if links, ok := item["links"].([]interface{}); ok && len(links) > 0 {
		var linkElements []attr.Value
		for _, linkRaw := range links {
			if linkMap, ok := linkRaw.(map[string]interface{}); ok {
				linkValue := datasource_schemas.LinksValue{}
				if v, ok := linkMap["name"].(string); ok {
					linkValue.Name = types.StringValue(v)
				} else {
					linkValue.Name = types.StringNull()
				}
				if v, ok := linkMap["uri"].(string); ok {
					linkValue.Uri = types.StringValue(v)
				} else {
					linkValue.Uri = types.StringNull()
				}
				linkElements = append(linkElements, linkValue)
			}
		}
		result.Links, _ = types.ListValue(linksElementType, linkElements)
	} else {
		result.Links, _ = types.ListValue(linksElementType, []attr.Value{})
	}

	// Map tags
	tagsElementType := datasource_schemas.TagsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_schemas.TagsValue{}.AttributeTypes(ctx),
		},
	}
	if tags, ok := item["tags"].([]interface{}); ok && len(tags) > 0 {
		var tagElements []attr.Value
		for _, tagRaw := range tags {
			if tagMap, ok := tagRaw.(map[string]interface{}); ok {
				tagValue := datasource_schemas.TagsValue{}
				if v, ok := tagMap["name"].(string); ok {
					tagValue.Name = types.StringValue(v)
				} else {
					tagValue.Name = types.StringNull()
				}
				if v, ok := tagMap["tagId"].(string); ok {
					tagValue.TagId = types.StringValue(v)
				} else {
					tagValue.TagId = types.StringNull()
				}
				tagElements = append(tagElements, tagValue)
			}
		}
		result.Tags, _ = types.ListValue(tagsElementType, tagElements)
	} else {
		result.Tags, _ = types.ListValue(tagsElementType, []attr.Value{})
	}

	return result
}
