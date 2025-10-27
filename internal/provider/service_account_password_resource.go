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
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_service_account_password"
)

var _ resource.Resource = (*service_account_passwordResource)(nil)
var _ resource.ResourceWithConfigure = (*service_account_passwordResource)(nil)
var _ resource.ResourceWithImportState = (*service_account_passwordResource)(nil)

func NewServiceAccountPasswordResource() resource.Resource {
	return &service_account_passwordResource{}
}

type service_account_passwordResource struct {
	client *client.GalaxyClient
}

// Extended model with service account ID
type ServiceAccountPasswordModelExtended struct {
	resource_service_account_password.ServiceAccountPasswordModel
	ServiceAccountId types.String `tfsdk:"service_account_id"`
}

func (r *service_account_passwordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account_password"
}

func (r *service_account_passwordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	baseSchema := resource_service_account_password.ServiceAccountPasswordResourceSchema(ctx)

	// Add service_account_id as a required field
	baseSchema.Attributes["service_account_id"] = schema.StringAttribute{
		Required:            true,
		Description:         "The ID of the service account to create a password for",
		MarkdownDescription: "The ID of the service account to create a password for",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}

	resp.Schema = baseSchema
}

func (r *service_account_passwordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *service_account_passwordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServiceAccountPasswordModelExtended

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract the service account ID from the plan
	serviceAccountID := plan.ServiceAccountId.ValueString()

	tflog.Debug(ctx, "Creating service_account_password", map[string]interface{}{"service_account_id": serviceAccountID})
	response, err := r.client.CreateServiceAccountPassword(ctx, serviceAccountID, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating service_account_password",
			"Could not create service_account_password: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created service_account_password", map[string]interface{}{"id": plan.ServiceAccountPasswordId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *service_account_passwordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServiceAccountPasswordModelExtended

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the password ID directly and get service account ID from state
	serviceAccountID := state.ServiceAccountId.ValueString()
	passwordID := state.ServiceAccountPasswordId.ValueString()

	tflog.Debug(ctx, "Reading service_account_password - DEBUG IDs", map[string]interface{}{
		"service_account_id": serviceAccountID,
		"password_id":        passwordID,
		"state_dump":         fmt.Sprintf("%+v", state),
	})
	response, err := r.client.GetServiceAccountPassword(ctx, serviceAccountID, passwordID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "ServiceAccountPassword not found, removing from state", map[string]interface{}{"service_account_id": serviceAccountID, "password_id": passwordID})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading service_account_password",
			"Could not read service_account_password "+passwordID+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *service_account_passwordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ServiceAccountPasswordModelExtended
	var state ServiceAccountPasswordModelExtended

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.ServiceAccountPasswordId.ValueString()
	request := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating service_account_password", map[string]interface{}{"id": id})
	response, err := r.client.UpdateServiceAccountPassword(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service_account_password",
			"Could not update service_account_password "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated service_account_password", map[string]interface{}{"id": plan.ServiceAccountPasswordId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *service_account_passwordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServiceAccountPasswordModelExtended

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the password ID directly and get service account ID from state
	serviceAccountID := state.ServiceAccountId.ValueString()
	passwordID := state.ServiceAccountPasswordId.ValueString()

	tflog.Debug(ctx, "Deleting service_account_password", map[string]interface{}{"service_account_id": serviceAccountID, "password_id": passwordID})
	err := r.client.DeleteServiceAccountPassword(ctx, serviceAccountID, passwordID)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting service_account_password",
				"Could not delete service_account_password "+passwordID+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted service_account_password", map[string]interface{}{"service_account_id": serviceAccountID, "password_id": passwordID})
}

func (r *service_account_passwordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// The ID should be in the format "service_account_id/password_id"
	// For example: "u-4457989253/uat-4320576338177954"
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID Format",
			"Expected import ID in format 'service_account_id/password_id', got: "+req.ID,
		)
		return
	}

	serviceAccountID := parts[0]
	passwordID := parts[1]

	// Set the individual ID fields in state
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_account_id"), serviceAccountID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("service_account_password_id"), passwordID)...)
}

// Helper methods
func (r *service_account_passwordResource) modelToCreateRequest(ctx context.Context, model *ServiceAccountPasswordModelExtended, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Map description if provided
	if !model.Description.IsNull() && !model.Description.IsUnknown() {
		request["description"] = model.Description.ValueString()
	}

	return request
}

func (r *service_account_passwordResource) modelToUpdateRequest(ctx context.Context, model *ServiceAccountPasswordModelExtended, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *service_account_passwordResource) updateModelFromResponse(ctx context.Context, model *ServiceAccountPasswordModelExtended, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
	if id, ok := response["serviceAccountPasswordId"].(string); ok {
		model.ServiceAccountPasswordId = types.StringValue(id)
	} else if id, ok := response["id"].(string); ok {
		model.ServiceAccountPasswordId = types.StringValue(id)
	} else {
		model.ServiceAccountPasswordId = types.StringNull()
	}

	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else {
		model.Description = types.StringNull()
	}

	if password, ok := response["password"].(string); ok {
		model.Password = types.StringValue(password)
	} else {
		model.Password = types.StringNull()
	}

	if passwordPrefix, ok := response["passwordPrefix"].(string); ok {
		model.PasswordPrefix = types.StringValue(passwordPrefix)
	} else {
		model.PasswordPrefix = types.StringNull()
	}

	if created, ok := response["created"].(string); ok {
		model.Created = types.StringValue(created)
	} else {
		model.Created = types.StringNull()
	}

	if lastLogin, ok := response["lastLogin"].(string); ok {
		model.LastLogin = types.StringValue(lastLogin)
	} else {
		model.LastLogin = types.StringNull()
	}
}
