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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_cluster"
)

var _ resource.Resource = (*clusterResource)(nil)

func NewClusterResource() resource.Resource {
	return &clusterResource{}
}

type clusterResource struct {
	client *client.GalaxyClient
}

func (r *clusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *clusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_cluster.ClusterResourceSchema(ctx)
}

func (r *clusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_cluster.ClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Apply edge case logic before creating
	r.applyEdgeCaseLogic(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert plan to API request
	clusterRequest := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Debug log the full request
	if debugJSON, err := json.Marshal(clusterRequest); err == nil {
		tflog.Info(ctx, "Creating cluster with request: "+string(debugJSON))
	}

	tflog.Debug(ctx, "Creating cluster", map[string]interface{}{
		"name": plan.Name.ValueString(),
	})

	// Create cluster via API
	clusterResp, err := r.client.CreateCluster(ctx, clusterRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating cluster",
			"Could not create cluster, unexpected error: "+err.Error(),
		)
		return
	}

	// Update plan with response data
	r.updateModelFromResponse(ctx, &plan, clusterResp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created cluster", map[string]interface{}{
		"cluster_id": plan.ClusterId.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_cluster.ClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.ClusterId.ValueString()

	tflog.Debug(ctx, "Reading cluster", map[string]interface{}{
		"cluster_id": clusterID,
	})

	// Read cluster via API
	clusterResp, err := r.client.GetCluster(ctx, clusterID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Cluster not found, removing from state", map[string]interface{}{
				"cluster_id": clusterID,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading cluster",
			"Could not read cluster "+clusterID+": "+err.Error(),
		)
		return
	}

	// Update state with response data - for Read operation, use the standard update
	r.updateModelFromResponse(ctx, &state, clusterResp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_cluster.ClusterModel
	var state resource_cluster.ClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Apply edge case logic before updating
	r.applyEdgeCaseLogic(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.ClusterId.ValueString()

	// Convert plan to API request
	updateRequest := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating cluster", map[string]interface{}{
		"cluster_id": clusterID,
	})

	// Update cluster via API
	clusterResp, err := r.client.UpdateCluster(ctx, clusterID, updateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating cluster",
			"Could not update cluster "+clusterID+": "+err.Error(),
		)
		return
	}

	// Update plan with response data
	r.updateModelFromResponse(ctx, &plan, clusterResp, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated cluster", map[string]interface{}{
		"cluster_id": plan.ClusterId.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_cluster.ClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.ClusterId.ValueString()

	tflog.Debug(ctx, "Deleting cluster", map[string]interface{}{
		"cluster_id": clusterID,
	})

	// Delete cluster via API
	err := r.client.DeleteCluster(ctx, clusterID)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting cluster",
				"Could not delete cluster "+clusterID+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted cluster", map[string]interface{}{
		"cluster_id": clusterID,
	})
}

// applyEdgeCaseLogic applies business logic edge cases to the cluster model
func (r *clusterResource) applyEdgeCaseLogic(model *resource_cluster.ClusterModel, diags *diag.Diagnostics) {
	// Edge case: WarpResiliencyEnabled should always be set to true if processing_mode includes "WarpSpeed"
	if !model.ProcessingMode.IsNull() && strings.Contains(model.ProcessingMode.ValueString(), "WarpSpeed") {
		model.WarpResiliencyEnabled = types.BoolValue(true)
	}

	// Edge case: result_cache_default_visibility_seconds should only be set if result_cache_enabled is true
	if !model.ResultCacheEnabled.IsNull() && !model.ResultCacheEnabled.ValueBool() {
		if !model.ResultCacheDefaultVisibilitySeconds.IsNull() && !model.ResultCacheDefaultVisibilitySeconds.IsUnknown() {
			// Reset the value if result cache is disabled
			model.ResultCacheDefaultVisibilitySeconds = types.Int64Null()
		}
	}
}

