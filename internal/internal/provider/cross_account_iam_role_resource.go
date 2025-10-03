package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_cross_account_iam_role"
)

var _ resource.Resource = (*cross_account_iam_roleResource)(nil)
var _ resource.ResourceWithConfigure = (*cross_account_iam_roleResource)(nil)
var _ resource.ResourceWithImportState = (*cross_account_iam_roleResource)(nil)

func NewCrossAccountIamRoleResource() resource.Resource {
	return &cross_account_iam_roleResource{}
}

type cross_account_iam_roleResource struct {
	client *client.GalaxyClient
}

func (r *cross_account_iam_roleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cross_account_iam_role"
}

func (r *cross_account_iam_roleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_cross_account_iam_role.CrossAccountIamRoleResourceSchema(ctx)
}

func (r *cross_account_iam_roleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *cross_account_iam_roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_cross_account_iam_role.CrossAccountIamRoleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating cross_account_iam_role")
	response, err := r.client.CreateCrossAccountIamRole(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating cross_account_iam_role",
			"Could not create cross_account_iam_role: "+err.Error(),
		)
		return
	}

	// For cross-account IAM role, the Create response may return a list-like structure
	// We only need to update the basic fields for the resource, ignoring pagination metadata
	if aliasName, ok := response["aliasName"].(string); ok {
		plan.AliasName = types.StringValue(aliasName)
	}

	if awsIamArn, ok := response["awsIamArn"].(string); ok {
		plan.AwsIamArn = types.StringValue(awsIamArn)
	}

	// Handle dependants list if present
	if dependants, ok := response["dependants"].([]interface{}); ok {
		elements := make([]string, len(dependants))
		for i, dep := range dependants {
			if depStr, ok := dep.(string); ok {
				elements[i] = depStr
			}
		}
		listVal, diag := types.ListValueFrom(ctx, types.StringType, elements)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
		} else {
			plan.Dependants = listVal
		}
	} else {
		plan.Dependants = types.ListNull(types.StringType)
	}

	tflog.Debug(ctx, "Created cross_account_iam_role", map[string]interface{}{"alias_name": plan.AliasName.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cross_account_iam_roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_cross_account_iam_role.CrossAccountIamRoleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aliasName := state.AliasName.ValueString()
	tflog.Debug(ctx, "Reading cross_account_iam_role", map[string]interface{}{"alias_name": aliasName})
	response, err := r.client.GetCrossAccountIamRole(ctx, aliasName)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "CrossAccountIamRole not found, removing from state", map[string]interface{}{"alias_name": aliasName})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading cross_account_iam_role",
			"Could not read cross_account_iam_role "+aliasName+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cross_account_iam_roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_cross_account_iam_role.CrossAccountIamRoleModel
	var state resource_cross_account_iam_role.CrossAccountIamRoleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aliasName := state.AliasName.ValueString()
	request := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating cross_account_iam_role", map[string]interface{}{"alias_name": aliasName})
	response, err := r.client.UpdateCrossAccountIamRole(ctx, aliasName, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating cross_account_iam_role",
			"Could not update cross_account_iam_role "+aliasName+": "+err.Error(),
		)
		return
	}

	// Update only the fields that exist in our resource model
	if aliasNameResp, ok := response["aliasName"].(string); ok {
		plan.AliasName = types.StringValue(aliasNameResp)
	}

	if awsIamArn, ok := response["awsIamArn"].(string); ok {
		plan.AwsIamArn = types.StringValue(awsIamArn)
	}

	// Handle dependants list if present
	if dependants, ok := response["dependants"].([]interface{}); ok {
		elements := make([]string, len(dependants))
		for i, dep := range dependants {
			if depStr, ok := dep.(string); ok {
				elements[i] = depStr
			}
		}
		listVal, diag := types.ListValueFrom(ctx, types.StringType, elements)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
		} else {
			plan.Dependants = listVal
		}
	} else {
		plan.Dependants = types.ListNull(types.StringType)
	}

	tflog.Debug(ctx, "Updated cross_account_iam_role", map[string]interface{}{"alias_name": plan.AliasName.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cross_account_iam_roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_cross_account_iam_role.CrossAccountIamRoleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	aliasName := state.AliasName.ValueString()
	awsIamArn := state.AwsIamArn.ValueString()
	tflog.Debug(ctx, "Deleting cross_account_iam_role", map[string]interface{}{"alias_name": aliasName, "aws_iam_arn": awsIamArn})
	err := r.client.DeleteCrossAccountIamRole(ctx, awsIamArn)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting cross_account_iam_role",
				"Could not delete cross_account_iam_role "+aliasName+" (ARN: "+awsIamArn+"): "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted cross_account_iam_role", map[string]interface{}{"alias_name": aliasName, "aws_iam_arn": awsIamArn})
}

// Helper methods
func (r *cross_account_iam_roleResource) modelToCreateRequest(ctx context.Context, model *resource_cross_account_iam_role.CrossAccountIamRoleModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	if !model.AliasName.IsNull() {
		request["aliasName"] = model.AliasName.ValueString()
	}
	if !model.AwsIamArn.IsNull() {
		request["awsIamArn"] = model.AwsIamArn.ValueString()
	}

	return request
}

func (r *cross_account_iam_roleResource) modelToUpdateRequest(ctx context.Context, model *resource_cross_account_iam_role.CrossAccountIamRoleModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	// Note: SyncToken not available in this model

	return request
}

func (r *cross_account_iam_roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by alias name
	aliasName := req.ID

	// Set the alias_name attribute from the import ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("alias_name"), aliasName)...)
}

func (r *cross_account_iam_roleResource) updateModelFromResponse(ctx context.Context, model *resource_cross_account_iam_role.CrossAccountIamRoleModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model - Note: This resource uses aliasName as the identifier

	if aliasName, ok := response["aliasName"].(string); ok {
		model.AliasName = types.StringValue(aliasName)
	}

	if awsIamArn, ok := response["awsIamArn"].(string); ok {
		model.AwsIamArn = types.StringValue(awsIamArn)
	}

	// Handle dependants list
	if dependants, ok := response["dependants"].([]interface{}); ok {
		elements := make([]string, len(dependants))
		for i, dep := range dependants {
			if depStr, ok := dep.(string); ok {
				elements[i] = depStr
			}
		}
		listVal, diag := types.ListValueFrom(ctx, types.StringType, elements)
		if diag.HasError() {
			diags.Append(diag...)
		} else {
			model.Dependants = listVal
		}
	} else {
		model.Dependants = types.ListNull(types.StringType)
	}
}
