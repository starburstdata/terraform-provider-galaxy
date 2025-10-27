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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
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

	id := config.Id.ValueString()
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
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *data_productDataSource) updateModelFromResponse(ctx context.Context, model *datasource_data_product.DataProductModel, response map[string]interface{}) {
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	}
	if dataProductId, ok := response["id"].(string); ok {
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

	// Map nested objects - simplified mapping for now
	if catalog, ok := response["catalog"].(map[string]interface{}); ok {
		catalogValue := datasource_data_product.CatalogValue{}
		if catalogId, ok := catalog["id"].(string); ok {
			catalogValue.CatalogId = types.StringValue(catalogId)
		} else if catalogId, ok := catalog["catalogId"].(string); ok {
			catalogValue.CatalogId = types.StringValue(catalogId)
		}
		if catalogName, ok := catalog["name"].(string); ok {
			catalogValue.CatalogName = types.StringValue(catalogName)
		} else if catalogName, ok := catalog["catalogName"].(string); ok {
			catalogValue.CatalogName = types.StringValue(catalogName)
		}
		if catalogKind, ok := catalog["catalogKind"].(string); ok {
			catalogValue.CatalogKind = types.StringValue(catalogKind)
		} else {
			catalogValue.CatalogKind = types.StringNull()
		}
		// Handle local_regions
		if localRegions, ok := catalog["localRegions"].([]interface{}); ok {
			regionList := make([]types.String, 0, len(localRegions))
			for _, region := range localRegions {
				if regionStr, ok := region.(string); ok {
					regionList = append(regionList, types.StringValue(regionStr))
				}
			}
			catalogValue.LocalRegions, _ = types.ListValueFrom(ctx, types.StringType, regionList)
		} else {
			catalogValue.LocalRegions = types.ListNull(types.StringType)
		}
		model.Catalog = catalogValue
	}

	// Map contacts and links lists - use correct types
	model.Contacts = types.ListNull(datasource_data_product.ContactsValue{}.Type(ctx))
	model.Links = types.ListNull(datasource_data_product.LinksValue{}.Type(ctx))

	// Map created by and modified by
	if createdBy, ok := response["createdBy"].(map[string]interface{}); ok {
		createdByValue := datasource_data_product.CreatedByValue{}
		if email, ok := createdBy["email"].(string); ok {
			createdByValue.Email = types.StringValue(email)
		} else {
			createdByValue.Email = types.StringNull()
		}
		if userId, ok := createdBy["userId"].(string); ok {
			createdByValue.UserId = types.StringValue(userId)
		} else if id, ok := createdBy["id"].(string); ok {
			createdByValue.UserId = types.StringValue(id)
		} else {
			createdByValue.UserId = types.StringNull()
		}
		model.CreatedBy = createdByValue
	}

	if modifiedBy, ok := response["modifiedBy"].(map[string]interface{}); ok {
		modifiedByValue := datasource_data_product.ModifiedByValue{}
		if email, ok := modifiedBy["email"].(string); ok {
			modifiedByValue.Email = types.StringValue(email)
		} else {
			modifiedByValue.Email = types.StringNull()
		}
		if userId, ok := modifiedBy["userId"].(string); ok {
			modifiedByValue.UserId = types.StringValue(userId)
		} else if id, ok := modifiedBy["id"].(string); ok {
			modifiedByValue.UserId = types.StringValue(id)
		} else {
			modifiedByValue.UserId = types.StringNull()
		}
		model.ModifiedBy = modifiedByValue
	}
}
