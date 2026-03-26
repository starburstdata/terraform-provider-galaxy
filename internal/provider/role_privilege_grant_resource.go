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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_role_privilege_grant"
)

var _ resource.Resource = (*role_privilege_grantResource)(nil)
var _ resource.ResourceWithConfigure = (*role_privilege_grantResource)(nil)
var _ resource.ResourceWithImportState = (*role_privilege_grantResource)(nil)

func NewRolePrivilegeGrantResource() resource.Resource {
	return &role_privilege_grantResource{}
}

type role_privilege_grantResource struct {
	client *client.GalaxyClient
}

func (r *role_privilege_grantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_privilege_grant"
}

func (r *role_privilege_grantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_role_privilege_grant.RolePrivilegeGrantResourceSchema(ctx)

	// list_all_privileges is a query parameter for the list API, not a resource property.
	// Override to Computed-only so users cannot set it.
	if attr, ok := s.Attributes["list_all_privileges"].(schema.BoolAttribute); ok {
		attr.Optional = false
		attr.Computed = true
		s.Attributes["list_all_privileges"] = attr
	}

	// Add RequiresReplace to identifying and required attributes since the API
	// does not support in-place updates. Terraform will destroy and recreate on any change.
	for _, name := range []string{"role_id", "entity_id", "entity_kind", "privilege", "grant_kind"} {
		if attr, ok := s.Attributes[name].(schema.StringAttribute); ok {
			attr.PlanModifiers = append(attr.PlanModifiers, stringplanmodifier.RequiresReplace())
			s.Attributes[name] = attr
		}
	}
	if attr, ok := s.Attributes["grant_option"].(schema.BoolAttribute); ok {
		attr.PlanModifiers = append(attr.PlanModifiers, boolplanmodifier.RequiresReplace())
		s.Attributes["grant_option"] = attr
	}

	// Optional+Computed scope fields use RequiresReplaceIfConfigured so they only
	// trigger replacement when the user explicitly sets them, not when computed.
	for _, name := range []string{"column_name", "schema_name", "table_name"} {
		if attr, ok := s.Attributes[name].(schema.StringAttribute); ok {
			attr.PlanModifiers = append(attr.PlanModifiers, stringplanmodifier.RequiresReplaceIfConfigured())
			s.Attributes[name] = attr
		}
	}

	resp.Schema = s
}

