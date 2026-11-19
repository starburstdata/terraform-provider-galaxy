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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_kafka_ingest_source"
)

var _ resource.Resource = (*kafkaIngestSourceResource)(nil)
var _ resource.ResourceWithConfigure = (*kafkaIngestSourceResource)(nil)
var _ resource.ResourceWithImportState = (*kafkaIngestSourceResource)(nil)

func NewKafkaIngestSourceResource() resource.Resource {
	return &kafkaIngestSourceResource{}
}

type kafkaIngestSourceResource struct {
	client *client.GalaxyClient
}

func (r *kafkaIngestSourceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_ingest_source"
}

func (r *kafkaIngestSourceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_kafka_ingest_source.KafkaIngestSourceResourceSchema(ctx)

	if attr, ok := s.Attributes["kafka_ingest_source_id"].(schema.StringAttribute); ok {
		attr.PlanModifiers = []planmodifier.String{stringplanmodifier.UseStateForUnknown()}
		s.Attributes["kafka_ingest_source_id"] = attr
	}

	resp.Schema = s
}

func (r *kafkaIngestSourceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *kafkaIngestSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_kafka_ingest_source.KafkaIngestSourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// password is WriteOnly: read from req.Config.
	var config resource_kafka_ingest_source.KafkaIngestSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.KafkaAuth.Password = config.KafkaAuth.Password

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating kafka_ingest_source")
	response, err := r.client.CreateKafkaIngestSource(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating kafka_ingest_source",
			"Could not create kafka ingest source: "+err.Error(),
		)
		return
	}

	r.unknownToNull(&plan)
	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	// Clear WriteOnly password before writing state; Plugin Framework nullifies this value in
	// plan, so state must match.
	plan.KafkaAuth.Password = types.StringNull()

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *kafkaIngestSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_kafka_ingest_source.KafkaIngestSourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.KafkaIngestSourceId.ValueString()
	response, err := r.client.GetKafkaIngestSource(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading kafka_ingest_source",
			"Could not read kafka_ingest_source "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *kafkaIngestSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_kafka_ingest_source.KafkaIngestSourceModel
	var state resource_kafka_ingest_source.KafkaIngestSourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// password is WriteOnly: read from req.Config.
	var config resource_kafka_ingest_source.KafkaIngestSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.KafkaAuth.Password = config.KafkaAuth.Password

	id := state.KafkaIngestSourceId.ValueString()
	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.UpdateKafkaIngestSource(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating kafka_ingest_source",
			"Could not update kafka_ingest_source "+id+": "+err.Error(),
		)
		return
	}

	plan.KafkaIngestSourceId = state.KafkaIngestSourceId
	r.unknownToNull(&plan)
	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *kafkaIngestSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_kafka_ingest_source.KafkaIngestSourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.KafkaIngestSourceId.ValueString()
	if err := r.client.DeleteKafkaIngestSource(ctx, id); err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting kafka_ingest_source",
				"Could not delete kafka_ingest_source "+id+": "+err.Error(),
			)
			return
		}
	}
}

func (r *kafkaIngestSourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("kafka_ingest_source_id"), req, resp)
}

func (r *kafkaIngestSourceResource) modelToCreateRequest(ctx context.Context, model *resource_kafka_ingest_source.KafkaIngestSourceModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	if !model.Name.IsNull() && !model.Name.IsUnknown() && model.Name.ValueString() != "" {
		request["name"] = model.Name.ValueString()
	}
	if !model.Description.IsNull() && !model.Description.IsUnknown() && model.Description.ValueString() != "" {
		request["description"] = model.Description.ValueString()
	}
	if !model.CloudRegionId.IsNull() && !model.CloudRegionId.IsUnknown() && model.CloudRegionId.ValueString() != "" {
		request["cloudRegionId"] = model.CloudRegionId.ValueString()
	}

	if !model.KafkaBrokers.IsNull() && !model.KafkaBrokers.IsUnknown() {
		var brokers []string
		d := model.KafkaBrokers.ElementsAs(ctx, &brokers, false)
		diags.Append(d...)
		if !diags.HasError() {
			request["kafkaBrokers"] = brokers
		}
	}

	// Galaxy's Jackson deserializer for KafkaAuth uses "typeKey" as the discriminator, not the
	// "authType" the OpenAPI spec shows. Send both for safety.
	auth := map[string]interface{}{}
	if !model.KafkaAuth.AuthType.IsNull() && !model.KafkaAuth.AuthType.IsUnknown() && model.KafkaAuth.AuthType.ValueString() != "" {
		auth["authType"] = model.KafkaAuth.AuthType.ValueString()
		auth["typeKey"] = model.KafkaAuth.AuthType.ValueString()
	}
	if !model.KafkaAuth.AwsRoleArn.IsNull() && !model.KafkaAuth.AwsRoleArn.IsUnknown() && model.KafkaAuth.AwsRoleArn.ValueString() != "" {
		auth["awsRoleArn"] = model.KafkaAuth.AwsRoleArn.ValueString()
	}
	if !model.KafkaAuth.Password.IsNull() && !model.KafkaAuth.Password.IsUnknown() && model.KafkaAuth.Password.ValueString() != "" {
		auth["password"] = model.KafkaAuth.Password.ValueString()
	}
	if !model.KafkaAuth.ScramType.IsNull() && !model.KafkaAuth.ScramType.IsUnknown() && model.KafkaAuth.ScramType.ValueString() != "" {
		auth["scramType"] = model.KafkaAuth.ScramType.ValueString()
	}
	if !model.KafkaAuth.Username.IsNull() && !model.KafkaAuth.Username.IsUnknown() && model.KafkaAuth.Username.ValueString() != "" {
		auth["username"] = model.KafkaAuth.Username.ValueString()
	}
	request["kafkaAuth"] = auth

	return request
}

