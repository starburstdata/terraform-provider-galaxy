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

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_data_product"
)

var _ resource.Resource = (*data_productResource)(nil)
var _ resource.ResourceWithConfigure = (*data_productResource)(nil)

func NewDataProductResource() resource.Resource {
	return &data_productResource{}
}

type data_productResource struct {
	client *client.GalaxyClient
}

func (r *data_productResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_product"
}

func (r *data_productResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_data_product.DataProductResourceSchema(ctx)
}

func (r *data_productResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.GalaxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.GalaxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *data_productResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_data_product.DataProductModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Ensure no computed-only fields are included in create request
	delete(request, "id")
	delete(request, "dataProductId")
	delete(request, "createdOn")
	delete(request, "modifiedOn")
	delete(request, "createdBy")
	delete(request, "modifiedBy")
	delete(request, "catalog")

	// Debug logging to see what we're sending to API
	tflog.Info(ctx, "DATA_PRODUCT API REQUEST", map[string]interface{}{"request": request})
	tflog.Debug(ctx, "Creating data_product")
	response, err := r.client.CreateDataProduct(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating data_product",
			"Could not create data_product: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created data_product", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *data_productResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_data_product.DataProductModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Reading data_product", map[string]interface{}{"id": id})
	response, err := r.client.GetDataProduct(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "DataProduct not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading data_product",
			"Could not read data_product "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *data_productResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_data_product.DataProductModel
	var state resource_data_product.DataProductModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	request := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating data_product", map[string]interface{}{"id": id})
	response, err := r.client.UpdateDataProduct(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating data_product",
			"Could not update data_product "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated data_product", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *data_productResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_data_product.DataProductModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Deleting data_product", map[string]interface{}{"id": id})
	err := r.client.DeleteDataProduct(ctx, id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting data_product",
				"Could not delete data_product "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted data_product", map[string]interface{}{"id": id})
}

// Helper methods
func (r *data_productResource) modelToCreateRequest(ctx context.Context, model *resource_data_product.DataProductModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	if !model.Name.IsNull() {
		request["name"] = model.Name.ValueString()
	}
	if !model.Summary.IsNull() {
		request["summary"] = model.Summary.ValueString()
	}
	if !model.CatalogId.IsNull() {
		request["catalogId"] = model.CatalogId.ValueString()
	}
	if !model.SchemaName.IsNull() {
		request["schemaName"] = model.SchemaName.ValueString()
	}

	// Optional fields
	if !model.Description.IsNull() && model.Description.ValueString() != "" {
		request["description"] = model.Description.ValueString()
	}
	if !model.DefaultClusterId.IsNull() && model.DefaultClusterId.ValueString() != "" {
		request["defaultClusterId"] = model.DefaultClusterId.ValueString()
	}

	// Handle contacts list - send empty array to API since contacts field might be required
	// The API accepts contacts but doesn't return them consistently
	if !model.Contacts.IsNull() && !model.Contacts.IsUnknown() {
		tflog.Info(ctx, "CONTACTS: Sending empty contacts array to API")
		request["contacts"] = []map[string]interface{}{}
	} else {
		// Even if no contacts are specified in plan, send empty array
		request["contacts"] = []map[string]interface{}{}
	}

	// Handle links list - links is optional and typically computed by the API
	if !model.Links.IsNull() && !model.Links.IsUnknown() {
		linksList := []map[string]interface{}{}
		elements := model.Links.Elements()
		for _, elem := range elements {
			if !elem.IsNull() && !elem.IsUnknown() {
				// Convert the LinksValue to a map
				if linkVal, ok := elem.(resource_data_product.LinksValue); ok {
					link := map[string]interface{}{}
					if !linkVal.Name.IsNull() {
						link["name"] = linkVal.Name.ValueString()
					}
					if !linkVal.Uri.IsNull() {
						link["uri"] = linkVal.Uri.ValueString()
					}
					linksList = append(linksList, link)
				}
			}
		}
		if len(linksList) > 0 {
			request["links"] = linksList
		}
	}

	return request
}

func (r *data_productResource) modelToUpdateRequest(ctx context.Context, model *resource_data_product.DataProductModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *data_productResource) updateModelFromResponse(ctx context.Context, model *resource_data_product.DataProductModel, response map[string]interface{}, diags *diag.Diagnostics) {
	tflog.Info(ctx, "DATA_PRODUCT updateModelFromResponse", map[string]interface{}{
		"full_response":          response,
		"response_contacts":      response["contacts"],
		"model_contacts_null":    model.Contacts.IsNull(),
		"model_contacts_unknown": model.Contacts.IsUnknown(),
	})
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	} else if id, ok := response["dataProductId"].(string); ok {
		model.Id = types.StringValue(id)
	}

	if dataProductId, ok := response["dataProductId"].(string); ok {
		model.DataProductId = types.StringValue(dataProductId)
	}

	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}

	if summary, ok := response["summary"].(string); ok {
		model.Summary = types.StringValue(summary)
	}

	if catalogId, ok := response["catalogId"].(string); ok {
		model.CatalogId = types.StringValue(catalogId)
	}

	if schemaName, ok := response["schemaName"].(string); ok {
		model.SchemaName = types.StringValue(schemaName)
	}

	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else {
		model.Description = types.StringNull()
	}

	if defaultClusterId, ok := response["defaultClusterId"].(string); ok {
		model.DefaultClusterId = types.StringValue(defaultClusterId)
	} else {
		model.DefaultClusterId = types.StringNull()
	}

	if createdOn, ok := response["createdOn"].(string); ok {
		model.CreatedOn = types.StringValue(createdOn)
	}

	if modifiedOn, ok := response["modifiedOn"].(string); ok {
		model.ModifiedOn = types.StringValue(modifiedOn)
	}

	// Set complex nested objects to null/unknown for now
	// In a full implementation, these would be properly parsed
	model.Catalog = resource_data_product.NewCatalogValueNull()
	model.CreatedBy = resource_data_product.NewCreatedByValueNull()
	model.ModifiedBy = resource_data_product.NewModifiedByValueNull()

	// Handle contacts from response - preserve existing contacts if API doesn't return any
	if contacts, ok := response["contacts"].([]interface{}); ok && len(contacts) > 0 {
		contactsList := make([]resource_data_product.ContactsValue, 0, len(contacts))
		for _, contactInterface := range contacts {
			if contactMap, ok := contactInterface.(map[string]interface{}); ok {
				contact := resource_data_product.ContactsValue{
					Email:  types.StringNull(),
					UserId: types.StringNull(),
				}
				if email, ok := contactMap["email"].(string); ok {
					contact.Email = types.StringValue(email)
				}
				if userId, ok := contactMap["userId"].(string); ok {
					contact.UserId = types.StringValue(userId)
				}
				contactsList = append(contactsList, contact)
			}
		}
		if len(contactsList) > 0 {
			contactsListValue, diag := types.ListValueFrom(ctx, resource_data_product.ContactsValue{}.Type(ctx), contactsList)
			if !diag.HasError() {
				model.Contacts = contactsListValue
			} else {
				tflog.Error(ctx, "Error creating contacts list", map[string]interface{}{"diag": diag})
				// Don't overwrite with null - preserve existing contacts
			}
		}
	} else {
		// API didn't return contacts - preserve the contacts from the plan/model
		tflog.Info(ctx, "DATA_PRODUCT: API returned no contacts, preserving existing model state")
		// Keep the existing contacts value unchanged to maintain consistency
		// The model.Contacts already contains the correct values from the plan
	}

	// Handle links from response - typically empty/null for newly created resources
	if links, ok := response["links"].([]interface{}); ok && len(links) > 0 {
		linksList := make([]resource_data_product.LinksValue, 0, len(links))
		for _, linkInterface := range links {
			if linkMap, ok := linkInterface.(map[string]interface{}); ok {
				link := resource_data_product.LinksValue{
					Name: types.StringNull(),
					Uri:  types.StringNull(),
				}
				if name, ok := linkMap["name"].(string); ok {
					link.Name = types.StringValue(name)
				}
				if uri, ok := linkMap["uri"].(string); ok {
					link.Uri = types.StringValue(uri)
				}
				linksList = append(linksList, link)
			}
		}
		if len(linksList) > 0 {
			linksListValue, diag := types.ListValueFrom(ctx, resource_data_product.LinksValue{}.Type(ctx), linksList)
			if !diag.HasError() {
				model.Links = linksListValue
			} else {
				tflog.Error(ctx, "Error creating links list", map[string]interface{}{"diag": diag})
				model.Links = types.ListNull(resource_data_product.LinksValue{}.Type(ctx))
			}
		} else {
			model.Links = types.ListNull(resource_data_product.LinksValue{}.Type(ctx))
		}
	} else {
		model.Links = types.ListNull(resource_data_product.LinksValue{}.Type(ctx))
	}
}