// modelToCreateRequest converts the Terraform model to API create request
func (r *clusterResource) modelToCreateRequest(ctx context.Context, model *resource_cluster.ClusterModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	if !model.Name.IsNull() {
		request["name"] = model.Name.ValueString()
	}
	if !model.CloudRegionId.IsNull() {
		request["cloudRegionId"] = model.CloudRegionId.ValueString()
	}
	if !model.MinWorkers.IsNull() {
		request["minWorkers"] = model.MinWorkers.ValueInt64()
	}
	if !model.MaxWorkers.IsNull() {
		request["maxWorkers"] = model.MaxWorkers.ValueInt64()
	}
	if !model.PrivateLinkCluster.IsNull() {
		request["privateLinkCluster"] = model.PrivateLinkCluster.ValueBool()
	}
	if !model.ResultCacheEnabled.IsNull() {
		request["resultCacheEnabled"] = model.ResultCacheEnabled.ValueBool()
	}
	if !model.WarpResiliencyEnabled.IsNull() {
		request["warpResiliencyEnabled"] = model.WarpResiliencyEnabled.ValueBool()
	}

	// Optional fields
	if !model.IdleStopMinutes.IsNull() && !model.IdleStopMinutes.IsUnknown() {
		request["idleStopMinutes"] = model.IdleStopMinutes.ValueInt64()
	}
	if !model.ProcessingMode.IsNull() && !model.ProcessingMode.IsUnknown() && model.ProcessingMode.ValueString() != "" {
		request["processingMode"] = model.ProcessingMode.ValueString()
	}
	if !model.Replicas.IsNull() && !model.Replicas.IsUnknown() && model.Replicas.ValueInt64() > 0 {
		request["replicas"] = model.Replicas.ValueInt64()
	}
	if !model.ResultCacheDefaultVisibilitySeconds.IsNull() && !model.ResultCacheDefaultVisibilitySeconds.IsUnknown() {
		request["resultCacheDefaultVisibilitySeconds"] = model.ResultCacheDefaultVisibilitySeconds.ValueInt64()
	}

	// Handle catalog references
	if !model.CatalogRefs.IsNull() && !model.CatalogRefs.IsUnknown() {
		catalogRefs := make([]string, 0)
		elements := make([]types.String, 0, len(model.CatalogRefs.Elements()))
		model.CatalogRefs.ElementsAs(ctx, &elements, false)
		for _, elem := range elements {
			if !elem.IsNull() && !elem.IsUnknown() {
				catalogRefs = append(catalogRefs, elem.ValueString())
			}
		}
		request["catalogRefs"] = catalogRefs
	} else {
		// catalogRefs is required, so provide an empty array if not set
		request["catalogRefs"] = []string{}
	}

	return request
}

