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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_data_quality_schedule"
)

var _ resource.Resource = (*dataQualityScheduleResource)(nil)
var _ resource.ResourceWithConfigure = (*dataQualityScheduleResource)(nil)
var _ resource.ResourceWithImportState = (*dataQualityScheduleResource)(nil)

func NewDataQualityScheduleResource() resource.Resource {
	return &dataQualityScheduleResource{}
}

type dataQualityScheduleResource struct {
	client *client.GalaxyClient
}

func (r *dataQualityScheduleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_quality_schedule"
}

func (r *dataQualityScheduleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_data_quality_schedule.DataQualityScheduleResourceSchema(ctx)

	// catalog_id, schema_id, table_id are path params - make required + RequiresReplace
	for _, name := range []string{"catalog_id", "schema_id", "table_id"} {
		if attr, ok := s.Attributes[name].(schema.StringAttribute); ok {
			attr.Required = true
			attr.Optional = false
			attr.Computed = false
			attr.PlanModifiers = []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			}
			s.Attributes[name] = attr
		}
	}

	// data_quality_schedule_id is assigned at creation and never changes
	if attr, ok := s.Attributes["data_quality_schedule_id"].(schema.StringAttribute); ok {
		attr.PlanModifiers = []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		}
		s.Attributes["data_quality_schedule_id"] = attr
	}

	resp.Schema = s
}

