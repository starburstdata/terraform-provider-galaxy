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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_mongodb_catalog"
)

var _ resource.Resource = (*mongodb_catalogResource)(nil)
var _ resource.ResourceWithConfigure = (*mongodb_catalogResource)(nil)
var _ resource.ResourceWithImportState = (*mongodb_catalogResource)(nil)

func NewMongodbCatalogResource() resource.Resource {
	return &mongodb_catalogResource{}
}

type mongodb_catalogResource struct {
	client *client.GalaxyClient
}

// Extended model with additional fields
type MongodbCatalogModelExtended struct {
	resource_mongodb_catalog.MongodbCatalogModel
	Host     types.String `tfsdk:"host"`
	Database types.String `tfsdk:"database"`
}

func (r *mongodb_catalogResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mongodb_catalog"
}

func (r *mongodb_catalogResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	baseSchema := resource_mongodb_catalog.MongodbCatalogResourceSchema(ctx)

	// Add missing fields for Mongodb connection
	baseSchema.Attributes["host"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		DeprecationMessage:  "Use hosts instead. This field will be removed in a future version.",
		Description:         "Mongodb host (required for direct and sshTunnel connection types)",
		MarkdownDescription: "Mongodb host (required for direct and sshTunnel connection types)",
	}
	baseSchema.Attributes["database"] = schema.StringAttribute{
		Optional:            true,
		Computed:            true,
		DeprecationMessage:  "Include the database in the hosts field (e.g., host:port/database). This field will be removed in a future version.",
		Description:         "Mongodb database",
		MarkdownDescription: "Mongodb database",
	}

	// Fix: validate is a request-only parameter, not returned by API.
	// Setting Computed=false ensures it's sent with update requests.
	if attr, ok := baseSchema.Attributes["validate"].(schema.BoolAttribute); ok {
		attr.Computed = false
		baseSchema.Attributes["validate"] = attr
	}

	resp.Schema = baseSchema
}