func (r *role_privilege_grantResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *role_privilege_grantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_role_privilege_grant.RolePrivilegeGrantModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating role_privilege_grant")
	response, err := r.client.CreateRolePrivilegeGrant(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating role_privilege_grant",
			"Could not create role_privilege_grant: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created role_privilege_grant", map[string]interface{}{"id": plan.EntityId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *role_privilege_grantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_role_privilege_grant.RolePrivilegeGrantModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleId := state.RoleId.ValueString()
	entityId := state.EntityId.ValueString()
	privilege := state.Privilege.ValueString()
	grantKind := state.GrantKind.ValueString()

	// Pass scope fields so FindRolePrivilegeGrant matches the correct grant
	// when multiple grants exist for the same entity/privilege/grantKind.
	schemaName := ""
	if !state.SchemaName.IsNull() && !state.SchemaName.IsUnknown() {
		schemaName = state.SchemaName.ValueString()
	}
	tableName := ""
	if !state.TableName.IsNull() && !state.TableName.IsUnknown() {
		tableName = state.TableName.ValueString()
	}
	columnName := ""
	if !state.ColumnName.IsNull() && !state.ColumnName.IsUnknown() {
		columnName = state.ColumnName.ValueString()
	}

	tflog.Debug(ctx, "Reading role_privilege_grant", map[string]interface{}{
		"roleId":    roleId,
		"entityId":  entityId,
		"privilege": privilege,
		"grantKind": grantKind,
	})

	grant, err := r.client.FindRolePrivilegeGrant(ctx, roleId, entityId, privilege, grantKind, schemaName, tableName, columnName)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading role privileges", err.Error())
		return
	}

	if grant == nil {
		tflog.Warn(ctx, "Role privilege grant not found, removing from state", map[string]interface{}{
			"roleId":    roleId,
			"entityId":  entityId,
			"privilege": privilege,
			"grantKind": grantKind,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	r.updateModelFromResponse(ctx, &state, grant, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *role_privilege_grantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes use RequiresReplace, so Update should never be called.
	// If it is, just persist the plan to state.
	var plan resource_role_privilege_grant.RolePrivilegeGrantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *role_privilege_grantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_role_privilege_grant.RolePrivilegeGrantModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the revoke request using the same approach as Update
	roleId := state.RoleId.ValueString()
	entityId := state.EntityId.ValueString()
	entityKind := state.EntityKind.ValueString()
	privilege := state.Privilege.ValueString()

	tflog.Debug(ctx, "Deleting role_privilege_grant", map[string]interface{}{
		"roleId":    roleId,
		"entityId":  entityId,
		"privilege": privilege,
	})

	// Create the revoke request
	revokeRequest := make(map[string]interface{})
	revokeRequest["entityId"] = entityId
	revokeRequest["entityKind"] = entityKind
	revokeRequest["privilege"] = privilege
	revokeRequest["revokeAction"] = "RemoveRoleGrant"

	// Include optional scope fields if set
	if !state.ColumnName.IsNull() && state.ColumnName.ValueString() != "" {
		revokeRequest["columnName"] = state.ColumnName.ValueString()
	}
	if !state.SchemaName.IsNull() && state.SchemaName.ValueString() != "" {
		revokeRequest["schemaName"] = state.SchemaName.ValueString()
	}
	if !state.TableName.IsNull() && state.TableName.ValueString() != "" {
		revokeRequest["tableName"] = state.TableName.ValueString()
	}

	err := r.client.RevokeRolePrivilege(ctx, roleId, revokeRequest)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting role_privilege_grant",
				"Could not delete role_privilege_grant: "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted role_privilege_grant", map[string]interface{}{
		"roleId":    roleId,
		"entityId":  entityId,
		"privilege": privilege,
	})
}

func (r *role_privilege_grantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import formats:
	//   basic:    role_id/entity_id/entity_kind/privilege/grant_kind
	//   extended: role_id/entity_id/entity_kind/privilege/grant_kind/schema_name/table_name/column_name
	// The extended format disambiguates when multiple grants share the same entity/privilege/grantKind.
	parts := strings.Split(req.ID, "/")
	if len(parts) != 5 && len(parts) != 8 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: role_id/entity_id/entity_kind/privilege/grant_kind[/schema_name/table_name/column_name], got: %s", req.ID),
		)
		return
	}

	fields := []struct {
		name  string
		value string
	}{
		{"role_id", parts[0]},
		{"entity_id", parts[1]},
		{"entity_kind", parts[2]},
		{"privilege", parts[3]},
		{"grant_kind", parts[4]},
	}
	if len(parts) == 8 {
		fields = append(fields,
			struct{ name, value string }{"schema_name", parts[5]},
			struct{ name, value string }{"table_name", parts[6]},
			struct{ name, value string }{"column_name", parts[7]},
		)
	}

	for _, f := range fields {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(f.name), f.value)...)
	}
}

