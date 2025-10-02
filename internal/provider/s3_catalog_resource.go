package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_s3_catalog"
)

var _ resource.Resource = (*s3_catalogResource)(nil)
var _ resource.ResourceWithConfigure = (*s3_catalogResource)(nil)
var _ resource.ResourceWithModifyPlan = (*s3_catalogResource)(nil)

func NewS3CatalogResource() resource.Resource {
	return &s3_catalogResource{}
}

type s3_catalogResource struct {
	client *client.GalaxyClient
}

func (r *s3_catalogResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_catalog"
}

func (r *s3_catalogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_s3_catalog.S3CatalogResourceSchema(ctx)
}

func (r *s3_catalogResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *s3_catalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_s3_catalog.S3CatalogModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Initialize optional computed fields to null if not provided in config (before API call)
	if plan.AccessKey.IsUnknown() {
		plan.AccessKey = types.StringNull()
	}
	if plan.SecretKey.IsUnknown() {
		plan.SecretKey = types.StringNull()
	}
	if plan.GlueAccessKey.IsUnknown() {
		plan.GlueAccessKey = types.StringNull()
	}
	if plan.GlueRoleArn.IsUnknown() {
		plan.GlueRoleArn = types.StringNull()
	}
	if plan.GlueSecretKey.IsUnknown() {
		plan.GlueSecretKey = types.StringNull()
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating s3_catalog")
	response, err := r.client.CreateCatalog(ctx, "s3", request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating s3_catalog",
			"Could not create s3_catalog: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created s3_catalog", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *s3_catalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_s3_catalog.S3CatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Reading s3_catalog", map[string]interface{}{"id": id})
	response, err := r.client.GetCatalog(ctx, "s3", id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "S3Catalog not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading s3_catalog",
			"Could not read s3_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *s3_catalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_s3_catalog.S3CatalogModel
	var state resource_s3_catalog.S3CatalogModel

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

	tflog.Debug(ctx, "Updating s3_catalog", map[string]interface{}{"id": id})
	response, err := r.client.UpdateCatalog(ctx, "s3", id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating s3_catalog",
			"Could not update s3_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated s3_catalog", map[string]interface{}{"id": plan.Id.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *s3_catalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_s3_catalog.S3CatalogModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.Id.ValueString()
	tflog.Debug(ctx, "Deleting s3_catalog", map[string]interface{}{"id": id})
	err := r.client.DeleteCatalog(ctx, "s3", id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting s3_catalog",
				"Could not delete s3_catalog "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted s3_catalog", map[string]interface{}{"id": id})
}

// ModifyPlan ensures mutually exclusive credentials are enforced at plan time.
// If role_arn is set we forcibly null out access_key and secret_key so they are
// not sent in update requests even if they remain in prior state.
func (r *s3_catalogResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If plan is null (destroy) exit
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan resource_s3_catalog.S3CatalogModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle regular S3 authentication mutual exclusivity
	if !plan.RoleArn.IsNull() && plan.RoleArn.ValueString() != "" {
		// Always null keys if roleArn set
		if !plan.AccessKey.IsNull() || !plan.SecretKey.IsNull() {
			tflog.Debug(ctx, "Nulling access_key/secret_key because role_arn is set")
		}
		plan.AccessKey = types.StringNull()
		plan.SecretKey = types.StringNull()
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}
}

// Helper methods
func (r *s3_catalogResource) modelToCreateRequest(ctx context.Context, model *resource_s3_catalog.S3CatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	request["name"] = model.Name.ValueString()
	request["metastoreType"] = model.MetastoreType.ValueString()
	request["readOnly"] = model.ReadOnly.ValueBool()

	// Get metastore type for field mapping
	metastoreType := model.MetastoreType.ValueString()

	// Optional fields
	if !model.Description.IsNull() {
		request["description"] = model.Description.ValueString()
	}

	// Handle metastore type specific fields
	if metastoreType == "galaxy" || metastoreType == "glue" {
		// Fields for galaxy and glue metastores
		if !model.DefaultBucket.IsNull() {
			request["defaultBucket"] = model.DefaultBucket.ValueString()
		}
		if !model.DefaultDataLocation.IsNull() {
			request["defaultDataLocation"] = model.DefaultDataLocation.ValueString()
		}
	}

	// Region field for glue metastores
	if metastoreType == "glue" {
		if !model.Region.IsNull() {
			request["region"] = model.Region.ValueString()
		}
	}

	// Fields for hive metastores
	if metastoreType == "hive" {
		if !model.HiveMetastoreHost.IsNull() {
			request["hiveMetastoreHost"] = model.HiveMetastoreHost.ValueString()
		}
		if !model.HiveMetastorePort.IsNull() {
			request["hiveMetastorePort"] = model.HiveMetastorePort.ValueInt64()
		}
		if !model.SshTunnelId.IsNull() {
			request["sshTunnelId"] = model.SshTunnelId.ValueString()
		}

	}

	// Authentication handling - use base fields for all metastore types
	roleArnSet := !model.RoleArn.IsNull() && !model.RoleArn.IsUnknown() && model.RoleArn.ValueString() != ""
	if roleArnSet {
		// Warn user if they attempted to set both
		if !model.AccessKey.IsNull() || !model.SecretKey.IsNull() {
			diags.AddWarning(
				"Ignoring access_key and/or secret_key",
				"role_arn is set so access_key and secret_key are ignored. Remove them from configuration to silence this warning.",
			)
		}
		request["roleArn"] = model.RoleArn.ValueString()
		// Guarantee keys are not included
		delete(request, "accessKey")
		delete(request, "secretKey")
	} else {
		if !model.AccessKey.IsNull() {
			request["accessKey"] = model.AccessKey.ValueString()
		}
		if !model.SecretKey.IsNull() {
			request["secretKey"] = model.SecretKey.ValueString()
		}
	}

	if !model.DefaultTableFormat.IsNull() {
		request["defaultTableFormat"] = model.DefaultTableFormat.ValueString()
	}
	if !model.ExternalTableCreationEnabled.IsNull() {
		request["externalTableCreationEnabled"] = model.ExternalTableCreationEnabled.ValueBool()
	}
	if !model.ExternalTableWritesEnabled.IsNull() {
		request["externalTableWritesEnabled"] = model.ExternalTableWritesEnabled.ValueBool()
	}

	if !model.Validate.IsNull() && !model.Validate.IsUnknown() {
		request["validate"] = model.Validate.ValueBool()
	}

	return request
}

func (r *s3_catalogResource) modelToUpdateRequest(ctx context.Context, model *resource_s3_catalog.S3CatalogModel, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	// Safety: if both roleArn and any key fields somehow present, drop keys
	if _, hasRole := request["roleArn"]; hasRole {
		// Omit accessKey/secretKey entirely when roleArn is used.
		delete(request, "accessKey")
		delete(request, "secretKey")
	}

	return request
}

func (r *s3_catalogResource) updateModelFromResponse(ctx context.Context, model *resource_s3_catalog.S3CatalogModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Map response fields to model
	// Use catalogId as the ID for s3_catalog
	if catalogId, ok := response["catalogId"].(string); ok {
		model.Id = types.StringValue(catalogId)
		model.CatalogId = types.StringValue(catalogId)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	}

	// Get metastore type for field handling
	metastoreType := ""
	if mt, ok := response["metastoreType"].(string); ok {
		metastoreType = mt
		model.MetastoreType = types.StringValue(mt)
	}

	if readOnly, ok := response["readOnly"].(bool); ok {
		model.ReadOnly = types.BoolValue(readOnly)
	}

	// Handle discriminator fields based on metastore type
	if defaultBucket, ok := response["defaultBucket"].(string); ok {
		model.DefaultBucket = types.StringValue(defaultBucket)
	} else {
		model.DefaultBucket = types.StringNull()
	}
	if defaultDataLocation, ok := response["defaultDataLocation"].(string); ok {
		model.DefaultDataLocation = types.StringValue(defaultDataLocation)
	} else {
		model.DefaultDataLocation = types.StringNull()
	}

	// Glue-specific fields
	if metastoreType == "glue" {
		if region, ok := response["region"].(string); ok {
			model.Region = types.StringValue(region)
		} else {
			model.Region = types.StringNull()
		}
	} else {
		// Null out glue-specific fields for non-glue metastores
		model.Region = types.StringNull()
	}

	// Hive-specific fields
	if metastoreType == "hive" {
		if hiveMetastoreHost, ok := response["hiveMetastoreHost"].(string); ok {
			model.HiveMetastoreHost = types.StringValue(hiveMetastoreHost)
		} else {
			model.HiveMetastoreHost = types.StringNull()
		}
		if hiveMetastorePort, ok := response["hiveMetastorePort"].(float64); ok {
			model.HiveMetastorePort = types.Int64Value(int64(hiveMetastorePort))
		} else {
			model.HiveMetastorePort = types.Int64Null()
		}
		if sshTunnelId, ok := response["sshTunnelId"].(string); ok {
			model.SshTunnelId = types.StringValue(sshTunnelId)
		} else {
			model.SshTunnelId = types.StringNull()
		}
	} else {
		// Null out hive-specific fields for non-hive metastores
		model.HiveMetastoreHost = types.StringNull()
		model.HiveMetastorePort = types.Int64Null()
		model.SshTunnelId = types.StringNull()
	}

	// Handle authentication fields for all metastore types
	if accessKey, ok := response["accessKey"].(string); ok && accessKey != "" {
		model.AccessKey = types.StringValue(accessKey)
	} else {
		model.AccessKey = types.StringNull()
	}

	if secretKey, ok := response["secretKey"].(string); ok && secretKey != "" && secretKey != "<Value is encrypted>" {
		model.SecretKey = types.StringValue(secretKey)
	}

	if roleArn, ok := response["roleArn"].(string); ok {
		model.RoleArn = types.StringValue(roleArn)
		// Clear keys in state if roleArn now governs auth
		model.AccessKey = types.StringNull()
		model.SecretKey = types.StringNull()
	} else {
		model.RoleArn = types.StringNull()
	}

	if defaultTableFormat, ok := response["defaultTableFormat"].(string); ok {
		model.DefaultTableFormat = types.StringValue(defaultTableFormat)
	}
	if externalTableCreationEnabled, ok := response["externalTableCreationEnabled"].(bool); ok {
		model.ExternalTableCreationEnabled = types.BoolValue(externalTableCreationEnabled)
	}
	if externalTableWritesEnabled, ok := response["externalTableWritesEnabled"].(bool); ok {
		model.ExternalTableWritesEnabled = types.BoolValue(externalTableWritesEnabled)
	}

	// Handle glue-specific authentication fields - these should be null for most cases
	// as they represent separate authentication for the Glue service itself (not commonly used)
	if glueAccessKey, ok := response["glueAccessKey"].(string); ok && glueAccessKey != "" {
		model.GlueAccessKey = types.StringValue(glueAccessKey)
	} else {
		model.GlueAccessKey = types.StringNull()
	}

	if glueRoleArn, ok := response["glueRoleArn"].(string); ok && glueRoleArn != "" {
		model.GlueRoleArn = types.StringValue(glueRoleArn)
	} else {
		model.GlueRoleArn = types.StringNull()
	}

	if glueSecretKey, ok := response["glueSecretKey"].(string); ok && glueSecretKey != "" && glueSecretKey != "<Value is encrypted>" {
		model.GlueSecretKey = types.StringValue(glueSecretKey)
	} else {
		model.GlueSecretKey = types.StringNull()
	}

	// Handle validate field
	if validate, ok := response["validate"].(bool); ok {
		model.Validate = types.BoolValue(validate)
	} else if model.Validate.IsUnknown() {
		model.Validate = types.BoolNull()
	}

}
