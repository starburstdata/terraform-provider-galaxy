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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_data_product"
)

var _ datasource.DataSource = (*data_productDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*data_productDataSource)(nil)

func NewDataProductDataSource() datasource.DataSource {
	return &data_productDataSource{}
}

type data_productDataSource struct {
	client *client.GalaxyClient
}

func (d *data_productDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_product"
}

func (d *data_productDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_data_product.DataProductDataSourceSchema(ctx)
}

func (d *data_productDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *data_productDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_data_product.DataProductModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.DataProductId.ValueString()
	tflog.Debug(ctx, "Reading data_product", map[string]interface{}{"id": id})

	response, err := d.client.GetDataProduct(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading data_product",
			"Could not read data_product "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *data_productDataSource) updateModelFromResponse(ctx context.Context, model *datasource_data_product.DataProductModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
	if dataProductId, ok := response["dataProductId"].(string); ok {
		model.DataProductId = types.StringValue(dataProductId)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	}
	if summary, ok := response["summary"].(string); ok {
		model.Summary = types.StringValue(summary)
	}
	if createdOn, ok := response["createdOn"].(string); ok {
		model.CreatedOn = types.StringValue(createdOn)
	}
	if modifiedOn, ok := response["modifiedOn"].(string); ok {
		model.ModifiedOn = types.StringValue(modifiedOn)
	}
	if defaultClusterId, ok := response["defaultClusterId"].(string); ok {
		model.DefaultClusterId = types.StringValue(defaultClusterId)
	}
	if businessContext, ok := response["businessContext"].(string); ok {
		model.BusinessContext = types.StringValue(businessContext)
	} else {
		model.BusinessContext = types.StringNull()
	}

	// Map nested objects - simplified mapping for now
	if catalog, ok := response["catalog"].(map[string]interface{}); ok {
		catalogAttrTypes := datasource_data_product.CatalogValue{}.AttributeTypes(ctx)
		catalogAttrs := map[string]attr.Value{}
		if catalogId, ok := catalog["id"].(string); ok {
			catalogAttrs["catalog_id"] = types.StringValue(catalogId)
		} else if catalogId, ok := catalog["catalogId"].(string); ok {
			catalogAttrs["catalog_id"] = types.StringValue(catalogId)
		} else {
			catalogAttrs["catalog_id"] = types.StringNull()
		}
		if catalogName, ok := catalog["name"].(string); ok {
			catalogAttrs["catalog_name"] = types.StringValue(catalogName)
		} else if catalogName, ok := catalog["catalogName"].(string); ok {
			catalogAttrs["catalog_name"] = types.StringValue(catalogName)
		} else {
			catalogAttrs["catalog_name"] = types.StringNull()
		}
		if catalogKind, ok := catalog["catalogKind"].(string); ok {
			catalogAttrs["catalog_kind"] = types.StringValue(catalogKind)
		} else {
			catalogAttrs["catalog_kind"] = types.StringNull()
		}
		// Handle local_regions
		if localRegions, ok := catalog["localRegions"].([]interface{}); ok {
			regionList := make([]types.String, 0, len(localRegions))
			for _, region := range localRegions {
				if regionStr, ok := region.(string); ok {
					regionList = append(regionList, types.StringValue(regionStr))
				}
			}
			catalogAttrs["local_regions"], _ = types.ListValueFrom(ctx, types.StringType, regionList)
		} else {
			catalogAttrs["local_regions"] = types.ListNull(types.StringType)
		}
		catalogValue, d := datasource_data_product.NewCatalogValue(catalogAttrTypes, catalogAttrs)
		diags.Append(d...)
		if !d.HasError() {
			model.Catalog = catalogValue
		}
	}

	// Map contacts from response
	if contacts, ok := response["contacts"].([]interface{}); ok && len(contacts) > 0 {
		contactsList := make([]datasource_data_product.ContactsValue, 0, len(contacts))
		for _, contactInterface := range contacts {
			if contactMap, ok := contactInterface.(map[string]interface{}); ok {
				attrTypes := datasource_data_product.ContactsValue{}.AttributeTypes(ctx)
				attrs := map[string]attr.Value{}
				if email, ok := contactMap["email"].(string); ok {
					attrs["email"] = types.StringValue(email)
				} else {
					attrs["email"] = types.StringNull()
				}
				if userId, ok := contactMap["userId"].(string); ok {
					attrs["user_id"] = types.StringValue(userId)
				} else {
					attrs["user_id"] = types.StringNull()
				}
				contactValue, d := datasource_data_product.NewContactsValue(attrTypes, attrs)
				diags.Append(d...)
				if !d.HasError() {
					contactsList = append(contactsList, contactValue)
				}
			}
		}
		contactsListValue, d := types.ListValueFrom(ctx, datasource_data_product.ContactsValue{}.Type(ctx), contactsList)
		diags.Append(d...)
		if !d.HasError() {
			model.Contacts = contactsListValue
		}
	} else {
		contactsListValue, d := types.ListValueFrom(ctx, datasource_data_product.ContactsValue{}.Type(ctx), []datasource_data_product.ContactsValue{})
		diags.Append(d...)
		if !d.HasError() {
			model.Contacts = contactsListValue
		}
	}

	// Map links from response
	if links, ok := response["links"].([]interface{}); ok && len(links) > 0 {
		linksList := make([]datasource_data_product.LinksValue, 0, len(links))
		for _, linkInterface := range links {
			if linkMap, ok := linkInterface.(map[string]interface{}); ok {
				attrTypes := datasource_data_product.LinksValue{}.AttributeTypes(ctx)
				attrs := map[string]attr.Value{}
				if name, ok := linkMap["name"].(string); ok {
					attrs["name"] = types.StringValue(name)
				} else {
					attrs["name"] = types.StringNull()
				}
				if uri, ok := linkMap["uri"].(string); ok {
					attrs["uri"] = types.StringValue(uri)
				} else {
					attrs["uri"] = types.StringNull()
				}
				linkValue, d := datasource_data_product.NewLinksValue(attrTypes, attrs)
				diags.Append(d...)
				if !d.HasError() {
					linksList = append(linksList, linkValue)
				}
			}
		}
		linksListValue, d := types.ListValueFrom(ctx, datasource_data_product.LinksValue{}.Type(ctx), linksList)
		diags.Append(d...)
		if !d.HasError() {
			model.Links = linksListValue
		}
	} else {
		linksListValue, d := types.ListValueFrom(ctx, datasource_data_product.LinksValue{}.Type(ctx), []datasource_data_product.LinksValue{})
		diags.Append(d...)
		if !d.HasError() {
			model.Links = linksListValue
		}
	}

	// Map created by and modified by
	if createdBy, ok := response["createdBy"].(map[string]interface{}); ok {
		createdByAttrTypes := datasource_data_product.CreatedByValue{}.AttributeTypes(ctx)
		createdByAttrs := map[string]attr.Value{}
		if email, ok := createdBy["email"].(string); ok {
			createdByAttrs["email"] = types.StringValue(email)
		} else {
			createdByAttrs["email"] = types.StringNull()
		}
		if userId, ok := createdBy["userId"].(string); ok {
			createdByAttrs["user_id"] = types.StringValue(userId)
		} else if id, ok := createdBy["id"].(string); ok {
			createdByAttrs["user_id"] = types.StringValue(id)
		} else {
			createdByAttrs["user_id"] = types.StringNull()
		}
		createdByValue, d := datasource_data_product.NewCreatedByValue(createdByAttrTypes, createdByAttrs)
		diags.Append(d...)
		if !d.HasError() {
			model.CreatedBy = createdByValue
		}
	}

	if modifiedBy, ok := response["modifiedBy"].(map[string]interface{}); ok {
		modifiedByAttrTypes := datasource_data_product.ModifiedByValue{}.AttributeTypes(ctx)
		modifiedByAttrs := map[string]attr.Value{}
		if email, ok := modifiedBy["email"].(string); ok {
			modifiedByAttrs["email"] = types.StringValue(email)
		} else {
			modifiedByAttrs["email"] = types.StringNull()
		}
		if userId, ok := modifiedBy["userId"].(string); ok {
			modifiedByAttrs["user_id"] = types.StringValue(userId)
		} else if id, ok := modifiedBy["id"].(string); ok {
			modifiedByAttrs["user_id"] = types.StringValue(id)
		} else {
			modifiedByAttrs["user_id"] = types.StringNull()
		}
		modifiedByValue, d := datasource_data_product.NewModifiedByValue(modifiedByAttrTypes, modifiedByAttrs)
		diags.Append(d...)
		if !d.HasError() {
			model.ModifiedBy = modifiedByValue
		}
	}
}