func (r *mongodb_catalogResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *mongodb_catalogResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MongodbCatalogModelExtended

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Initialize optional computed fields to null if not provided in config (before API call)
	if plan.Hosts.IsUnknown() {
		plan.Hosts = types.StringNull()
	}
	if plan.Host.IsUnknown() {
		plan.Host = types.StringNull()
	}
	if plan.Database.IsUnknown() {
		plan.Database = types.StringNull()
	}
	if plan.PrivateLinkId.IsUnknown() {
		plan.PrivateLinkId = types.StringNull()
	}
	if plan.SshTunnelId.IsUnknown() {
		plan.SshTunnelId = types.StringNull()
	}

	if plan.Password.IsNull() || plan.Password.IsUnknown() || plan.Password.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Missing required field",
			"password cannot be empty for MongoDB catalog",
		)
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating mongodb_catalog", map[string]interface{}{"request": request})
	response, err := r.client.CreateCatalog(ctx, "mongodb", request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating mongodb_catalog",
			"Could not create mongodb_catalog: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created mongodb_catalog", map[string]interface{}{"id": plan.CatalogId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mongodb_catalogResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MongodbCatalogModelExtended

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	tflog.Debug(ctx, "Reading mongodb_catalog", map[string]interface{}{"id": id})
	response, err := r.client.GetCatalog(ctx, "mongodb", id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "MongodbCatalog not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading mongodb_catalog",
			"Could not read mongodb_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *mongodb_catalogResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MongodbCatalogModelExtended
	var state MongodbCatalogModelExtended

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

	tflog.Debug(ctx, "Updating mongodb_catalog", map[string]interface{}{"id": id})
	response, err := r.client.UpdateCatalog(ctx, "mongodb", id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating mongodb_catalog",
			"Could not update mongodb_catalog "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated mongodb_catalog", map[string]interface{}{"id": plan.CatalogId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *mongodb_catalogResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MongodbCatalogModelExtended

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.CatalogId.ValueString()
	tflog.Debug(ctx, "Deleting mongodb_catalog", map[string]interface{}{"id": id})
	err := r.client.DeleteCatalog(ctx, "mongodb", id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting mongodb_catalog",
				"Could not delete mongodb_catalog "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted mongodb_catalog", map[string]interface{}{"id": id})
}

func (r *mongodb_catalogResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("catalog_id"), req, resp)
}

// Helper methods
func (r *mongodb_catalogResource) modelToCreateRequest(ctx context.Context, model *MongodbCatalogModelExtended, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	request["name"] = model.Name.ValueString()
	request["readOnly"] = model.ReadOnly.ValueBool()
	request["username"] = model.Username.ValueString()

	// password is write-only and not returned by the API. Omit on empty so PATCH
	// after import preserves the existing credential. Create enforces required. ENG-9975.
	if !model.Password.IsNull() && !model.Password.IsUnknown() && model.Password.ValueString() != "" {
		request["password"] = model.Password.ValueString()
	}

	// Determine connection type from provided fields
	hasHost := model.Host.ValueString() != ""
	hasHosts := model.Hosts.ValueString() != ""
	hasPrivateLinkId := model.PrivateLinkId.ValueString() != ""
	hasSshTunnelId := model.SshTunnelId.ValueString() != ""

	if model.ConnectionType.ValueString() != "" {
		request["connectionType"] = model.ConnectionType.ValueString()
	} else if hasPrivateLinkId {
		request["connectionType"] = "privateLink"
	} else if hasSshTunnelId {
		request["connectionType"] = "sshTunnel"
	} else {
		request["connectionType"] = "direct"
	}

	// Resolve hosts value from hosts field or deprecated host+database fields
	resolveHosts := func() (string, bool) {
		if hasHosts {
			return model.Hosts.ValueString(), true
		}
		if hasHost {
			h := model.Host.ValueString()
			if db := model.Database.ValueString(); db != "" {
				h = fmt.Sprintf("%s/%s", h, db)
			}
			return h, true
		}
		return "", false
	}

	connType := request["connectionType"].(string)
	switch connType {
	case "direct":
		if hosts, ok := resolveHosts(); ok {
			request["hosts"] = hosts
		} else {
			diags.AddError("Missing required field", "hosts (or deprecated host) is required when connection_type is direct for mongodb_catalog")
			return request
		}
	case "sshTunnel":
		if !hasSshTunnelId {
			diags.AddError("Missing required field", "ssh_tunnel_id is required when connection_type is sshTunnel for mongodb_catalog")
			return request
		}
		if hosts, ok := resolveHosts(); ok {
			request["hosts"] = hosts
		} else {
			diags.AddError("Missing required field", "hosts (or deprecated host) is required when connection_type is sshTunnel for mongodb_catalog")
			return request
		}
		request["sshTunnelId"] = model.SshTunnelId.ValueString()
	case "privateLink":
		if !hasPrivateLinkId {
			diags.AddError("Missing required field", "private_link_id is required when connection_type is privateLink for mongodb_catalog")
			return request
		}
		request["privateLinkId"] = model.PrivateLinkId.ValueString()
	default:
		diags.AddError("Invalid connection_type", fmt.Sprintf("connection_type must be one of: direct, sshTunnel, privateLink. Got: %s", connType))
		return request
	}

	if model.Description.ValueString() != "" {
		request["description"] = model.Description.ValueString()
	}

	if model.CloudKind.ValueString() != "" {
		request["cloudKind"] = model.CloudKind.ValueString()
	}

	// Handle regions field
	if !model.Regions.IsNull() && !model.Regions.IsUnknown() {
		var regions []string
		diags.Append(model.Regions.ElementsAs(ctx, &regions, false)...)
		if !diags.HasError() {
			request["regions"] = regions
		}
	}

	// Handle optional boolean fields
	if !model.DnsSeedListEnabled.IsNull() && !model.DnsSeedListEnabled.IsUnknown() {
		request["dnsSeedListEnabled"] = model.DnsSeedListEnabled.ValueBool()
	}

	if !model.FederatedDatabaseEnabled.IsNull() && !model.FederatedDatabaseEnabled.IsUnknown() {
		request["federatedDatabaseEnabled"] = model.FederatedDatabaseEnabled.ValueBool()
	}

	if !model.TlsEnabled.IsNull() && !model.TlsEnabled.IsUnknown() {
		request["tlsEnabled"] = model.TlsEnabled.ValueBool()
	}

	if !model.Validate.IsNull() && !model.Validate.IsUnknown() {
		request["validate"] = model.Validate.ValueBool()
	}

	return request
}

func (r *mongodb_catalogResource) modelToUpdateRequest(ctx context.Context, model *MongodbCatalogModelExtended, diags *diag.Diagnostics) map[string]interface{} {
	request := r.modelToCreateRequest(ctx, model, diags)

	return request
}

func (r *mongodb_catalogResource) updateModelFromResponse(ctx context.Context, model *MongodbCatalogModelExtended, response map[string]interface{}, diags *diag.Diagnostics) {
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

	// Parse hosts field back to host and database
	if hosts, ok := response["hosts"].(string); ok && hosts != "" {
		model.Hosts = types.StringValue(hosts)
		// Parse format: host:port/database or just host
		if idx := strings.LastIndex(hosts, "/"); idx != -1 {
			model.Host = types.StringValue(hosts[:idx])
			model.Database = types.StringValue(hosts[idx+1:])
		} else {
			model.Host = types.StringValue(hosts)
		}
	} else {
		if model.Hosts.IsUnknown() {
			model.Hosts = types.StringNull()
		}
		if model.Host.IsUnknown() {
			model.Host = types.StringNull()
		}
		if model.Database.IsUnknown() {
			model.Database = types.StringNull()
		}
	}

	// MongoDB doesn't have separate port field

	if username, ok := response["username"].(string); ok {
		model.Username = types.StringValue(username)
	}

	// Password is write-only, keep existing value

	if cloudKind, ok := response["cloudKind"].(string); ok {
		model.CloudKind = types.StringValue(cloudKind)
	} else if model.CloudKind.IsUnknown() {
		model.CloudKind = types.StringNull()
	}

	if connectionType, ok := response["connectionType"].(string); ok {
		model.ConnectionType = types.StringValue(connectionType)
	}

	// Handle regions field
	if regions, ok := response["regions"].([]interface{}); ok {
		regionsList := make([]types.String, len(regions))
		for i, r := range regions {
			if regionStr, ok := r.(string); ok {
				regionsList[i] = types.StringValue(regionStr)
			}
		}
		model.Regions, _ = types.ListValueFrom(ctx, types.StringType, regionsList)
	}

	// Handle optional boolean fields
	if dnsSeedListEnabled, ok := response["dnsSeedListEnabled"].(bool); ok {
		model.DnsSeedListEnabled = types.BoolValue(dnsSeedListEnabled)
	} else if model.DnsSeedListEnabled.IsUnknown() {
		model.DnsSeedListEnabled = types.BoolNull()
	}

	if federatedDatabaseEnabled, ok := response["federatedDatabaseEnabled"].(bool); ok {
		model.FederatedDatabaseEnabled = types.BoolValue(federatedDatabaseEnabled)
	} else if model.FederatedDatabaseEnabled.IsUnknown() {
		model.FederatedDatabaseEnabled = types.BoolNull()
	}

	if tlsEnabled, ok := response["tlsEnabled"].(bool); ok {
		model.TlsEnabled = types.BoolValue(tlsEnabled)
	} else if model.TlsEnabled.IsUnknown() {
		model.TlsEnabled = types.BoolNull()
	}

	// Handle optional connection-specific fields
	if privateLinkId, ok := response["privateLinkId"].(string); ok && privateLinkId != "" {
		model.PrivateLinkId = types.StringValue(privateLinkId)
	} else if model.PrivateLinkId.IsUnknown() {
		model.PrivateLinkId = types.StringNull()
	}

	if sshTunnelId, ok := response["sshTunnelId"].(string); ok && sshTunnelId != "" {
		model.SshTunnelId = types.StringValue(sshTunnelId)
	} else if model.SshTunnelId.IsUnknown() {
		model.SshTunnelId = types.StringNull()
	}

	// Handle validate field
	if validate, ok := response["validate"].(bool); ok {
		model.Validate = types.BoolValue(validate)
	} else if model.Validate.IsUnknown() {
		model.Validate = types.BoolNull()
	}

}