// modelToUpdateRequest converts the Terraform model to API update request
func (r *clusterResource) modelToUpdateRequest(ctx context.Context, model *resource_cluster.ClusterModel, diags *diag.Diagnostics) map[string]interface{} {
	// For update, we use the same structure as create but exclude computed fields
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

// updateModelFromResponse updates the Terraform model with API response data
func (r *clusterResource) updateModelFromResponse(ctx context.Context, model *resource_cluster.ClusterModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Set computed fields from response
	if clusterId, ok := response["clusterId"].(string); ok {
		model.ClusterId = types.StringValue(clusterId)
	}

	if clusterState, ok := response["clusterState"].(string); ok {
		model.ClusterState = types.StringValue(clusterState)
	}

	if batchCluster, ok := response["batchCluster"].(bool); ok {
		model.BatchCluster = types.BoolValue(batchCluster)
	} else {
		// Set to false if not present in response
		model.BatchCluster = types.BoolValue(false)
	}

	if warpSpeedCluster, ok := response["warpSpeedCluster"].(bool); ok {
		model.WarpSpeedCluster = types.BoolValue(warpSpeedCluster)
	} else {
		// Set to false if not present in response
		model.WarpSpeedCluster = types.BoolValue(false)
	}

	// Edge case: trino_uri field remains unknown after apply when cluster is in DISABLED state. Only set when ENABLED.
	if trinoUri, ok := response["trinoUri"].(string); ok {
		if model.ClusterState.ValueString() == "ENABLED" {
			model.TrinoUri = types.StringValue(trinoUri)
		} else {
			// Set to null when cluster is not enabled
			model.TrinoUri = types.StringNull()
		}
	} else {
		// Set to null if not present in response
		model.TrinoUri = types.StringNull()
	}

	// Update other fields that might have changed
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}

	if cloudRegionId, ok := response["cloudRegionId"].(string); ok {
		model.CloudRegionId = types.StringValue(cloudRegionId)
	}

	if minWorkers, ok := response["minWorkers"].(float64); ok {
		model.MinWorkers = types.Int64Value(int64(minWorkers))
	}

	if maxWorkers, ok := response["maxWorkers"].(float64); ok {
		model.MaxWorkers = types.Int64Value(int64(maxWorkers))
	}

	if idleStopMinutes, ok := response["idleStopMinutes"].(float64); ok {
		model.IdleStopMinutes = types.Int64Value(int64(idleStopMinutes))
	}

	if privateLinkCluster, ok := response["privateLinkCluster"].(bool); ok {
		model.PrivateLinkCluster = types.BoolValue(privateLinkCluster)
	}

	if processingMode, ok := response["processingMode"].(string); ok {
		model.ProcessingMode = types.StringValue(processingMode)
	} else {
		// If processingMode is not in response, check if we have a known value in the model
		if model.ProcessingMode.IsUnknown() {
			// Set to null if it was unknown (not specified by user)
			model.ProcessingMode = types.StringNull()
		}
		// Otherwise keep the current plan value (e.g., "WarpSpeed" specified by user)
	}

	if resultCacheEnabled, ok := response["resultCacheEnabled"].(bool); ok {
		model.ResultCacheEnabled = types.BoolValue(resultCacheEnabled)
	}

	if resultCacheDefaultVisibilitySeconds, ok := response["resultCacheDefaultVisibilitySeconds"].(float64); ok {
		model.ResultCacheDefaultVisibilitySeconds = types.Int64Value(int64(resultCacheDefaultVisibilitySeconds))
	} else {
		// Preserve the configured value if the API doesn't return this field
		// This handles the case where the user specifies the value but the API
		// doesn't echo it back in the response
		if model.ResultCacheDefaultVisibilitySeconds.IsUnknown() {
			model.ResultCacheDefaultVisibilitySeconds = types.Int64Null()
		}
		// Otherwise keep the existing model value (user's configured value)
	}

	if warpResiliencyEnabled, ok := response["warpResiliencyEnabled"].(bool); ok {
		model.WarpResiliencyEnabled = types.BoolValue(warpResiliencyEnabled)
	}

	if replicas, ok := response["replicas"].(float64); ok {
		model.Replicas = types.Int64Value(int64(replicas))
	}

	// Handle catalog references - preserve original plan order to avoid consistency errors
	if catalogRefs, ok := response["catalogRefs"].([]interface{}); ok {
		// Convert API response to a map for lookup
		responseCatalogs := make(map[string]bool)
		for _, ref := range catalogRefs {
			if refStr, ok := ref.(string); ok {
				responseCatalogs[refStr] = true
			}
		}

		// Preserve the original order from the plan/model
		// Only update if we have the same catalogs (validation that all planned catalogs exist)
		if !model.CatalogRefs.IsNull() && !model.CatalogRefs.IsUnknown() {
			planElements := make([]types.String, 0, len(model.CatalogRefs.Elements()))
			model.CatalogRefs.ElementsAs(ctx, &planElements, false)

			// Verify all plan elements exist in response
			allFound := true
			for _, elem := range planElements {
				if !elem.IsNull() && !elem.IsUnknown() {
					if !responseCatalogs[elem.ValueString()] {
						allFound = false
						break
					}
				}
			}

			// If all plan catalogs are found in response, keep the original plan order
			if allFound {
				tflog.Info(ctx, "CATALOG_REFS: Preserving plan order to maintain consistency")
				// Keep the existing model.CatalogRefs unchanged to preserve order
			} else {
				tflog.Warn(ctx, "CATALOG_REFS: Plan and response mismatch, using response order")
				// Fall back to response order if there's a mismatch
				catalogRefElements := make([]types.String, len(catalogRefs))
				for i, ref := range catalogRefs {
					if refStr, ok := ref.(string); ok {
						catalogRefElements[i] = types.StringValue(refStr)
					}
				}
				listValue, listDiag := types.ListValueFrom(ctx, types.StringType, catalogRefElements)
				diags.Append(listDiag...)
				if !listDiag.HasError() {
					model.CatalogRefs = listValue
				}
			}
		} else {
			// No plan data, use response order
			catalogRefElements := make([]types.String, len(catalogRefs))
			for i, ref := range catalogRefs {
				if refStr, ok := ref.(string); ok {
					catalogRefElements[i] = types.StringValue(refStr)
				}
			}
			listValue, listDiag := types.ListValueFrom(ctx, types.StringType, catalogRefElements)
			diags.Append(listDiag...)
			if !listDiag.HasError() {
				model.CatalogRefs = listValue
			}
		}
	}
}
