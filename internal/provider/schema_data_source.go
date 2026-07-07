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

	catalogID := config.CatalogId.ValueString()
	schemaID := config.SchemaId.ValueString()
	tflog.Debug(ctx, "Reading schema", map[string]interface{}{
		"catalog_id": catalogID,
		"schema_id":  schemaID,
	})

	schemaData, err := d.client.GetSchema(ctx, catalogID, schemaID)
	if err != nil {
		if client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Schema not found",
				fmt.Sprintf("Could not find schema with ID %s in catalog %s", schemaID, catalogID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading schema",
			"Could not read schema: "+err.Error(),
		)
		return
	}

	d.updateModelFromResponse(ctx, &config, schemaData)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *schemaDataSource) updateModelFromResponse(ctx context.Context, model *datasource_schema.SchemaModel, response map[string]interface{}) {
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else {
		model.Description = types.StringNull()
	}

	// Map contacts list
	contactElementType := datasource_schema.ContactsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_schema.ContactsValue{}.AttributeTypes(ctx),
		},
	}
	if contacts, ok := response["contacts"].([]interface{}); ok && len(contacts) > 0 {
		contactList := make([]datasource_schema.ContactsValue, 0, len(contacts))
		for _, contact := range contacts {
			if contactMap, ok := contact.(map[string]interface{}); ok {
				contactAttrs := map[string]attr.Value{
					"user_id": types.StringNull(),
					"email":   types.StringNull(),
				}
				if userId, ok := contactMap["userId"].(string); ok {
					contactAttrs["user_id"] = types.StringValue(userId)
				}
				if email, ok := contactMap["email"].(string); ok {
					contactAttrs["email"] = types.StringValue(email)
				}
				contactValue, diags := datasource_schema.NewContactsValue(
					datasource_schema.ContactsValue{}.AttributeTypes(ctx),
					contactAttrs,
				)
				if diags.HasError() {
					tflog.Error(ctx, fmt.Sprintf("Error creating contact value: %v", diags))
					continue
				}
				contactList = append(contactList, contactValue)
			}
		}
		listValue, diags := types.ListValueFrom(ctx, contactElementType, contactList)
		if !diags.HasError() {
			model.Contacts = listValue
		} else {
			model.Contacts, _ = types.ListValueFrom(ctx, contactElementType, []datasource_schema.ContactsValue{})
		}
	} else {
		model.Contacts, _ = types.ListValueFrom(ctx, contactElementType, []datasource_schema.ContactsValue{})
	}

	// Map links list
	linkElementType := datasource_schema.LinksType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_schema.LinksValue{}.AttributeTypes(ctx),
		},
	}
	if links, ok := response["links"].([]interface{}); ok && len(links) > 0 {
		linkList := make([]datasource_schema.LinksValue, 0, len(links))
		for _, link := range links {
			if linkMap, ok := link.(map[string]interface{}); ok {
				linkAttrs := map[string]attr.Value{
					"name": types.StringNull(),
					"uri":  types.StringNull(),
				}
				if name, ok := linkMap["name"].(string); ok {
					linkAttrs["name"] = types.StringValue(name)
				}
				if uri, ok := linkMap["uri"].(string); ok {
					linkAttrs["uri"] = types.StringValue(uri)
				}
				linkValue, diags := datasource_schema.NewLinksValue(
					datasource_schema.LinksValue{}.AttributeTypes(ctx),
					linkAttrs,
				)
				if diags.HasError() {
					tflog.Error(ctx, fmt.Sprintf("Error creating link value: %v", diags))
					continue
				}
				linkList = append(linkList, linkValue)
			}
		}
		listValue, diags := types.ListValueFrom(ctx, linkElementType, linkList)
		if !diags.HasError() {
			model.Links = listValue
		} else {
			model.Links, _ = types.ListValueFrom(ctx, linkElementType, []datasource_schema.LinksValue{})
		}
	} else {
		model.Links, _ = types.ListValueFrom(ctx, linkElementType, []datasource_schema.LinksValue{})
	}

	// Map owner
	if owner, ok := response["owner"].(map[string]interface{}); ok {
		ownerAttrs := map[string]attr.Value{
			"role_id":   types.StringNull(),
			"role_name": types.StringNull(),
		}
		if roleId, ok := owner["roleId"].(string); ok {
			ownerAttrs["role_id"] = types.StringValue(roleId)
		}
		if roleName, ok := owner["roleName"].(string); ok {
			ownerAttrs["role_name"] = types.StringValue(roleName)
		}
		ownerValue, diags := datasource_schema.NewOwnerValue(
			datasource_schema.OwnerValue{}.AttributeTypes(ctx),
			ownerAttrs,
		)
		if !diags.HasError() {
			model.Owner = ownerValue
		} else {
			model.Owner = datasource_schema.NewOwnerValueNull()
		}
	} else {
		model.Owner = datasource_schema.NewOwnerValueNull()
	}

	// Map tags list
	tagElementType := datasource_schema.TagsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_schema.TagsValue{}.AttributeTypes(ctx),
		},
	}
	if tags, ok := response["tags"].([]interface{}); ok && len(tags) > 0 {
		tagList := make([]datasource_schema.TagsValue, 0, len(tags))
		for _, tag := range tags {
			if tagMap, ok := tag.(map[string]interface{}); ok {
				tagAttrs := map[string]attr.Value{
					"tag_id": types.StringNull(),
					"name":   types.StringNull(),
				}
				if tagId, ok := tagMap["tagId"].(string); ok {
					tagAttrs["tag_id"] = types.StringValue(tagId)
				}
				if name, ok := tagMap["name"].(string); ok {
					tagAttrs["name"] = types.StringValue(name)
				}
				tagValue, diags := datasource_schema.NewTagsValue(
					datasource_schema.TagsValue{}.AttributeTypes(ctx),
					tagAttrs,
				)
				if diags.HasError() {
					tflog.Error(ctx, fmt.Sprintf("Error creating tag value: %v", diags))
					continue
				}
				tagList = append(tagList, tagValue)
			}
		}
		listValue, diags := types.ListValueFrom(ctx, tagElementType, tagList)
		if !diags.HasError() {
			model.Tags = listValue
		} else {
			model.Tags, _ = types.ListValueFrom(ctx, tagElementType, []datasource_schema.TagsValue{})
		}
	} else {
		model.Tags, _ = types.ListValueFrom(ctx, tagElementType, []datasource_schema.TagsValue{})
	}
}