// Helper methods
func (r *role_privilege_grantResource) modelToCreateRequest(ctx context.Context, model *resource_role_privilege_grant.RolePrivilegeGrantModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	if !model.RoleId.IsNull() && !model.RoleId.IsUnknown() && model.RoleId.ValueString() != "" {
		request["roleId"] = model.RoleId.ValueString()
	}
	if !model.EntityId.IsNull() {
		request["entityId"] = model.EntityId.ValueString()
	}
	if !model.EntityKind.IsNull() {
		request["entityKind"] = model.EntityKind.ValueString()
	}
	if !model.GrantKind.IsNull() {
		request["grantKind"] = model.GrantKind.ValueString()
	}
	if !model.Privilege.IsNull() {
		request["privilege"] = model.Privilege.ValueString()
	}
	if !model.GrantOption.IsNull() {
		request["grantOption"] = model.GrantOption.ValueBool()
	}

	// Optional fields
	if !model.ColumnName.IsNull() && !model.ColumnName.IsUnknown() && model.ColumnName.ValueString() != "" {
		request["columnName"] = model.ColumnName.ValueString()
	}
	if !model.SchemaName.IsNull() && !model.SchemaName.IsUnknown() && model.SchemaName.ValueString() != "" {
		request["schemaName"] = model.SchemaName.ValueString()
	}
	if !model.TableName.IsNull() && !model.TableName.IsUnknown() && model.TableName.ValueString() != "" {
		request["tableName"] = model.TableName.ValueString()
	}

	return request
}

func (r *role_privilege_grantResource) updateModelFromResponse(ctx context.Context, model *resource_role_privilege_grant.RolePrivilegeGrantModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// For privilege grants, we don't modify the EntityId since it's an input parameter
	// that should remain consistent (it represents the entity being granted privileges on)

	// Update other fields from response if they exist
	if roleId, ok := response["roleId"].(string); ok {
		model.RoleId = types.StringValue(roleId)
	}

	if entityId, ok := response["entityId"].(string); ok {
		model.EntityId = types.StringValue(entityId)
	}

	// The API may promote entityKind (e.g., Table → Column with added columnName: "*").
	// Preserve the user's configured value to avoid unnecessary replacement. Only update
	// during import (when entityKind comes from ImportState, not user config).
	if entityKind, ok := response["entityKind"].(string); ok {
		if model.EntityKind.IsNull() || model.EntityKind.IsUnknown() {
			model.EntityKind = types.StringValue(entityKind)
		} else if model.EntityKind.ValueString() != entityKind {
			tflog.Debug(ctx, "API returned different entityKind than configured (likely promotion), preserving user value", map[string]interface{}{
				"configured": model.EntityKind.ValueString(),
				"api":        entityKind,
			})
		}
	}

	if grantKind, ok := response["grantKind"].(string); ok {
		model.GrantKind = types.StringValue(grantKind)
	}

	if privilege, ok := response["privilege"].(string); ok {
		model.Privilege = types.StringValue(privilege)
	}

	// grant_option is Required — the user's config is always authoritative.
	// The list API may return stale values during eventual consistency windows,
	// so we only set it from the API during import (when it's null because
	// ImportState doesn't populate it).
	if model.GrantOption.IsNull() {
		if grantOption, ok := response["grantOption"].(bool); ok {
			model.GrantOption = types.BoolValue(grantOption)
		}
	}

	// For optional scope fields (columnName, schemaName, tableName):
	// - If the API returns a value, use it (the API preserves wildcards as-is)
	// - If the API omits the field and model has no value, set null
	// - If the API omits the field but model already has a value, keep it
	//   (handles entity_kind promotion where the grant request didn't include all fields)
	if columnName, ok := response["columnName"].(string); ok {
		model.ColumnName = types.StringValue(columnName)
	} else if model.ColumnName.IsNull() || model.ColumnName.IsUnknown() {
		model.ColumnName = types.StringNull()
	}

	if schemaName, ok := response["schemaName"].(string); ok {
		model.SchemaName = types.StringValue(schemaName)
	} else if model.SchemaName.IsNull() || model.SchemaName.IsUnknown() {
		model.SchemaName = types.StringNull()
	}

	if tableName, ok := response["tableName"].(string); ok {
		model.TableName = types.StringValue(tableName)
	} else if model.TableName.IsNull() || model.TableName.IsUnknown() {
		model.TableName = types.StringNull()
	}

	// list_all_privileges is not a resource property; always null
	model.ListAllPrivileges = types.BoolNull()
}
