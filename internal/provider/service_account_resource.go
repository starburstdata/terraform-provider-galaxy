package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_service_account"
)

var _ resource.Resource = (*service_accountResource)(nil)
var _ resource.ResourceWithConfigure = (*service_accountResource)(nil)

func NewServiceAccountResource() resource.Resource {
	return &service_accountResource{}
}

type service_accountResource struct {
	client *client.GalaxyClient
}

func (r *service_accountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (r *service_accountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_service_account.ServiceAccountResourceSchema(ctx)
}

func (r *service_accountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *service_accountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_service_account.ServiceAccountModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating service_account")
	response, err := r.client.CreateServiceAccount(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating service_account",
			"Could not create service_account: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created service_account", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *service_accountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_service_account.ServiceAccountModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Reading service_account", map[string]interface{}{"id": id})
	response, err := r.client.GetServiceAccount(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "ServiceAccount not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading service_account",
			"Could not read service_account "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *service_accountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_service_account.ServiceAccountModel
	var state resource_service_account.ServiceAccountModel

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

	tflog.Debug(ctx, "Updating service_account", map[string]interface{}{"id": id})
	response, err := r.client.UpdateServiceAccount(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating service_account",
			"Could not update service_account "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated service_account", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *service_accountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_service_account.ServiceAccountModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Deleting service_account", map[string]interface{}{"id": id})
	err := r.client.DeleteServiceAccount(ctx, id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting service_account",
				"Could not delete service_account "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted service_account", map[string]interface{}{"id": id})
}

// Helper methods
func (r *service_accountResource) modelToCreateRequest(ctx context.Context, model *resource_service_account.ServiceAccountModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Map username
	if !model.Username.IsNull() && !model.Username.IsUnknown() {
		request["username"] = model.Username.ValueString()
	}

	// Map withInitialPassword
	if !model.WithInitialPassword.IsNull() && !model.WithInitialPassword.IsUnknown() {
		request["withInitialPassword"] = model.WithInitialPassword.ValueBool()
	}

	// Map additional role IDs - always include even if empty
	if !model.AdditionalRoleIds.IsNull() && !model.AdditionalRoleIds.IsUnknown() {
		var roleIds []string
		diags.Append(model.AdditionalRoleIds.ElementsAs(ctx, &roleIds, false)...)
		request["additionalRoleIds"] = roleIds
	} else {
		// Provide empty array if not specified
		request["additionalRoleIds"] = []string{}
	}

	// Map roleId if provided
	if !model.RoleId.IsNull() && !model.RoleId.IsUnknown() {
		request["roleId"] = model.RoleId.ValueString()
	}

	return request
}

func (r *service_accountResource) modelToUpdateRequest(ctx context.Context, model *resource_service_account.ServiceAccountModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *service_accountResource) updateModelFromResponse(ctx context.Context, model *resource_service_account.ServiceAccountModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
	if id, ok := response["id"].(string); ok {
		model.Id = types.StringValue(id)
	} else if id, ok := response["serviceAccountId"].(string); ok {
		model.Id = types.StringValue(id)
		model.ServiceAccountId = types.StringValue(id)
	}

	if username, ok := response["username"].(string); ok {
		model.Username = types.StringValue(username)
	}

	if userName, ok := response["userName"].(string); ok {
		model.UserName = types.StringValue(userName)
	}

	if roleId, ok := response["roleId"].(string); ok {
		model.RoleId = types.StringValue(roleId)
	}

	// Handle additional role IDs
	if additionalRoleIds, ok := response["additionalRoleIds"].([]interface{}); ok {
		var roleIds []attr.Value
		for _, id := range additionalRoleIds {
			if roleId, ok := id.(string); ok {
				roleIds = append(roleIds, types.StringValue(roleId))
			}
		}
		if len(roleIds) > 0 {
			model.AdditionalRoleIds = types.ListValueMust(types.StringType, roleIds)
		}
	}

	// Handle passwords array if present
	passwordAttrTypes := map[string]attr.Type{
		"created":                     types.StringType,
		"description":                 types.StringType,
		"last_login":                  types.StringType,
		"password":                    types.StringType,
		"password_prefix":             types.StringType,
		"service_account_password_id": types.StringType,
	}

	if passwords, ok := response["passwords"].([]interface{}); ok && len(passwords) > 0 {
		var passwordList []attr.Value
		for _, p := range passwords {
			if passwordMap, ok := p.(map[string]interface{}); ok {
				passwordAttrs := map[string]attr.Value{}

				if val, ok := passwordMap["created"].(string); ok {
					passwordAttrs["created"] = types.StringValue(val)
				} else {
					passwordAttrs["created"] = types.StringNull()
				}

				if val, ok := passwordMap["description"].(string); ok {
					passwordAttrs["description"] = types.StringValue(val)
				} else {
					passwordAttrs["description"] = types.StringNull()
				}

				if val, ok := passwordMap["lastLogin"].(string); ok {
					passwordAttrs["last_login"] = types.StringValue(val)
				} else {
					passwordAttrs["last_login"] = types.StringNull()
				}

				if val, ok := passwordMap["password"].(string); ok {
					passwordAttrs["password"] = types.StringValue(val)
				} else {
					passwordAttrs["password"] = types.StringNull()
				}

				if val, ok := passwordMap["passwordPrefix"].(string); ok {
					passwordAttrs["password_prefix"] = types.StringValue(val)
				} else {
					passwordAttrs["password_prefix"] = types.StringNull()
				}

				if val, ok := passwordMap["serviceAccountPasswordId"].(string); ok {
					passwordAttrs["service_account_password_id"] = types.StringValue(val)
				} else {
					passwordAttrs["service_account_password_id"] = types.StringNull()
				}

				passwordObj, _ := types.ObjectValue(passwordAttrTypes, passwordAttrs)
				passwordList = append(passwordList, passwordObj)
			}
		}

		if len(passwordList) > 0 {
			model.Passwords, _ = types.ListValue(types.ObjectType{AttrTypes: passwordAttrTypes}, passwordList)
		} else {
			model.Passwords = types.ListNull(types.ObjectType{AttrTypes: passwordAttrTypes})
		}
	} else {
		model.Passwords = types.ListNull(types.ObjectType{AttrTypes: passwordAttrTypes})
	}

	// Map other fields based on actual response structure
}
