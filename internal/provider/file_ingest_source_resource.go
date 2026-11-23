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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_file_ingest_source"
)

var _ resource.Resource = (*fileIngestSourceResource)(nil)
var _ resource.ResourceWithConfigure = (*fileIngestSourceResource)(nil)
var _ resource.ResourceWithImportState = (*fileIngestSourceResource)(nil)

func NewFileIngestSourceResource() resource.Resource {
	return &fileIngestSourceResource{}
}

// fileIngestSourceResourceModel adds the hand-written secret_access_key_version attribute to the generated model.
type fileIngestSourceResourceModel struct {
	resource_file_ingest_source.FileIngestSourceModel
	SecretAccessKeyVersion types.String `tfsdk:"secret_access_key_version"`
}

type fileIngestSourceResource struct {
	client *client.GalaxyClient
}

func (r *fileIngestSourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_ingest_source"
}

func (r *fileIngestSourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_file_ingest_source.FileIngestSourceResourceSchema(ctx)

	if attr, ok := s.Attributes["file_ingest_source_id"].(schema.StringAttribute); ok {
		attr.PlanModifiers = []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
		s.Attributes["file_ingest_source_id"] = attr
	}

	// secret_access_key is write-only: it lives in config only and never enters plan or state.
	if creds, ok := s.Attributes["cloud_storage_credentials"].(schema.SingleNestedAttribute); ok {
		if key, ok := creds.Attributes["secret_access_key"].(schema.StringAttribute); ok {
			key.Optional = true
			key.Computed = false
			key.Sensitive = true
			key.WriteOnly = true
			creds.Attributes["secret_access_key"] = key
			s.Attributes["cloud_storage_credentials"] = creds
		}
	}

	// Because secret_access_key is write-only, changing it alone never produces a diff. Changing
	// secret_access_key_version forces an Update, which sends the configured secret_access_key.
	s.Attributes["secret_access_key_version"] = schema.StringAttribute{
		Optional:            true,
		Description:         "Non-secret value that, when changed, triggers an update that sends the configured cloud_storage_credentials.secret_access_key. Use it to rotate the key.",
		MarkdownDescription: "Non-secret value that, when changed, triggers an update that sends the configured `cloud_storage_credentials.secret_access_key`. Use it to rotate the key.",
	}

	resp.Schema = s
}

func (r *fileIngestSourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.GalaxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.GalaxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	r.client = c
}

func (r *fileIngestSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fileIngestSourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// credentials_type is WriteOnly: read from req.Config.
	var config fileIngestSourceResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.CloudStorageCredentials.CredentialsType = config.CloudStorageCredentials.CredentialsType
	plan.CloudStorageCredentials.SecretAccessKey = config.CloudStorageCredentials.SecretAccessKey

	request := r.modelToCreateRequest(&plan.FileIngestSourceModel)

	tflog.Debug(ctx, "Creating file_ingest_source")
	response, err := r.client.CreateFileIngestSource(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating file_ingest_source",
			"Could not create file ingest source: "+err.Error(),
		)
		return
	}

	r.unknownToNull(&plan.FileIngestSourceModel)
	r.updateModelFromResponse(ctx, &plan.FileIngestSourceModel, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *fileIngestSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fileIngestSourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.FileIngestSourceId.ValueString()
	response, err := r.client.GetFileIngestSource(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading file_ingest_source",
			"Could not read file_ingest_source "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state.FileIngestSourceModel, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *fileIngestSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan fileIngestSourceResourceModel
	var state fileIngestSourceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// credentials_type is WriteOnly: read from req.Config.
	var config fileIngestSourceResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.CloudStorageCredentials.CredentialsType = config.CloudStorageCredentials.CredentialsType
	plan.CloudStorageCredentials.SecretAccessKey = config.CloudStorageCredentials.SecretAccessKey

	id := state.FileIngestSourceId.ValueString()
	request := r.modelToCreateRequest(&plan.FileIngestSourceModel)

	response, err := r.client.UpdateFileIngestSource(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating file_ingest_source",
			"Could not update file_ingest_source "+id+": "+err.Error(),
		)
		return
	}

	plan.FileIngestSourceId = state.FileIngestSourceId
	r.unknownToNull(&plan.FileIngestSourceModel)
	r.updateModelFromResponse(ctx, &plan.FileIngestSourceModel, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *fileIngestSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state fileIngestSourceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.FileIngestSourceId.ValueString()
	if err := r.client.DeleteFileIngestSource(ctx, id); err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting file_ingest_source",
				"Could not delete file_ingest_source "+id+": "+err.Error(),
			)
			return
		}
	}
}

func (r *fileIngestSourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("file_ingest_source_id"), req, resp)
}

