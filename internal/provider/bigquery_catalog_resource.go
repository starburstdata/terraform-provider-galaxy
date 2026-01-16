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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_bigquery_catalog"
)

var _ resource.Resource = (*bigquery_catalogResource)(nil)
var _ resource.ResourceWithConfigure = (*bigquery_catalogResource)(nil)

func NewBigqueryCatalogResource() resource.Resource {
	return &bigquery_catalogResource{}
}

type bigquery_catalogResource struct {
	client *client.GalaxyClient
}

func (r *bigquery_catalogResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bigquery_catalog"
}

func (r *bigquery_catalogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_bigquery_catalog.BigqueryCatalogResourceSchema(ctx)

	// Fix: validate is a request-only parameter, not returned by API.
	// Setting Computed=false ensures it's sent with update requests.
	if attr, ok := s.Attributes["validate"].(schema.BoolAttribute); ok {
		attr.Computed = false
		s.Attributes["validate"] = attr
	}

	resp.Schema = s
}

func (r *bigquery_catalogResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bigquery_catalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_bigquery_catalog.BigqueryCatalogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating bigquery_catalog")
	response, err := r.client.CreateCatalog(ctx, "bigquery", request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating bigquery_catalog",
			"Could not create bigquery_catalog: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created bigquery_catalog", map[string]interface{}{"id": plan.CatalogId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bigquery_catalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_bigquery_catalog.BigqueryCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	tflog.Debug(ctx, "Reading bigquery_catalog", map[string]interface{}{"id": id})
	response, err := r.client.GetCatalog(ctx, "bigquery", id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "BigqueryCatalog not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading bigquery_catalog",
			"Could not read bigquery_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bigquery_catalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_bigquery_catalog.BigqueryCatalogModel
	var state resource_bigquery_catalog.BigqueryCatalogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	request := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating bigquery_catalog", map[string]interface{}{"id": id})
	response, err := r.client.UpdateCatalog(ctx, "bigquery", id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating bigquery_catalog",
			"Could not update bigquery_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated bigquery_catalog", map[string]interface{}{"id": plan.CatalogId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bigquery_catalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_bigquery_catalog.BigqueryCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	tflog.Debug(ctx, "Deleting bigquery_catalog", map[string]interface{}{"id": id})
	err := r.client.DeleteCatalog(ctx, "bigquery", id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting bigquery_catalog",
				"Could not delete bigquery_catalog "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted bigquery_catalog", map[string]interface{}{"id": id})
}

// Helper methods
func (r *bigquery_catalogResource) modelToCreateRequest(ctx context.Context, model *resource_bigquery_catalog.BigqueryCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	request["name"] = model.Name.ValueString()
	request["readOnly"] = model.ReadOnly.ValueBool()

	// Debug credentials_key
	credentialsKey := model.CredentialsKey.ValueString()
	tflog.Debug(ctx, "BigQuery catalog credentials_key", map[string]interface{}{
		"credentialsKey": credentialsKey,
		"isNull":         model.CredentialsKey.IsNull(),
		"isUnknown":      model.CredentialsKey.IsUnknown(),
		"isEmpty":        credentialsKey == "",
	})

	if credentialsKey == "" {
		diags.AddError(
			"Missing required field",
			"credentials_key cannot be empty for BigQuery catalog",
		)
		return request
	}

	request["credentialsKey"] = credentialsKey

	// Optional fields
	if !model.Description.IsNull() && !model.Description.IsUnknown() && model.Description.ValueString() != "" {
		request["description"] = model.Description.ValueString()
	}

	if !model.ParentProjectId.IsNull() && !model.ParentProjectId.IsUnknown() && model.ParentProjectId.ValueString() != "" {
		request["parentProjectId"] = model.ParentProjectId.ValueString()
	}

	if !model.ProjectId.IsNull() && !model.ProjectId.IsUnknown() && model.ProjectId.ValueString() != "" {
		request["projectId"] = model.ProjectId.ValueString()
	}

	if !model.Validate.IsNull() && !model.Validate.IsUnknown() {
		request["validate"] = model.Validate.ValueBool()
	}

	return request
}

func (r *bigquery_catalogResource) modelToUpdateRequest(ctx context.Context, model *resource_bigquery_catalog.BigqueryCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *bigquery_catalogResource) updateModelFromResponse(ctx context.Context, model *resource_bigquery_catalog.BigqueryCatalogModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
	// Use catalogId as the ID
	if catalogId, ok := response["catalogId"].(string); ok {
		model.CatalogId = types.StringValue(catalogId)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else if model.Description.IsUnknown() {
		model.Description = types.StringNull()
	}
	if readOnly, ok := response["readOnly"].(bool); ok {
		model.ReadOnly = types.BoolValue(readOnly)
	}
	if credentialsKey, ok := response["credentialsKey"].(string); ok {
		// Don't update credentials if they're encrypted
		if credentialsKey != "<Value is encrypted>" {
			model.CredentialsKey = types.StringValue(credentialsKey)
		}
	}

	// Map additional BigQuery specific fields
	if parentProjectId, ok := response["parentProjectId"].(string); ok {
		model.ParentProjectId = types.StringValue(parentProjectId)
	} else if model.ParentProjectId.IsUnknown() {
		model.ParentProjectId = types.StringNull()
	}

	if projectId, ok := response["projectId"].(string); ok {
		model.ProjectId = types.StringValue(projectId)
	} else if model.ProjectId.IsUnknown() {
		model.ProjectId = types.StringNull()
	}

	if validate, ok := response["validate"].(bool); ok {
		model.Validate = types.BoolValue(validate)
	} else if model.Validate.IsUnknown() {
		model.Validate = types.BoolNull()
	}
}
