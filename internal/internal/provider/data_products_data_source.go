package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_data_products"
)

var _ datasource.DataSource = (*data_productsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*data_productsDataSource)(nil)

func NewDataProductsDataSource() datasource.DataSource {
	return &data_productsDataSource{}
}

type data_productsDataSource struct {
	client *client.GalaxyClient
}

func (d *data_productsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_products"
}

func (d *data_productsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_data_products.DataProductsDataSourceSchema(ctx)
}

func (d *data_productsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *data_productsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_data_products.DataProductsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading data products with automatic pagination")

	// Use automatic pagination to get ALL data products across all pages
	allDataProducts, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/dataProduct")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading data products",
			"Could not read data products: "+err.Error(),
		)
		return
	}

	// Convert []interface{} to []map[string]interface{} for mapping
	var dataProductMaps []map[string]interface{}
	for _, dataProductInterface := range allDataProducts {
		if dataProductMap, ok := dataProductInterface.(map[string]interface{}); ok {
			dataProductMaps = append(dataProductMaps, dataProductMap)
		}
	}

	// Map API response to model
	if len(dataProductMaps) > 0 {
		dataProducts, err := d.mapDataProductsResult(ctx, dataProductMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping data products response",
				"Could not map data products response: "+err.Error(),
			)
			return
		}
		config.Result = dataProducts
	} else {
		elementType := datasource_data_products.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_data_products.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_data_products.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *data_productsDataSource) mapDataProductsResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	dataProducts := make([]datasource_data_products.ResultValue, 0)

	for _, dataProductMap := range result {
		dataProduct := d.mapSingleDataProduct(ctx, dataProductMap)
		dataProducts = append(dataProducts, dataProduct)
	}

	elementType := datasource_data_products.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_data_products.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, dataProducts)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *data_productsDataSource) mapSingleDataProduct(ctx context.Context, dataProductMap map[string]interface{}) datasource_data_products.ResultValue {
	attributeTypes := datasource_data_products.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map data product ID
	if dataProductId, ok := dataProductMap["dataProductId"].(string); ok {
		attributes["data_product_id"] = types.StringValue(dataProductId)
	} else {
		attributes["data_product_id"] = types.StringNull()
	}

	// Map name
	if name, ok := dataProductMap["name"].(string); ok {
		attributes["name"] = types.StringValue(name)
	} else {
		attributes["name"] = types.StringNull()
	}

	// Map description
	if description, ok := dataProductMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	// Map summary
	if summary, ok := dataProductMap["summary"].(string); ok {
		attributes["summary"] = types.StringValue(summary)
	} else {
		attributes["summary"] = types.StringNull()
	}

	// Map default cluster ID
	if defaultClusterId, ok := dataProductMap["defaultClusterId"].(string); ok {
		attributes["default_cluster_id"] = types.StringValue(defaultClusterId)
	} else {
		attributes["default_cluster_id"] = types.StringNull()
	}

	// Map created on
	if createdOn, ok := dataProductMap["createdOn"].(string); ok {
		attributes["created_on"] = types.StringValue(createdOn)
	} else {
		attributes["created_on"] = types.StringNull()
	}

	// Map modified on
	if modifiedOn, ok := dataProductMap["modifiedOn"].(string); ok {
		attributes["modified_on"] = types.StringValue(modifiedOn)
	} else {
		attributes["modified_on"] = types.StringNull()
	}

	// Map complex nested objects
	attributes["catalog"] = d.mapCatalogObject(ctx, dataProductMap)
	attributes["created_by"] = d.mapCreatedByObject(ctx, dataProductMap)
	attributes["modified_by"] = d.mapModifiedByObject(ctx, dataProductMap)
	attributes["contacts"] = d.mapContactsList(ctx, dataProductMap)
	attributes["links"] = d.mapLinksList(ctx, dataProductMap)

	// Create the ResultValue using the constructor
	dataProduct, diags := datasource_data_products.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating data product ResultValue: %v", diags))
		return datasource_data_products.NewResultValueNull()
	}

	return dataProduct
}

func (d *data_productsDataSource) mapCatalogObject(ctx context.Context, dataProductMap map[string]interface{}) types.Object {
	if catalogData, ok := dataProductMap["catalog"].(map[string]interface{}); ok {
		attributeTypes := datasource_data_products.CatalogValue{}.AttributeTypes(ctx)
		attributes := map[string]attr.Value{}

		if catalogId, ok := catalogData["catalogId"].(string); ok {
			attributes["catalog_id"] = types.StringValue(catalogId)
		} else {
			attributes["catalog_id"] = types.StringNull()
		}

		if name, ok := catalogData["name"].(string); ok {
			attributes["name"] = types.StringValue(name)
		} else {
			attributes["name"] = types.StringNull()
		}

		catalogValue, diags := datasource_data_products.NewCatalogValue(attributeTypes, attributes)
		if diags.HasError() {
			tflog.Error(ctx, fmt.Sprintf("Error creating catalog value: %v", diags))
			catalogObjectValue, _ := datasource_data_products.NewCatalogValueNull().ToObjectValue(ctx)
			return catalogObjectValue
		}

		catalogObjectValue, _ := catalogValue.ToObjectValue(ctx)
		return catalogObjectValue
	}

	catalogObjectValue, _ := datasource_data_products.NewCatalogValueNull().ToObjectValue(ctx)
	return catalogObjectValue
}

