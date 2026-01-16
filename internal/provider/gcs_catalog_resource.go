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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_gcs_catalog"
)

var _ resource.Resource = (*gcs_catalogResource)(nil)
var _ resource.ResourceWithConfigure = (*gcs_catalogResource)(nil)

func NewGcsCatalogResource() resource.Resource {
	return &gcs_catalogResource{}
}

type gcs_catalogResource struct {
	client *client.GalaxyClient
}

// No need for extended model - use base model directly

func (r *gcs_catalogResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gcs_catalog"
}

func (r *gcs_catalogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_gcs_catalog.GcsCatalogResourceSchema(ctx)

	// Fix: validate is a request-only parameter, not returned by API.
	// Setting Computed=false ensures it's sent with update requests.
	if attr, ok := s.Attributes["validate"].(schema.BoolAttribute); ok {
		attr.Computed = false
		s.Attributes["validate"] = attr
	}

	resp.Schema = s
}

func (r *gcs_catalogResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *gcs_catalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_gcs_catalog.GcsCatalogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Initialize optional computed fields to null if not provided in config (before API call)
	if plan.HiveMetastoreHost.IsUnknown() {
		plan.HiveMetastoreHost = types.StringNull()
	}
	if plan.HiveMetastorePort.IsUnknown() {
		plan.HiveMetastorePort = types.Int64Null()
	}
	if plan.SshTunnelId.IsUnknown() {
		plan.SshTunnelId = types.StringNull()
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating gcs_catalog")
	response, err := r.client.CreateCatalog(ctx, "gcs", request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating gcs_catalog",
			"Could not create gcs_catalog: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created gcs_catalog", map[string]interface{}{"id": plan.CatalogId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gcs_catalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_gcs_catalog.GcsCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	tflog.Debug(ctx, "Reading gcs_catalog", map[string]interface{}{"id": id})
	response, err := r.client.GetCatalog(ctx, "gcs", id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "GcsCatalog not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading gcs_catalog",
			"Could not read gcs_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *gcs_catalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_gcs_catalog.GcsCatalogModel
	var state resource_gcs_catalog.GcsCatalogModel

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

	tflog.Debug(ctx, "Updating gcs_catalog", map[string]interface{}{"id": id})
	response, err := r.client.UpdateCatalog(ctx, "gcs", id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating gcs_catalog",
			"Could not update gcs_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated gcs_catalog", map[string]interface{}{"id": plan.CatalogId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *gcs_catalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_gcs_catalog.GcsCatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	tflog.Debug(ctx, "Deleting gcs_catalog", map[string]interface{}{"id": id})
	err := r.client.DeleteCatalog(ctx, "gcs", id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting gcs_catalog",
				"Could not delete gcs_catalog "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted gcs_catalog", map[string]interface{}{"id": id})
}

// Helper methods
func (r *gcs_catalogResource) modelToCreateRequest(ctx context.Context, model *resource_gcs_catalog.GcsCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	request["name"] = model.Name.ValueString()
	request["readOnly"] = model.ReadOnly.ValueBool()
	request["metastoreType"] = model.MetastoreType.ValueString()
	request["credentialsKey"] = model.CredentialsKey.ValueString()

	// Fields required for galaxy metastore
	if model.MetastoreType.ValueString() == "galaxy" {
		if !model.DefaultBucket.IsNull() && !model.DefaultBucket.IsUnknown() && model.DefaultBucket.ValueString() != "" {
			request["defaultBucket"] = model.DefaultBucket.ValueString()
		}
		if !model.DefaultDataLocation.IsNull() && !model.DefaultDataLocation.IsUnknown() && model.DefaultDataLocation.ValueString() != "" {
			request["defaultDataLocation"] = model.DefaultDataLocation.ValueString()
		}
	}

	// Optional fields
	if !model.Description.IsNull() && !model.Description.IsUnknown() && model.Description.ValueString() != "" {
		request["description"] = model.Description.ValueString()
	}

	if !model.DefaultTableFormat.IsNull() && !model.DefaultTableFormat.IsUnknown() && model.DefaultTableFormat.ValueString() != "" {
		request["defaultTableFormat"] = model.DefaultTableFormat.ValueString()
	}

	if !model.ExternalTableCreationEnabled.IsNull() && !model.ExternalTableCreationEnabled.IsUnknown() {
		request["externalTableCreationEnabled"] = model.ExternalTableCreationEnabled.ValueBool()
	}

	if !model.ExternalTableWritesEnabled.IsNull() && !model.ExternalTableWritesEnabled.IsUnknown() {
		request["externalTableWritesEnabled"] = model.ExternalTableWritesEnabled.ValueBool()
	}

	if !model.Validate.IsNull() && !model.Validate.IsUnknown() {
		request["validate"] = model.Validate.ValueBool()
	}

	return request
}

func (r *gcs_catalogResource) modelToUpdateRequest(ctx context.Context, model *resource_gcs_catalog.GcsCatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *gcs_catalogResource) updateModelFromResponse(ctx context.Context, model *resource_gcs_catalog.GcsCatalogModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
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

	if metastoreType, ok := response["metastoreType"].(string); ok {
		model.MetastoreType = types.StringValue(metastoreType)
	}

	// CredentialsKey is write-only, keep existing value

	if defaultBucket, ok := response["defaultBucket"].(string); ok {
		model.DefaultBucket = types.StringValue(defaultBucket)
	} else if model.DefaultBucket.IsUnknown() {
		model.DefaultBucket = types.StringNull()
	}

	if defaultDataLocation, ok := response["defaultDataLocation"].(string); ok {
		model.DefaultDataLocation = types.StringValue(defaultDataLocation)
	} else if model.DefaultDataLocation.IsUnknown() {
		model.DefaultDataLocation = types.StringNull()
	}

	if defaultTableFormat, ok := response["defaultTableFormat"].(string); ok {
		model.DefaultTableFormat = types.StringValue(defaultTableFormat)
	} else if model.DefaultTableFormat.IsUnknown() {
		model.DefaultTableFormat = types.StringNull()
	}

	if externalTableCreationEnabled, ok := response["externalTableCreationEnabled"].(bool); ok {
		model.ExternalTableCreationEnabled = types.BoolValue(externalTableCreationEnabled)
	} else if model.ExternalTableCreationEnabled.IsUnknown() {
		model.ExternalTableCreationEnabled = types.BoolNull()
	}

	if externalTableWritesEnabled, ok := response["externalTableWritesEnabled"].(bool); ok {
		model.ExternalTableWritesEnabled = types.BoolValue(externalTableWritesEnabled)
	} else if model.ExternalTableWritesEnabled.IsUnknown() {
		model.ExternalTableWritesEnabled = types.BoolNull()
	}

	// Handle hive metastore fields - these should be null for GCS with galaxy metastore
	if hiveMetastoreHost, ok := response["hiveMetastoreHost"].(string); ok && hiveMetastoreHost != "" {
		model.HiveMetastoreHost = types.StringValue(hiveMetastoreHost)
	} else {
		model.HiveMetastoreHost = types.StringNull()
	}

	if hiveMetastorePort, ok := response["hiveMetastorePort"].(float64); ok {
		model.HiveMetastorePort = types.Int64Value(int64(hiveMetastorePort))
	} else {
		model.HiveMetastorePort = types.Int64Null()
	}

	if sshTunnelId, ok := response["sshTunnelId"].(string); ok && sshTunnelId != "" {
		model.SshTunnelId = types.StringValue(sshTunnelId)
	} else {
		model.SshTunnelId = types.StringNull()
	}

	// Handle validate field
	if validate, ok := response["validate"].(bool); ok {
		model.Validate = types.BoolValue(validate)
	} else if model.Validate.IsUnknown() {
		model.Validate = types.BoolNull()
	}

}
