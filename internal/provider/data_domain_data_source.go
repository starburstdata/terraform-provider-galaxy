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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_data_domain"
)

var _ datasource.DataSource = (*dataDomainDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*dataDomainDataSource)(nil)

func NewDataDomainDataSource() datasource.DataSource {
	return &dataDomainDataSource{}
}

type dataDomainDataSource struct {
	client *client.GalaxyClient
}

func (d *dataDomainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_domain"
}

func (d *dataDomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_data_domain.DataDomainDataSourceSchema(ctx)
}

func (d *dataDomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.GalaxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.GalaxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *dataDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_data_domain.DataDomainModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.DataDomainId.ValueString()
	tflog.Debug(ctx, "Reading data domain", map[string]interface{}{"id": id})

	response, err := d.client.GetDataDomain(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading data domain",
			"Could not read data domain "+id+": "+err.Error(),
		)
		return
	}

	d.updateModelFromResponse(ctx, &config, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *dataDomainDataSource) updateModelFromResponse(ctx context.Context, model *datasource_data_domain.DataDomainModel, response map[string]interface{}, diags *diag.Diagnostics) {
	if id, ok := response["dataDomainId"].(string); ok {
		model.DataDomainId = types.StringValue(id)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	} else {
		model.Name = types.StringNull()
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else {
		model.Description = types.StringNull()
	}
	if createdOn, ok := response["createdOn"].(string); ok {
		model.CreatedOn = types.StringValue(createdOn)
	} else {
		model.CreatedOn = types.StringNull()
	}
	if modifiedOn, ok := response["modifiedOn"].(string); ok {
		model.ModifiedOn = types.StringValue(modifiedOn)
	} else {
		model.ModifiedOn = types.StringNull()
	}
	if isDefault, ok := response["isDefault"].(bool); ok {
		model.IsDefault = types.BoolValue(isDefault)
	} else {
		model.IsDefault = types.BoolNull()
	}

	contactAttrs := buildUserAttrs(response["contact"])
	contact, dd := datasource_data_domain.NewContactValue(datasource_data_domain.ContactValue{}.AttributeTypes(ctx), contactAttrs)
	diags.Append(dd...)
	model.Contact = contact

	createdByAttrs := buildUserAttrs(response["createdBy"])
	createdBy, dd := datasource_data_domain.NewCreatedByValue(datasource_data_domain.CreatedByValue{}.AttributeTypes(ctx), createdByAttrs)
	diags.Append(dd...)
	model.CreatedBy = createdBy

	modifiedByAttrs := buildUserAttrs(response["modifiedBy"])
	modifiedBy, dd := datasource_data_domain.NewModifiedByValue(datasource_data_domain.ModifiedByValue{}.AttributeTypes(ctx), modifiedByAttrs)
	diags.Append(dd...)
	model.ModifiedBy = modifiedBy

	owningAttrs := map[string]attr.Value{
		"role_id":   types.StringNull(),
		"role_name": types.StringNull(),
	}
	if roleMap, ok := response["owningRole"].(map[string]interface{}); ok {
		if v, ok := roleMap["roleId"].(string); ok {
			owningAttrs["role_id"] = types.StringValue(v)
		}
		if v, ok := roleMap["roleName"].(string); ok {
			owningAttrs["role_name"] = types.StringValue(v)
		}
	}
	owning, dd := datasource_data_domain.NewOwningRoleValue(datasource_data_domain.OwningRoleValue{}.AttributeTypes(ctx), owningAttrs)
	diags.Append(dd...)
	model.OwningRole = owning

	elementType := datasource_data_domain.AssignedDataProductsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_data_domain.AssignedDataProductsValue{}.AttributeTypes(ctx),
		},
	}
	items := make([]datasource_data_domain.AssignedDataProductsValue, 0)
	if arr, ok := response["assignedDataProducts"].([]interface{}); ok {
		for _, item := range arr {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			attrs := map[string]attr.Value{
				"data_product_id": types.StringNull(),
				"name":            types.StringNull(),
			}
			if v, ok := m["dataProductId"].(string); ok {
				attrs["data_product_id"] = types.StringValue(v)
			}
			if v, ok := m["name"].(string); ok {
				attrs["name"] = types.StringValue(v)
			}
			v, ddd := datasource_data_domain.NewAssignedDataProductsValue(datasource_data_domain.AssignedDataProductsValue{}.AttributeTypes(ctx), attrs)
			diags.Append(ddd...)
			items = append(items, v)
		}
	}
	listValue, ddd := types.ListValueFrom(ctx, elementType, items)
	diags.Append(ddd...)
	model.AssignedDataProducts = listValue
}