func (r *kafkaIngestSourceResource) unknownToNull(model *resource_kafka_ingest_source.KafkaIngestSourceModel) {
	if model.Description.IsUnknown() {
		model.Description = types.StringNull()
	}
	if model.CloudRegionId.IsUnknown() {
		model.CloudRegionId = types.StringNull()
	}
}

func (r *kafkaIngestSourceResource) updateModelFromResponse(ctx context.Context, model *resource_kafka_ingest_source.KafkaIngestSourceModel, response map[string]interface{}, diags *diag.Diagnostics) {
	if id, ok := response["kafkaIngestSourceId"].(string); ok {
		model.KafkaIngestSourceId = types.StringValue(id)
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
	if cloudRegionId, ok := response["cloudRegionId"].(string); ok {
		model.CloudRegionId = types.StringValue(cloudRegionId)
	} else if model.CloudRegionId.IsUnknown() {
		model.CloudRegionId = types.StringNull()
	}

	if brokers, ok := response["kafkaBrokers"].([]interface{}); ok {
		elems := make([]attr.Value, 0, len(brokers))
		for _, b := range brokers {
			if s, ok := b.(string); ok {
				elems = append(elems, types.StringValue(s))
			}
		}
		listValue, d := types.ListValue(types.StringType, elems)
		diags.Append(d...)
		model.KafkaBrokers = listValue
	} else if model.KafkaBrokers.IsUnknown() {
		listValue, d := types.ListValue(types.StringType, []attr.Value{})
		diags.Append(d...)
		model.KafkaBrokers = listValue
	}

	// kafka_auth: start from model values so non-Computed fields (auth_type) are preserved from
	// the plan; the API response only echoes a subset. Password is WriteOnly and left null.
	authAttrs := map[string]attr.Value{
		"auth_type":    model.KafkaAuth.AuthType,
		"aws_role_arn": types.StringNull(),
		"password":     types.StringNull(),
		"scram_type":   types.StringNull(),
		"username":     types.StringNull(),
	}
	if model.KafkaAuth.AuthType.IsUnknown() {
		authAttrs["auth_type"] = types.StringNull()
	}
	if !model.KafkaAuth.ScramType.IsNull() && !model.KafkaAuth.ScramType.IsUnknown() {
		authAttrs["scram_type"] = model.KafkaAuth.ScramType
	}
	if !model.KafkaAuth.Username.IsNull() && !model.KafkaAuth.Username.IsUnknown() {
		authAttrs["username"] = model.KafkaAuth.Username
	}
	if !model.KafkaAuth.AwsRoleArn.IsNull() && !model.KafkaAuth.AwsRoleArn.IsUnknown() {
		authAttrs["aws_role_arn"] = model.KafkaAuth.AwsRoleArn
	}
	if authMap, ok := response["kafkaAuth"].(map[string]interface{}); ok {
		if v, ok := authMap["authType"].(string); ok {
			authAttrs["auth_type"] = types.StringValue(v)
		}
		if v, ok := authMap["awsRoleArn"].(string); ok {
			authAttrs["aws_role_arn"] = types.StringValue(v)
		}
		if v, ok := authMap["scramType"].(string); ok {
			authAttrs["scram_type"] = types.StringValue(v)
		}
		if v, ok := authMap["username"].(string); ok {
			authAttrs["username"] = types.StringValue(v)
		}
	}
	authValue, d := resource_kafka_ingest_source.NewKafkaAuthValue(resource_kafka_ingest_source.KafkaAuthValue{}.AttributeTypes(ctx), authAttrs)
	diags.Append(d...)
	model.KafkaAuth = authValue
}