func (r *fileIngestSourceResource) modelToCreateRequest(model *resource_file_ingest_source.FileIngestSourceModel) map[string]interface{} {
	request := make(map[string]interface{})

	if !model.Name.IsNull() && !model.Name.IsUnknown() && model.Name.ValueString() != "" {
		request["name"] = model.Name.ValueString()
	}
	if !model.Description.IsNull() && !model.Description.IsUnknown() && model.Description.ValueString() != "" {
		request["description"] = model.Description.ValueString()
	}
	if !model.Bucket.IsNull() && !model.Bucket.IsUnknown() && model.Bucket.ValueString() != "" {
		request["bucket"] = model.Bucket.ValueString()
	}
	if !model.BucketCloudRegionId.IsNull() && !model.BucketCloudRegionId.IsUnknown() && model.BucketCloudRegionId.ValueString() != "" {
		request["bucketCloudRegionId"] = model.BucketCloudRegionId.ValueString()
	}
	if !model.PathPrefix.IsNull() && !model.PathPrefix.IsUnknown() && model.PathPrefix.ValueString() != "" {
		request["pathPrefix"] = model.PathPrefix.ValueString()
	}
	if !model.SourceCloudRegionId.IsNull() && !model.SourceCloudRegionId.IsUnknown() && model.SourceCloudRegionId.ValueString() != "" {
		request["sourceCloudRegionId"] = model.SourceCloudRegionId.ValueString()
	}

	creds := map[string]interface{}{}
	if !model.CloudStorageCredentials.CredentialsType.IsNull() && !model.CloudStorageCredentials.CredentialsType.IsUnknown() && model.CloudStorageCredentials.CredentialsType.ValueString() != "" {
		creds["credentialsType"] = model.CloudStorageCredentials.CredentialsType.ValueString()
		creds["typeKey"] = model.CloudStorageCredentials.CredentialsType.ValueString()
	}
	if !model.CloudStorageCredentials.AccessKeyId.IsNull() && !model.CloudStorageCredentials.AccessKeyId.IsUnknown() && model.CloudStorageCredentials.AccessKeyId.ValueString() != "" {
		creds["accessKeyId"] = model.CloudStorageCredentials.AccessKeyId.ValueString()
	}
	if !model.CloudStorageCredentials.SecretAccessKey.IsNull() && !model.CloudStorageCredentials.SecretAccessKey.IsUnknown() && model.CloudStorageCredentials.SecretAccessKey.ValueString() != "" {
		creds["secretAccessKey"] = model.CloudStorageCredentials.SecretAccessKey.ValueString()
	}
	if !model.CloudStorageCredentials.ExternalRoleArn.IsNull() && !model.CloudStorageCredentials.ExternalRoleArn.IsUnknown() && model.CloudStorageCredentials.ExternalRoleArn.ValueString() != "" {
		creds["externalRoleArn"] = model.CloudStorageCredentials.ExternalRoleArn.ValueString()
	}
	request["cloudStorageCredentials"] = creds

	return request
}

func (r *fileIngestSourceResource) unknownToNull(model *resource_file_ingest_source.FileIngestSourceModel) {
	if model.Description.IsUnknown() {
		model.Description = types.StringNull()
	}
	if model.PathPrefix.IsUnknown() {
		model.PathPrefix = types.StringNull()
	}
	if model.Prefix.IsUnknown() {
		model.Prefix = types.StringNull()
	}
	if model.SourceCloudRegionId.IsUnknown() {
		model.SourceCloudRegionId = types.StringNull()
	}
}

func (r *fileIngestSourceResource) updateModelFromResponse(ctx context.Context, model *resource_file_ingest_source.FileIngestSourceModel, response map[string]interface{}, diags *diag.Diagnostics) {
	if id, ok := response["fileIngestSourceId"].(string); ok {
		model.FileIngestSourceId = types.StringValue(id)
		model.Id = types.StringValue(id)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else if model.Description.IsUnknown() {
		model.Description = types.StringNull()
	}
	if bucket, ok := response["bucket"].(string); ok {
		model.Bucket = types.StringValue(bucket)
	}
	if bucketCloudRegionId, ok := response["bucketCloudRegionId"].(string); ok {
		model.BucketCloudRegionId = types.StringValue(bucketCloudRegionId)
	}
	if pathPrefix, ok := response["pathPrefix"].(string); ok {
		model.PathPrefix = types.StringValue(pathPrefix)
	} else if model.PathPrefix.IsUnknown() {
		model.PathPrefix = types.StringNull()
	}
	if prefix, ok := response["prefix"].(string); ok {
		model.Prefix = types.StringValue(prefix)
	} else if model.Prefix.IsUnknown() {
		model.Prefix = types.StringNull()
	}
	if sourceCloudRegionId, ok := response["sourceCloudRegionId"].(string); ok {
		model.SourceCloudRegionId = types.StringValue(sourceCloudRegionId)
	} else if model.SourceCloudRegionId.IsUnknown() {
		model.SourceCloudRegionId = types.StringNull()
	}

	// credentials_type and secret_access_key are WriteOnly - must be null in state. Other fields come from model (plan) or response.
	credsAttrs := map[string]attr.Value{
		"access_key_id":     types.StringNull(),
		"credentials_type":  types.StringNull(),
		"external_role_arn": types.StringNull(),
		"secret_access_key": types.StringNull(),
	}
	if !model.CloudStorageCredentials.AccessKeyId.IsNull() && !model.CloudStorageCredentials.AccessKeyId.IsUnknown() {
		credsAttrs["access_key_id"] = model.CloudStorageCredentials.AccessKeyId
	}
	if !model.CloudStorageCredentials.ExternalRoleArn.IsNull() && !model.CloudStorageCredentials.ExternalRoleArn.IsUnknown() {
		credsAttrs["external_role_arn"] = model.CloudStorageCredentials.ExternalRoleArn
	}
	if credMap, ok := response["cloudStorageCredentials"].(map[string]interface{}); ok {
		if v, ok := credMap["accessKeyId"].(string); ok {
			credsAttrs["access_key_id"] = types.StringValue(v)
		}
		if v, ok := credMap["externalRoleArn"].(string); ok {
			credsAttrs["external_role_arn"] = types.StringValue(v)
		}
	}
	credsValue, d := resource_file_ingest_source.NewCloudStorageCredentialsValue(resource_file_ingest_source.CloudStorageCredentialsValue{}.AttributeTypes(ctx), credsAttrs)
	diags.Append(d...)
	model.CloudStorageCredentials = credsValue
}