func (r *dataQualityScheduleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *dataQualityScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_data_quality_schedule.DataQualityScheduleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := plan.CatalogId.ValueString()
	schemaID := plan.SchemaId.ValueString()
	tableID := plan.TableId.ValueString()

	request := map[string]interface{}{
		"clusterId":      plan.ClusterId.ValueString(),
		"cronExpression": plan.CronExpression.ValueString(),
		"roleId":         plan.RoleId.ValueString(),
	}
	if !plan.Timezone.IsNull() && !plan.Timezone.IsUnknown() && plan.Timezone.ValueString() != "" {
		request["timezone"] = plan.Timezone.ValueString()
	}

	tflog.Debug(ctx, "Creating data quality schedule", map[string]interface{}{"catalog_id": catalogID, "schema_id": schemaID, "table_id": tableID})
	response, err := r.client.CreateDataQualitySchedule(ctx, catalogID, schemaID, tableID, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating data quality schedule",
			"Could not create data quality schedule: "+err.Error()+clusterPermissionErrorHint(err),
		)
		return
	}

	// Set defaults for computed fields
	if plan.Enabled.IsUnknown() {
		plan.Enabled = types.BoolNull()
	}
	if plan.NextExecution.IsUnknown() {
		plan.NextExecution = types.StringNull()
	}
	if plan.Timezone.IsUnknown() {
		plan.Timezone = types.StringNull()
	}

	r.updateModelFromResponse(ctx, &plan, response)

	tflog.Debug(ctx, "Created data quality schedule", map[string]interface{}{"id": plan.DataQualityScheduleId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dataQualityScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_data_quality_schedule.DataQualityScheduleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := state.CatalogId.ValueString()
	schemaID := state.SchemaId.ValueString()
	tableID := state.TableId.ValueString()

	tflog.Debug(ctx, "Reading data quality schedule", map[string]interface{}{"catalog_id": catalogID, "schema_id": schemaID, "table_id": tableID})
	response, err := r.client.GetDataQualitySchedule(ctx, catalogID, schemaID, tableID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Data quality schedule not found, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading data quality schedule",
			"Could not read data quality schedule: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *dataQualityScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_data_quality_schedule.DataQualityScheduleModel
	var state resource_data_quality_schedule.DataQualityScheduleModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := state.CatalogId.ValueString()
	schemaID := state.SchemaId.ValueString()
	tableID := state.TableId.ValueString()
	scheduleID := state.DataQualityScheduleId.ValueString()

	request := map[string]interface{}{
		"clusterId":      plan.ClusterId.ValueString(),
		"cronExpression": plan.CronExpression.ValueString(),
		"roleId":         plan.RoleId.ValueString(),
	}
	if !plan.Timezone.IsNull() && !plan.Timezone.IsUnknown() && plan.Timezone.ValueString() != "" {
		request["timezone"] = plan.Timezone.ValueString()
	}

	tflog.Debug(ctx, "Updating data quality schedule", map[string]interface{}{"id": scheduleID})
	response, err := r.client.UpdateDataQualitySchedule(ctx, catalogID, schemaID, tableID, scheduleID, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating data quality schedule",
			"Could not update data quality schedule "+scheduleID+": "+err.Error()+clusterPermissionErrorHint(err),
		)
		return
	}

	plan.CatalogId = state.CatalogId
	plan.SchemaId = state.SchemaId
	plan.TableId = state.TableId
	plan.DataQualityScheduleId = state.DataQualityScheduleId
	r.updateModelFromResponse(ctx, &plan, response)

	tflog.Debug(ctx, "Updated data quality schedule", map[string]interface{}{"id": plan.DataQualityScheduleId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *dataQualityScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_data_quality_schedule.DataQualityScheduleModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	catalogID := state.CatalogId.ValueString()
	schemaID := state.SchemaId.ValueString()
	tableID := state.TableId.ValueString()
	scheduleID := state.DataQualityScheduleId.ValueString()

	tflog.Debug(ctx, "Deleting data quality schedule", map[string]interface{}{"id": scheduleID})
	err := r.client.DeleteDataQualitySchedule(ctx, catalogID, schemaID, tableID, scheduleID)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting data quality schedule",
				"Could not delete data quality schedule "+scheduleID+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted data quality schedule", map[string]interface{}{"id": scheduleID})
}

func (r *dataQualityScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// clusterPermissionErrorHint returns an actionable suffix for the common permission error
// hit when role_id lacks USE_CLUSTER on cluster_id. Creating (or owning) a cluster only
// grants the caller CREATE_CLUSTER on it - it does NOT grant USE_CLUSTER to the role that
// will execute data quality checks/schedules there. That role needs USE_CLUSTER granted
// independently (e.g. via a galaxy_role_privilege_grant with entity_kind = "Cluster",
// privilege = "UseCluster"), or the Galaxy API rejects the request with a 403
// PERMISSION_DENIED "Operation not allowed: GET_CLUSTER" error.
func clusterPermissionErrorHint(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if strings.Contains(msg, "403") && strings.Contains(msg, "GET_CLUSTER") {
		return " (hint: the role_id used to execute this schedule needs the \"UseCluster\" privilege " +
			"granted on cluster_id - creating a cluster only grants the creator \"CreateCluster\", not " +
			"\"UseCluster\". Grant it with a galaxy_role_privilege_grant resource: entity_kind = \"Cluster\", " +
			"entity_id = <cluster_id>, privilege = \"UseCluster\", grant_kind = \"Allow\".)"
	}
	return ""
}

func (r *dataQualityScheduleResource) updateModelFromResponse(ctx context.Context, model *resource_data_quality_schedule.DataQualityScheduleModel, response map[string]interface{}) {
	if scheduleId, ok := response["dataQualityScheduleId"].(string); ok {
		model.DataQualityScheduleId = types.StringValue(scheduleId)
		model.Id = types.StringValue(scheduleId)
	}
	if catalogId, ok := response["catalogId"].(string); ok {
		model.CatalogId = types.StringValue(catalogId)
	}
	if schemaId, ok := response["schemaId"].(string); ok {
		model.SchemaId = types.StringValue(schemaId)
	}
	if tableId, ok := response["tableId"].(string); ok {
		model.TableId = types.StringValue(tableId)
	}
	if clusterId, ok := response["clusterId"].(string); ok {
		model.ClusterId = types.StringValue(clusterId)
	}
	if cronExpression, ok := response["cronExpression"].(string); ok {
		model.CronExpression = types.StringValue(cronExpression)
	}
	if roleId, ok := response["roleId"].(string); ok {
		model.RoleId = types.StringValue(roleId)
	}
	if timezone, ok := response["timezone"].(string); ok {
		model.Timezone = types.StringValue(timezone)
	} else {
		model.Timezone = types.StringNull()
	}
	if enabled, ok := response["enabled"].(bool); ok {
		model.Enabled = types.BoolValue(enabled)
	} else {
		model.Enabled = types.BoolNull()
	}
	if nextExecution, ok := response["nextExecution"].(string); ok {
		model.NextExecution = types.StringValue(nextExecution)
	} else {
		model.NextExecution = types.StringNull()
	}

	// Map data_quality_checks list
	elementType := resource_data_quality_schedule.DataQualityChecksType{
		ObjectType: types.ObjectType{
			AttrTypes: resource_data_quality_schedule.DataQualityChecksValue{}.AttributeTypes(ctx),
		},
	}
	attributeTypes := resource_data_quality_schedule.DataQualityChecksValue{}.AttributeTypes(ctx)
	if checks, ok := response["dataQualityChecks"].([]interface{}); ok && len(checks) > 0 {
		var checkElements []attr.Value
		for _, checkRaw := range checks {
			if checkMap, ok := checkRaw.(map[string]interface{}); ok {
				attributes := map[string]attr.Value{}
				if v, ok := checkMap["dataQualityCheckId"].(string); ok {
					attributes["data_quality_check_id"] = types.StringValue(v)
				} else {
					attributes["data_quality_check_id"] = types.StringNull()
				}
				if v, ok := checkMap["name"].(string); ok {
					attributes["name"] = types.StringValue(v)
				} else {
					attributes["name"] = types.StringNull()
				}
				checkValue, diags := resource_data_quality_schedule.NewDataQualityChecksValue(attributeTypes, attributes)
				if diags.HasError() {
					continue
				}
				checkElements = append(checkElements, checkValue)
			}
		}
		model.DataQualityChecks, _ = types.ListValue(elementType, checkElements)
	} else {
		model.DataQualityChecks, _ = types.ListValue(elementType, []attr.Value{})
	}
}
