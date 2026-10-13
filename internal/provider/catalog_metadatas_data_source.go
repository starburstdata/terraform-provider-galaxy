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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_catalog_metadatas"
)

var _ datasource.DataSource = (*catalogMetadatasDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*catalogMetadatasDataSource)(nil)

func NewCatalogMetadatasDataSource() datasource.DataSource {
	return &catalogMetadatasDataSource{}
}

type catalogMetadatasDataSource struct {
	client *client.GalaxyClient
}

func (d *catalogMetadatasDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_catalog_metadatas"
}

func (d *catalogMetadatasDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_catalog_metadatas.CatalogMetadatasDataSourceSchema(ctx)
}

func (d *catalogMetadatasDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *catalogMetadatasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_catalog_metadatas.CatalogMetadatasModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := config.CatalogId.ValueString()
	tflog.Debug(ctx, "Reading catalog_metadatas", map[string]interface{}{"catalog_id": catalogID})

	response, err := d.client.GetCatalogMetadata(ctx, catalogID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading catalog_metadatas",
			"Could not read catalog metadata for "+catalogID+": "+err.Error(),
		)
		return
	}

	diags := d.updateModelFromResponse(ctx, &config, response)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *catalogMetadatasDataSource) updateModelFromResponse(ctx context.Context, model *datasource_catalog_metadatas.CatalogMetadatasModel, response map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if v, ok := response["catalogId"].(string); ok {
		model.CatalogId = types.StringValue(v)
	}
	if v, ok := response["catalogName"].(string); ok {
		model.CatalogName = types.StringValue(v)
	} else {
		model.CatalogName = types.StringNull()
	}
	if v, ok := response["description"].(string); ok {
		model.Description = types.StringValue(v)
	} else {
		model.Description = types.StringNull()
	}

	// Map owner
	if ownerData, ok := response["owner"].(map[string]interface{}); ok {
		ownerValue := datasource_catalog_metadatas.OwnerValue{
			RoleId:   types.StringNull(),
			RoleName: types.StringNull(),
		}
		if v, ok := ownerData["roleId"].(string); ok {
			ownerValue.RoleId = types.StringValue(v)
		}
		if v, ok := ownerData["roleName"].(string); ok {
			ownerValue.RoleName = types.StringValue(v)
		}
		ownerObj, d2 := datasource_catalog_metadatas.NewOwnerValue(ownerValue.AttributeTypes(ctx), map[string]attr.Value{
			"role_id":   ownerValue.RoleId,
			"role_name": ownerValue.RoleName,
		})
		diags.Append(d2...)
		if !d2.HasError() {
			model.Owner = ownerObj
		}
	} else {
		model.Owner = datasource_catalog_metadatas.NewOwnerValueNull()
	}

	// Map contacts
	contactsElementType := datasource_catalog_metadatas.ContactsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_catalog_metadatas.ContactsValue{}.AttributeTypes(ctx),
		},
	}
	contactsAttrTypes := datasource_catalog_metadatas.ContactsValue{}.AttributeTypes(ctx)
	if contacts, ok := response["contacts"].([]interface{}); ok && len(contacts) > 0 {
		var contactElements []attr.Value
		for _, contactRaw := range contacts {
			if contactMap, ok := contactRaw.(map[string]interface{}); ok {
				contactAttrs := map[string]attr.Value{}
				if v, ok := contactMap["email"].(string); ok {
					contactAttrs["email"] = types.StringValue(v)
				} else {
					contactAttrs["email"] = types.StringNull()
				}
				if v, ok := contactMap["userId"].(string); ok {
					contactAttrs["user_id"] = types.StringValue(v)
				} else {
					contactAttrs["user_id"] = types.StringNull()
				}
				contactValue, d := datasource_catalog_metadatas.NewContactsValue(contactsAttrTypes, contactAttrs)
				diags.Append(d...)
				if !d.HasError() {
					contactElements = append(contactElements, contactValue)
				}
			}
		}
		var ld diag.Diagnostics
		model.Contacts, ld = types.ListValue(contactsElementType, contactElements)
		diags.Append(ld...)
	} else {
		var ld diag.Diagnostics
		model.Contacts, ld = types.ListValue(contactsElementType, []attr.Value{})
		diags.Append(ld...)
	}

	// Map tags
	tagsElementType := datasource_catalog_metadatas.TagsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_catalog_metadatas.TagsValue{}.AttributeTypes(ctx),
		},
	}
	tagsAttrTypes := datasource_catalog_metadatas.TagsValue{}.AttributeTypes(ctx)
	if tags, ok := response["tags"].([]interface{}); ok && len(tags) > 0 {
		var tagElements []attr.Value
		for _, tagRaw := range tags {
			if tagMap, ok := tagRaw.(map[string]interface{}); ok {
				tagAttrs := map[string]attr.Value{}
				if v, ok := tagMap["name"].(string); ok {
					tagAttrs["name"] = types.StringValue(v)
				} else {
					tagAttrs["name"] = types.StringNull()
				}
				if v, ok := tagMap["tagId"].(string); ok {
					tagAttrs["tag_id"] = types.StringValue(v)
				} else {
					tagAttrs["tag_id"] = types.StringNull()
				}
				tagValue, d := datasource_catalog_metadatas.NewTagsValue(tagsAttrTypes, tagAttrs)
				diags.Append(d...)
				if !d.HasError() {
					tagElements = append(tagElements, tagValue)
				}
			}
		}
		var ld diag.Diagnostics
		model.Tags, ld = types.ListValue(tagsElementType, tagElements)
		diags.Append(ld...)
	} else {
		var ld diag.Diagnostics
		model.Tags, ld = types.ListValue(tagsElementType, []attr.Value{})
		diags.Append(ld...)
	}

	return diags
}
