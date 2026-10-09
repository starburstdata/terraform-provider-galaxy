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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_evaluation"
)

var _ resource.Resource = (*evaluationResource)(nil)
var _ resource.ResourceWithConfigure = (*evaluationResource)(nil)
var _ resource.ResourceWithImportState = (*evaluationResource)(nil)

func NewEvaluationResource() resource.Resource {
	return &evaluationResource{}
}

type evaluationResource struct {
	client *client.GalaxyClient
}

func (r *evaluationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_evaluation"
}

func (r *evaluationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_evaluation.EvaluationResourceSchema(ctx)

	// data_quality_check_id identifies the target check; the API has no way to move an evaluation
	// between checks, so any change must destroy and recreate the resource.
	if attr, ok := s.Attributes["data_quality_check_id"].(schema.StringAttribute); ok {
		attr.PlanModifiers = append(attr.PlanModifiers, stringplanmodifier.RequiresReplace())
		s.Attributes["data_quality_check_id"] = attr
	}

	// cluster_id is a request-only Create parameter (see Update; the API never returns it). Changing
	// it in config would silently do nothing without RequiresReplace. UseStateForUnknown preserves
	// the recorded value across no-op updates.
	if attr, ok := s.Attributes["cluster_id"].(schema.StringAttribute); ok {
		attr.PlanModifiers = append(attr.PlanModifiers,
			stringplanmodifier.UseStateForUnknown(),
			stringplanmodifier.RequiresReplaceIf(requiresReplaceIfClusterIdRecorded,
				"Changing cluster_id re-runs the evaluation.",
				"Changing cluster_id re-runs the evaluation."),
		)
		s.Attributes["cluster_id"] = attr
	}

	if attr, ok := s.Attributes["id"].(schema.StringAttribute); ok {
		attr.PlanModifiers = append(attr.PlanModifiers, stringplanmodifier.UseStateForUnknown())
		s.Attributes["id"] = attr
	}

	resp.Schema = s
}

// requiresReplaceIfClusterIdRecorded skips replacement when state has no cluster_id (e.g. after import), since the API never returns it
func requiresReplaceIfClusterIdRecorded(ctx context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	resp.RequiresReplace = !req.StateValue.IsNull()
}

