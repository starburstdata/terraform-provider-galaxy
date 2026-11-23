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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_data_domain"
)

var _ resource.Resource = (*dataDomainResource)(nil)
var _ resource.ResourceWithConfigure = (*dataDomainResource)(nil)
var _ resource.ResourceWithImportState = (*dataDomainResource)(nil)

func NewDataDomainResource() resource.Resource {
	return &dataDomainResource{}
}

type dataDomainResource struct {
	client *client.GalaxyClient
}

func (r *dataDomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_domain"
}

func (r *dataDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_data_domain.DataDomainResourceSchema(ctx)

	if attr, ok := s.Attributes["data_domain_id"].(schema.StringAttribute); ok {
		attr.PlanModifiers = []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
		s.Attributes["data_domain_id"] = attr
	}
	if attr, ok := s.Attributes["created_on"].(schema.StringAttribute); ok {
		attr.PlanModifiers = []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
		s.Attributes["created_on"] = attr
	}

	resp.Schema = s
}

func (r *dataDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.GalaxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.GalaxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *dataDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_data_domain.DataDomainModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(&plan)

	tflog.Debug(ctx, "Creating data domain")
	response, err := r.client.CreateDataDomain(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating data domain",
			"Could not create data domain: "+err.Error(),
		)
		return
	}

	r.unknownToNull(&plan)
	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created data domain", map[string]interface{}{"id": plan.DataDomainId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dataDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_data_domain.DataDomainModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.DataDomainId.ValueString()
	tflog.Debug(ctx, "Reading data domain", map[string]interface{}{"id": id})
	response, err := r.client.GetDataDomain(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading data domain",
			"Could not read data domain "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dataDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_data_domain.DataDomainModel
	var state resource_data_domain.DataDomainModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.DataDomainId.ValueString()
	request := r.modelToCreateRequest(&plan)

	response, err := r.client.UpdateDataDomain(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating data domain",
			"Could not update data domain "+id+": "+err.Error(),
		)
		return
	}

	plan.DataDomainId = state.DataDomainId
	r.unknownToNull(&plan)
	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dataDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_data_domain.DataDomainModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.DataDomainId.ValueString()
	if err := r.client.DeleteDataDomain(ctx, id); err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting data domain",
				"Could not delete data domain "+id+": "+err.Error(),
			)
			return
		}
	}
}

func (r *dataDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("data_domain_id"), req, resp)
}

func (r *dataDomainResource) modelToCreateRequest(model *resource_data_domain.DataDomainModel) map[string]interface{} {
	request := make(map[string]interface{})

	if !model.Name.IsNull() && !model.Name.IsUnknown() && model.Name.ValueString() != "" {
		request["name"] = model.Name.ValueString()
	}
	if !model.Description.IsNull() && !model.Description.IsUnknown() && model.Description.ValueString() != "" {
		request["description"] = model.Description.ValueString()
	}
	if !model.ContactUserId.IsNull() && !model.ContactUserId.IsUnknown() && model.ContactUserId.ValueString() != "" {
		request["contactUserId"] = model.ContactUserId.ValueString()
	}
	if !model.RoleId.IsNull() && !model.RoleId.IsUnknown() && model.RoleId.ValueString() != "" {
		request["roleId"] = model.RoleId.ValueString()
	}

	return request
}

func (r *dataDomainResource) unknownToNull(model *resource_data_domain.DataDomainModel) {
	if model.Description.IsUnknown() {
		model.Description = types.StringNull()
	}
	if model.ContactUserId.IsUnknown() {
		model.ContactUserId = types.StringNull()
	}
	if model.RoleId.IsUnknown() {
		model.RoleId = types.StringNull()
	}
	if model.CreatedOn.IsUnknown() {
		model.CreatedOn = types.StringNull()
	}
	if model.ModifiedOn.IsUnknown() {
		model.ModifiedOn = types.StringNull()
	}
	if model.IsDefault.IsUnknown() {
		model.IsDefault = types.BoolNull()
	}
}