func (d *data_productsDataSource) mapCreatedByObject(ctx context.Context, dataProductMap map[string]interface{}) types.Object {
	if createdByData, ok := dataProductMap["createdBy"].(map[string]interface{}); ok {
		attributeTypes := datasource_data_products.CreatedByValue{}.AttributeTypes(ctx)
		attributes := map[string]attr.Value{}

		if userId, ok := createdByData["userId"].(string); ok {
			attributes["user_id"] = types.StringValue(userId)
		} else {
			attributes["user_id"] = types.StringNull()
		}

		if email, ok := createdByData["email"].(string); ok {
			attributes["email"] = types.StringValue(email)
		} else {
			attributes["email"] = types.StringNull()
		}

		createdByValue, diags := datasource_data_products.NewCreatedByValue(attributeTypes, attributes)
		if diags.HasError() {
			tflog.Error(ctx, fmt.Sprintf("Error creating createdBy value: %v", diags))
			createdByObjectValue, _ := datasource_data_products.NewCreatedByValueNull().ToObjectValue(ctx)
			return createdByObjectValue
		}

		createdByObjectValue, _ := createdByValue.ToObjectValue(ctx)
		return createdByObjectValue
	}

	createdByObjectValue, _ := datasource_data_products.NewCreatedByValueNull().ToObjectValue(ctx)
	return createdByObjectValue
}

func (d *data_productsDataSource) mapModifiedByObject(ctx context.Context, dataProductMap map[string]interface{}) types.Object {
	if modifiedByData, ok := dataProductMap["modifiedBy"].(map[string]interface{}); ok {
		attributeTypes := datasource_data_products.ModifiedByValue{}.AttributeTypes(ctx)
		attributes := map[string]attr.Value{}

		if userId, ok := modifiedByData["userId"].(string); ok {
			attributes["user_id"] = types.StringValue(userId)
		} else {
			attributes["user_id"] = types.StringNull()
		}

		if email, ok := modifiedByData["email"].(string); ok {
			attributes["email"] = types.StringValue(email)
		} else {
			attributes["email"] = types.StringNull()
		}

		modifiedByValue, diags := datasource_data_products.NewModifiedByValue(attributeTypes, attributes)
		if diags.HasError() {
			tflog.Error(ctx, fmt.Sprintf("Error creating modifiedBy value: %v", diags))
			modifiedByObjectValue, _ := datasource_data_products.NewModifiedByValueNull().ToObjectValue(ctx)
			return modifiedByObjectValue
		}

		modifiedByObjectValue, _ := modifiedByValue.ToObjectValue(ctx)
		return modifiedByObjectValue
	}

	modifiedByObjectValue, _ := datasource_data_products.NewModifiedByValueNull().ToObjectValue(ctx)
	return modifiedByObjectValue
}

func (d *data_productsDataSource) mapContactsList(ctx context.Context, dataProductMap map[string]interface{}) types.List {
	if contactsData, ok := dataProductMap["contacts"].([]interface{}); ok {
		contactsList := make([]datasource_data_products.ContactsValue, 0, len(contactsData))

		for _, contactInterface := range contactsData {
			if contactMap, ok := contactInterface.(map[string]interface{}); ok {
				attributeTypes := datasource_data_products.ContactsValue{}.AttributeTypes(ctx)
				attributes := map[string]attr.Value{}

				if contactType, ok := contactMap["type"].(string); ok {
					attributes["type"] = types.StringValue(contactType)
				} else {
					attributes["type"] = types.StringNull()
				}

				if value, ok := contactMap["value"].(string); ok {
					attributes["value"] = types.StringValue(value)
				} else {
					attributes["value"] = types.StringNull()
				}

				contactValue, diags := datasource_data_products.NewContactsValue(attributeTypes, attributes)
				if !diags.HasError() {
					contactsList = append(contactsList, contactValue)
				}
			}
		}

		elementType := datasource_data_products.ContactsType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_data_products.ContactsValue{}.AttributeTypes(ctx),
			},
		}

		if len(contactsList) > 0 {
			listValue, _ := types.ListValueFrom(ctx, elementType, contactsList)
			return listValue
		}
	}

	elementType := datasource_data_products.ContactsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_data_products.ContactsValue{}.AttributeTypes(ctx),
		},
	}
	return types.ListNull(elementType)
}

func (d *data_productsDataSource) mapLinksList(ctx context.Context, dataProductMap map[string]interface{}) types.List {
	if linksData, ok := dataProductMap["links"].([]interface{}); ok {
		linksList := make([]datasource_data_products.LinksValue, 0, len(linksData))

		for _, linkInterface := range linksData {
			if linkMap, ok := linkInterface.(map[string]interface{}); ok {
				attributeTypes := datasource_data_products.LinksValue{}.AttributeTypes(ctx)
				attributes := map[string]attr.Value{}

				if linkType, ok := linkMap["type"].(string); ok {
					attributes["type"] = types.StringValue(linkType)
				} else {
					attributes["type"] = types.StringNull()
				}

				if url, ok := linkMap["url"].(string); ok {
					attributes["url"] = types.StringValue(url)
				} else {
					attributes["url"] = types.StringNull()
				}

				if description, ok := linkMap["description"].(string); ok {
					attributes["description"] = types.StringValue(description)
				} else {
					attributes["description"] = types.StringNull()
				}

				linkValue, diags := datasource_data_products.NewLinksValue(attributeTypes, attributes)
				if !diags.HasError() {
					linksList = append(linksList, linkValue)
				}
			}
		}

		elementType := datasource_data_products.LinksType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_data_products.LinksValue{}.AttributeTypes(ctx),
			},
		}

		if len(linksList) > 0 {
			listValue, _ := types.ListValueFrom(ctx, elementType, linksList)
			return listValue
		}
	}

	elementType := datasource_data_products.LinksType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_data_products.LinksValue{}.AttributeTypes(ctx),
		},
	}
	return types.ListNull(elementType)
}