func (r *evaluationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *evaluationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_evaluation.EvaluationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkID := plan.DataQualityCheckId.ValueString()

	request := map[string]interface{}{
		"dataQualityCheckId": checkID,
	}
	if !plan.ClusterId.IsNull() && !plan.ClusterId.IsUnknown() && plan.ClusterId.ValueString() != "" {
		request["clusterId"] = plan.ClusterId.ValueString()
	}

	tflog.Debug(ctx, "Creating evaluation", map[string]interface{}{"data_quality_check_id": checkID})
	if _, err := r.client.CreateEvaluation(ctx, checkID, request); err != nil {
		resp.Diagnostics.AddError(
			"Error creating evaluation",
			"Could not create evaluation: "+err.Error(),
		)
		return
	}

	// The evaluate endpoint returns 204 with no body; read back to populate computed fields.
	response, err := r.client.GetEvaluation(ctx, checkID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading evaluation",
			"Could not read evaluation for check "+checkID+" after triggering it: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response)

	if plan.Id.IsNull() || plan.Id.IsUnknown() {
		plan.DataQualityCheckId = types.StringValue(checkID)
		plan.Id = types.StringValue(checkID)
	}

	tflog.Debug(ctx, "Created evaluation", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *evaluationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_evaluation.EvaluationModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkID := state.DataQualityCheckId.ValueString()

	tflog.Debug(ctx, "Reading evaluation", map[string]interface{}{"data_quality_check_id": checkID})
	response, err := r.client.GetEvaluation(ctx, checkID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Evaluation not found, removing from state", map[string]interface{}{"data_quality_check_id": checkID})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading evaluation",
			"Could not read evaluation for check "+checkID+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *evaluationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This is a read-mostly resource. Update is a no-op that just re-reads state.
	var plan resource_evaluation.EvaluationModel
	var state resource_evaluation.EvaluationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkID := state.DataQualityCheckId.ValueString()

	response, err := r.client.GetEvaluation(ctx, checkID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading evaluation",
			"Could not read evaluation for check "+checkID+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response)

	if plan.Id.IsNull() || plan.Id.IsUnknown() {
		plan.DataQualityCheckId = types.StringValue(checkID)
		plan.Id = types.StringValue(checkID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *evaluationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Evaluations are not directly deletable; this is a no-op.
	tflog.Debug(ctx, "Delete evaluation: no-op (evaluations are managed server-side)")
}

func (r *evaluationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("data_quality_check_id"), req.ID)...)
}

func (r *evaluationResource) updateModelFromResponse(ctx context.Context, model *resource_evaluation.EvaluationModel, response map[string]interface{}) {
	if checkID, ok := response["dataQualityCheckId"].(string); ok {
		model.DataQualityCheckId = types.StringValue(checkID)
		model.Id = types.StringValue(checkID)
	}
	if catalogId, ok := response["catalogId"].(string); ok {
		model.CatalogId = types.StringValue(catalogId)
	} else {
		model.CatalogId = types.StringNull()
	}
	if category, ok := response["category"].(string); ok {
		model.Category = types.StringValue(category)
	} else {
		model.Category = types.StringNull()
	}
	// clusterId is request-only; the evaluation response never carries it back.
	if clusterId, ok := response["clusterId"].(string); ok {
		model.ClusterId = types.StringValue(clusterId)
	} else if model.ClusterId.IsUnknown() {
		model.ClusterId = types.StringNull()
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else {
		model.Description = types.StringNull()
	}
	if kind, ok := response["kind"].(string); ok {
		model.Kind = types.StringValue(kind)
	} else {
		model.Kind = types.StringNull()
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	} else {
		model.Name = types.StringNull()
	}
	if query, ok := response["query"].(string); ok {
		model.Query = types.StringValue(query)
	} else {
		model.Query = types.StringNull()
	}
	if schemaId, ok := response["schemaId"].(string); ok {
		model.SchemaId = types.StringValue(schemaId)
	} else {
		model.SchemaId = types.StringNull()
	}
	if severity, ok := response["severity"].(string); ok {
		model.Severity = types.StringValue(severity)
	} else {
		model.Severity = types.StringNull()
	}
	if tableId, ok := response["tableId"].(string); ok {
		model.TableId = types.StringValue(tableId)
	} else {
		model.TableId = types.StringNull()
	}

	// Map evaluations list
	elementType := resource_evaluation.EvaluationsType{
		ObjectType: types.ObjectType{
			AttrTypes: resource_evaluation.EvaluationsValue{}.AttributeTypes(ctx),
		},
	}
	evalAttributeTypes := resource_evaluation.EvaluationsValue{}.AttributeTypes(ctx)
	if evals, ok := response["evaluations"].([]interface{}); ok && len(evals) > 0 {
		var evalElements []attr.Value
		for _, evalRaw := range evals {
			if evalMap, ok := evalRaw.(map[string]interface{}); ok {
				evalAttributes := map[string]attr.Value{}
				if v, ok := evalMap["basedOnStatsAt"].(string); ok {
					evalAttributes["based_on_stats_at"] = types.StringValue(v)
				} else {
					evalAttributes["based_on_stats_at"] = types.StringNull()
				}
				if v, ok := evalMap["evaluatedAt"].(string); ok {
					evalAttributes["evaluated_at"] = types.StringValue(v)
				} else {
					evalAttributes["evaluated_at"] = types.StringNull()
				}
				if v, ok := evalMap["predicate"].(string); ok {
					evalAttributes["predicate"] = types.StringValue(v)
				} else {
					evalAttributes["predicate"] = types.StringNull()
				}
				if v, ok := evalMap["status"].(string); ok {
					evalAttributes["status"] = types.StringValue(v)
				} else {
					evalAttributes["status"] = types.StringNull()
				}
				evalValue, diags := resource_evaluation.NewEvaluationsValue(evalAttributeTypes, evalAttributes)
				if diags.HasError() {
					continue
				}
				evalElements = append(evalElements, evalValue)
			}
		}
		model.Evaluations, _ = types.ListValue(elementType, evalElements)
	} else {
		model.Evaluations, _ = types.ListValue(elementType, []attr.Value{})
	}
}