func (r *dataDomainResource) updateModelFromResponse(ctx context.Context, model *resource_data_domain.DataDomainModel, response map[string]interface{}, diags *diag.Diagnostics) {
	if id, ok := response["dataDomainId"].(string); ok {
		model.DataDomainId = types.StringValue(id)
		model.Id = types.StringValue(id)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else if model.Description.IsUnknown() {
		model.Description = types.StringNull()
	}
	if contactUserId, ok := response["contactUserId"].(string); ok {
		model.ContactUserId = types.StringValue(contactUserId)
	} else if model.ContactUserId.IsUnknown() {
		model.ContactUserId = types.StringNull()
	}
	if roleId, ok := response["roleId"].(string); ok {
		model.RoleId = types.StringValue(roleId)
	} else if roleMap, ok := response["owningRole"].(map[string]interface{}); ok {
		if rid, ok := roleMap["roleId"].(string); ok {
			model.RoleId = types.StringValue(rid)
		} else if model.RoleId.IsUnknown() {
			model.RoleId = types.StringNull()
		}
	} else if model.RoleId.IsUnknown() {
		model.RoleId = types.StringNull()
	}
	if createdOn, ok := response["createdOn"].(string); ok {
		model.CreatedOn = types.StringValue(createdOn)
	} else if model.CreatedOn.IsUnknown() {
		model.CreatedOn = types.StringNull()
	}
	if modifiedOn, ok := response["modifiedOn"].(string); ok {
		model.ModifiedOn = types.StringValue(modifiedOn)
	} else if model.ModifiedOn.IsUnknown() {
		model.ModifiedOn = types.StringNull()
	}
	if isDefault, ok := response["isDefault"].(bool); ok {
		model.IsDefault = types.BoolValue(isDefault)
	} else if model.IsDefault.IsUnknown() {
		model.IsDefault = types.BoolNull()
	}

	// Nested computed objects (Contact / CreatedBy / ModifiedBy share the same shape)
	contactAttrs := buildUserAttrs(response["contact"])
	contact, d := resource_data_domain.NewContactValue(resource_data_domain.ContactValue{}.AttributeTypes(ctx), contactAttrs)
	diags.Append(d...)
	model.Contact = contact

	createdByAttrs := buildUserAttrs(response["createdBy"])
	createdBy, d := resource_data_domain.NewCreatedByValue(resource_data_domain.CreatedByValue{}.AttributeTypes(ctx), createdByAttrs)
	diags.Append(d...)
	model.CreatedBy = createdBy

	modifiedByAttrs := buildUserAttrs(response["modifiedBy"])
	modifiedBy, d := resource_data_domain.NewModifiedByValue(resource_data_domain.ModifiedByValue{}.AttributeTypes(ctx), modifiedByAttrs)
	diags.Append(d...)
	model.ModifiedBy = modifiedBy

	model.OwningRole = buildOwningRoleValue(ctx, response["owningRole"], diags)
	model.AssignedDataProducts = buildAssignedDataProducts(ctx, response["assignedDataProducts"], diags)
}

func buildUserAttrs(raw interface{}) map[string]attr.Value {
	attrs := map[string]attr.Value{
		"email":   types.StringNull(),
		"user_id": types.StringNull(),
	}
	userMap, ok := raw.(map[string]interface{})
	if !ok {
		return attrs
	}
	if email, ok := userMap["email"].(string); ok {
		attrs["email"] = types.StringValue(email)
	}
	if userId, ok := userMap["userId"].(string); ok {
		attrs["user_id"] = types.StringValue(userId)
	}
	return attrs
}

func buildOwningRoleValue(ctx context.Context, raw interface{}, diags *diag.Diagnostics) resource_data_domain.OwningRoleValue {
	attrs := map[string]attr.Value{
		"role_id":   types.StringNull(),
		"role_name": types.StringNull(),
	}
	roleMap, ok := raw.(map[string]interface{})
	if !ok {
		v, d := resource_data_domain.NewOwningRoleValue(resource_data_domain.OwningRoleValue{}.AttributeTypes(ctx), attrs)
		diags.Append(d...)
		return v
	}
	if roleId, ok := roleMap["roleId"].(string); ok {
		attrs["role_id"] = types.StringValue(roleId)
	}
	if roleName, ok := roleMap["roleName"].(string); ok {
		attrs["role_name"] = types.StringValue(roleName)
	}
	v, d := resource_data_domain.NewOwningRoleValue(resource_data_domain.OwningRoleValue{}.AttributeTypes(ctx), attrs)
	diags.Append(d...)
	return v
}

func buildAssignedDataProducts(ctx context.Context, raw interface{}, diags *diag.Diagnostics) types.List {
	elementType := resource_data_domain.AssignedDataProductsType{
		ObjectType: types.ObjectType{
			AttrTypes: resource_data_domain.AssignedDataProductsValue{}.AttributeTypes(ctx),
		},
	}
	items := make([]resource_data_domain.AssignedDataProductsValue, 0)
	if arr, ok := raw.([]interface{}); ok {
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
			v, d := resource_data_domain.NewAssignedDataProductsValue(resource_data_domain.AssignedDataProductsValue{}.AttributeTypes(ctx), attrs)
			diags.Append(d...)
			items = append(items, v)
		}
	}
	listValue, d := types.ListValueFrom(ctx, elementType, items)
	diags.Append(d...)
	return listValue
}
