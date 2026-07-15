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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_policy"
)

var _ resource.Resource = (*policyResource)(nil)
var _ resource.ResourceWithConfigure = (*policyResource)(nil)
var _ resource.ResourceWithImportState = (*policyResource)(nil)

func NewPolicyResource() resource.Resource {
	return &policyResource{}
}

type policyResource struct {
	client *client.GalaxyClient
}

func (r *policyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (r *policyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	s := resource_policy.PolicyResourceSchema(ctx)

	// policy_id is assigned at creation and never changes. Without UseStateForUnknown, any
	// update to the policy causes Terraform to mark policy_id as "known after apply", which
	// propagates to downstream resources referencing it (e.g. galaxy_role_privilege_grant.entity_id)
	// and forces unnecessary destroy/recreate cycles.
	if attr, ok := s.Attributes["policy_id"].(schema.StringAttribute); ok {
		attr.PlanModifiers = []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		}
		s.Attributes["policy_id"] = attr
	}

	// created is assigned at creation and never changes. Without UseStateForUnknown, any update
	// to the policy causes Terraform to mark created as "known after apply".
	if attr, ok := s.Attributes["created"].(schema.StringAttribute); ok {
		attr.PlanModifiers = []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		}
		s.Attributes["created"] = attr
	}

	resp.Schema = s
}

func (r *policyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resource_policy.PolicyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	request := r.modelToCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating policy")
	response, err := r.client.CreatePolicy(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating policy",
			"Could not create policy: "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Created policy", map[string]interface{}{"id": plan.PolicyId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resource_policy.PolicyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.PolicyId.ValueString()
	tflog.Debug(ctx, "Reading policy", map[string]interface{}{"id": id})
	response, err := r.client.GetPolicy(ctx, id)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Policy not found, removing from state", map[string]interface{}{"id": id})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading policy",
			"Could not read policy "+id+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Policy API response in Read", map[string]interface{}{"response": response})

	// Store original state for comparison
	originalState := state

	r.updateModelFromResponse(ctx, &state, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Original state vs updated state", map[string]interface{}{
		"original_scopes_count": len(originalState.Scopes.Elements()),
		"updated_scopes_count":  len(state.Scopes.Elements()),
		"scopes_equal":          originalState.Scopes.Equal(state.Scopes),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan resource_policy.PolicyModel
	var state resource_policy.PolicyModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.PolicyId.ValueString()
	request := r.modelToUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating policy", map[string]interface{}{"id": id, "request": request})
	response, err := r.client.UpdatePolicy(ctx, id, request)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating policy",
			"Could not update policy "+id+": "+err.Error(),
		)
		return
	}

	r.updateModelFromResponse(ctx, &plan, response, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updated policy", map[string]interface{}{"id": plan.PolicyId.ValueString()})
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resource_policy.PolicyModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := state.PolicyId.ValueString()
	tflog.Debug(ctx, "Deleting policy", map[string]interface{}{"id": id})
	err := r.client.DeletePolicy(ctx, id)
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting policy",
				"Could not delete policy "+id+": "+err.Error(),
			)
			return
		}
	}

	tflog.Debug(ctx, "Deleted policy", map[string]interface{}{"id": id})
}

func (r *policyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("policy_id"), req, resp)
}

// Helper methods
func (r *policyResource) modelToCreateRequest(ctx context.Context, model *resource_policy.PolicyModel, diags *diag.Diagnostics) map[string]interface{} {
	request := make(map[string]interface{})

	// Required fields
	if !model.Name.IsNull() {
		request["name"] = model.Name.ValueString()
	}
	if !model.Description.IsNull() {
		request["description"] = model.Description.ValueString()
	}
	if !model.Predicate.IsNull() {
		request["predicate"] = model.Predicate.ValueString()
	}
	if !model.RoleId.IsNull() {
		request["roleId"] = model.RoleId.ValueString()
	}

	// Optional fields
	if !model.Expiration.IsNull() && !model.Expiration.IsUnknown() && model.Expiration.ValueString() != "" {
		request["expiration"] = model.Expiration.ValueString()
	}

	// Handle scopes list
	if !model.Scopes.IsNull() && !model.Scopes.IsUnknown() {
		scopes := []map[string]interface{}{}
		elements := model.Scopes.Elements()

		for _, elem := range elements {
			if elem.IsNull() || elem.IsUnknown() {
				continue
			}

			scopeVal, ok := elem.(resource_policy.ScopesValue)
			if !ok {
				diags.AddError("Invalid scope type", fmt.Sprintf("Expected ScopesValue, got %T", elem))
				continue
			}

			// Validate required fields
			if scopeVal.EntityId.IsNull() || scopeVal.EntityId.IsUnknown() {
				diags.AddError("Missing entity_id", "EntityId is required for policy scopes")
				continue
			}
			if scopeVal.EntityKind.IsNull() || scopeVal.EntityKind.IsUnknown() {
				diags.AddError("Missing entity_kind", "EntityKind is required for policy scopes")
				continue
			}

			scope := map[string]interface{}{
				"entityId":   scopeVal.EntityId.ValueString(),
				"entityKind": scopeVal.EntityKind.ValueString(),
			}

			// Handle column mask IDs
			columnMaskIds := []string{}
			if !scopeVal.ColumnMaskIds.IsNull() && !scopeVal.ColumnMaskIds.IsUnknown() {
				maskElements := make([]types.String, 0, len(scopeVal.ColumnMaskIds.Elements()))
				if elemDiags := scopeVal.ColumnMaskIds.ElementsAs(ctx, &maskElements, false); !elemDiags.HasError() {
					for _, maskElem := range maskElements {
						if !maskElem.IsNull() && !maskElem.IsUnknown() {
							columnMaskIds = append(columnMaskIds, maskElem.ValueString())
						}
					}
				} else {
					diags.Append(elemDiags...)
				}
			}
			scope["columnMaskIds"] = columnMaskIds

			// Handle row filter IDs
			rowFilterIds := []string{}
			if !scopeVal.RowFilterIds.IsNull() && !scopeVal.RowFilterIds.IsUnknown() {
				filterElements := make([]types.String, 0, len(scopeVal.RowFilterIds.Elements()))
				if elemDiags := scopeVal.RowFilterIds.ElementsAs(ctx, &filterElements, false); !elemDiags.HasError() {
					for _, filterElem := range filterElements {
						if !filterElem.IsNull() && !filterElem.IsUnknown() {
							rowFilterIds = append(rowFilterIds, filterElem.ValueString())
						}
					}
				} else {
					diags.Append(elemDiags...)
				}
			}
			scope["rowFilterIds"] = rowFilterIds

			// Optional string fields
			if !scopeVal.ColumnName.IsNull() && scopeVal.ColumnName.ValueString() != "" {
				scope["columnName"] = scopeVal.ColumnName.ValueString()
			}
			if !scopeVal.SchemaName.IsNull() && scopeVal.SchemaName.ValueString() != "" {
				scope["schemaName"] = scopeVal.SchemaName.ValueString()
			}
			if !scopeVal.TableName.IsNull() && scopeVal.TableName.ValueString() != "" {
				scope["tableName"] = scopeVal.TableName.ValueString()
			}

			// Handle privileges - only include if explicitly set and not empty
			tflog.Debug(ctx, "Processing privileges", map[string]interface{}{
				"privileges_is_null":    scopeVal.Privileges.IsNull(),
				"privileges_is_unknown": scopeVal.Privileges.IsUnknown(),
			})
			if !scopeVal.Privileges.IsNull() && !scopeVal.Privileges.IsUnknown() {
				privilegeAttrs := scopeVal.Privileges.Attributes()
				if len(privilegeAttrs) > 0 {
					privilegeMap := map[string]interface{}{}
					hasGrantKind := false
					hasPrivileges := false

					// Check grant_kind
					if grantKindAttr, exists := privilegeAttrs["grant_kind"]; exists {
						if grantKindVal, ok := grantKindAttr.(basetypes.StringValue); ok && !grantKindVal.IsNull() && !grantKindVal.IsUnknown() && grantKindVal.ValueString() != "" {
							privilegeMap["grantKind"] = grantKindVal.ValueString()
							hasGrantKind = true
						}
					}

					// Check privilege list
					if privilegeAttr, exists := privilegeAttrs["privilege"]; exists {
						if privilegeListVal, ok := privilegeAttr.(basetypes.ListValue); ok && !privilegeListVal.IsNull() && !privilegeListVal.IsUnknown() {
							privElements := make([]types.String, 0, len(privilegeListVal.Elements()))
							if elemDiags := privilegeListVal.ElementsAs(ctx, &privElements, false); !elemDiags.HasError() {
								privileges := []string{}
								for _, privElem := range privElements {
									if !privElem.IsNull() && !privElem.IsUnknown() && privElem.ValueString() != "" {
										privileges = append(privileges, privElem.ValueString())
									}
								}
								// Always include privilege list, even if empty
								privilegeMap["privilege"] = privileges
								hasPrivileges = true
							} else {
								diags.Append(elemDiags...)
							}
						}
					}

					// Include privileges if grant_kind is set (privilege list can be empty)
					tflog.Debug(ctx, "Privilege validation", map[string]interface{}{
						"hasGrantKind":  hasGrantKind,
						"hasPrivileges": hasPrivileges,
						"privilegeMap":  privilegeMap,
					})
					if hasGrantKind {
						scope["privileges"] = privilegeMap
					}
				}
			}

			scopes = append(scopes, scope)
		}
		request["scopes"] = scopes
	}

	return request
}

func (r *policyResource) modelToUpdateRequest(ctx context.Context, model *resource_policy.PolicyModel, diags *diag.Diagnostics) map[string]interface{} {
	return r.modelToCreateRequest(ctx, model, diags)
}

func (r *policyResource) updateModelFromResponse(ctx context.Context, model *resource_policy.PolicyModel, response map[string]interface{}, diags *diag.Diagnostics) {
	// Set ID - try both possible field names
	if id, ok := response["policyId"].(string); ok {
		model.PolicyId = types.StringValue(id)
	}

	// Set computed fields from API response
	if policyId, ok := response["policyId"].(string); ok {
		model.PolicyId = types.StringValue(policyId)
	}

	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}

	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	} else {
		model.Description = types.StringNull()
	}

	if predicate, ok := response["predicate"].(string); ok {
		model.Predicate = types.StringValue(predicate)
	}

	if roleId, ok := response["roleId"].(string); ok {
		model.RoleId = types.StringValue(roleId)
	}

	if expiration, ok := response["expiration"].(string); ok {
		model.Expiration = types.StringValue(expiration)
	} else {
		model.Expiration = types.StringNull()
	}

	if created, ok := response["created"].(string); ok {
		model.Created = types.StringValue(created)
	}

	if modified, ok := response["modified"].(string); ok {
		model.Modified = types.StringValue(modified)
	}

	// Update scopes from response to ensure consistency
	if scopesData, ok := response["scopes"].([]interface{}); ok {
		r.updateScopesFromResponse(ctx, model, scopesData, diags)
	}
}

func (r *policyResource) updateScopesFromResponse(ctx context.Context, model *resource_policy.PolicyModel, scopesData []interface{}, diags *diag.Diagnostics) {
	// Capture the planned scopes before they are overwritten, so column_mask_ids and
	// row_filter_ids (Required, order-sensitive attributes) can be reordered to match
	// the plan. The Galaxy API does not preserve submission order for these ID lists,
	// which otherwise causes "Provider produced inconsistent result after apply" on
	// every apply since Terraform requires non-computed attributes to match exactly.
	// Planned scopes are keyed by scope identity (entity + schema/table/column names) so a
	// response scope is matched to its plan counterpart regardless of scopes[] order.
	plannedScopesByIdentity := map[string]resource_policy.ScopesValue{}
	if !model.Scopes.IsNull() && !model.Scopes.IsUnknown() {
		planElements := make([]resource_policy.ScopesValue, 0, len(model.Scopes.Elements()))
		if elemDiags := model.Scopes.ElementsAs(ctx, &planElements, false); !elemDiags.HasError() {
			for _, s := range planElements {
				plannedScopesByIdentity[plannedScopeIdentity(s)] = s
			}
		}
	}

	// Convert API response scopes to model format
	scopesList := make([]resource_policy.ScopesValue, 0, len(scopesData))

	for _, scopeData := range scopesData {
		if scope, ok := scopeData.(map[string]interface{}); ok {
			var plannedColumnMaskIds, plannedRowFilterIds, plannedPrivilegeIds []string
			if plannedScope, found := plannedScopesByIdentity[responseScopeIdentity(scope)]; found {
				maskIds, maskDiags := stringListElements(ctx, plannedScope.ColumnMaskIds)
				diags.Append(maskDiags...)
				filterIds, filterDiags := stringListElements(ctx, plannedScope.RowFilterIds)
				diags.Append(filterDiags...)
				privIds, privDiags := plannedPrivilegeElements(ctx, plannedScope.Privileges)
				diags.Append(privDiags...)
				plannedColumnMaskIds, plannedRowFilterIds, plannedPrivilegeIds = maskIds, filterIds, privIds
			}
			scopeValue := r.mapAPIResponseScopeToModelValue(ctx, scope, plannedColumnMaskIds, plannedRowFilterIds, plannedPrivilegeIds, diags)
			if !diags.HasError() {
				scopesList = append(scopesList, scopeValue)
			}
		}
	}

	// Use proper custom type initialization
	scopesElementType := resource_policy.ScopesType{
		ObjectType: types.ObjectType{
			AttrTypes: resource_policy.ScopesValue{}.AttributeTypes(ctx),
		},
	}

	scopesListValue, listDiags := types.ListValueFrom(ctx, scopesElementType, scopesList)
	diags.Append(listDiags...)
	if !listDiags.HasError() {
		model.Scopes = scopesListValue
	}
}

func (r *policyResource) mapAPIResponseScopeToModelValue(ctx context.Context, scope map[string]interface{}, plannedColumnMaskIds []string, plannedRowFilterIds []string, plannedPrivilegeIds []string, diags *diag.Diagnostics) resource_policy.ScopesValue {
	attributeTypes := resource_policy.ScopesValue{}.AttributeTypes(ctx)
	attributes := make(map[string]attr.Value)

	// Map required string fields
	attributes["entity_id"] = types.StringValue("")
	if val, ok := scope["entityId"].(string); ok {
		attributes["entity_id"] = types.StringValue(val)
	}

	attributes["entity_kind"] = types.StringValue("")
	if val, ok := scope["entityKind"].(string); ok {
		attributes["entity_kind"] = types.StringValue(val)
	}

	// Map optional string fields - use defaults to match configuration
	attributes["column_name"] = types.StringValue("*")
	if val, ok := scope["columnName"].(string); ok && val != "" {
		attributes["column_name"] = types.StringValue(val)
	}

	attributes["schema_name"] = types.StringValue("*")
	if val, ok := scope["schemaName"].(string); ok && val != "" {
		attributes["schema_name"] = types.StringValue(val)
	}

	attributes["table_name"] = types.StringValue("*")
	if val, ok := scope["tableName"].(string); ok && val != "" {
		attributes["table_name"] = types.StringValue(val)
	}

	// Handle column_mask_ids list. The API does not preserve submission order, so reorder
	// its response to match the plan (matching on set membership) to avoid "Provider
	// produced inconsistent result after apply" for this Required, order-sensitive attribute.
	columnMaskIdStrings := make([]string, 0)
	if maskIdsData, ok := scope["columnMaskIds"].([]interface{}); ok {
		for _, maskId := range maskIdsData {
			if maskIdStr, ok := maskId.(string); ok {
				columnMaskIdStrings = append(columnMaskIdStrings, maskIdStr)
			}
		}
	}
	columnMaskIdStrings = reorderToMatchPlan(plannedColumnMaskIds, columnMaskIdStrings)
	columnMaskIds := make([]attr.Value, 0, len(columnMaskIdStrings))
	for _, maskIdStr := range columnMaskIdStrings {
		columnMaskIds = append(columnMaskIds, types.StringValue(maskIdStr))
	}
	columnMaskIdsList, _ := types.ListValue(types.StringType, columnMaskIds)
	attributes["column_mask_ids"] = columnMaskIdsList

	// Handle row_filter_ids list. Same order-preservation treatment as column_mask_ids above.
	rowFilterIdStrings := make([]string, 0)
	if filterIdsData, ok := scope["rowFilterIds"].([]interface{}); ok {
		for _, filterId := range filterIdsData {
			if filterIdStr, ok := filterId.(string); ok {
				rowFilterIdStrings = append(rowFilterIdStrings, filterIdStr)
			}
		}
	}
	rowFilterIdStrings = reorderToMatchPlan(plannedRowFilterIds, rowFilterIdStrings)
	rowFilterIds := make([]attr.Value, 0, len(rowFilterIdStrings))
	for _, filterIdStr := range rowFilterIdStrings {
		rowFilterIds = append(rowFilterIds, types.StringValue(filterIdStr))
	}
	rowFilterIdsList, _ := types.ListValue(types.StringType, rowFilterIds)
	attributes["row_filter_ids"] = rowFilterIdsList

	// Handle privileges object - only include if present in API response
	if privData, ok := scope["privileges"].(map[string]interface{}); ok {
		privAttrs := make(map[string]attr.Value)
		privAttrs["grant_kind"] = types.StringValue("Allow") // Default value
		privAttrs["privilege"] = types.ListValueMust(types.StringType, []attr.Value{})

		if grantKind, ok := privData["grantKind"].(string); ok {
			privAttrs["grant_kind"] = types.StringValue(grantKind)
		}

		if privilegeList, ok := privData["privilege"].([]interface{}); ok {
			privStrings := make([]string, 0, len(privilegeList))
			for _, priv := range privilegeList {
				if privStr, ok := priv.(string); ok {
					privStrings = append(privStrings, privStr)
				}
			}
			privStrings = reorderToMatchPlan(plannedPrivilegeIds, privStrings)
			if len(privStrings) > 0 {
				privs := make([]attr.Value, 0, len(privStrings))
				for _, privStr := range privStrings {
					privs = append(privs, types.StringValue(privStr))
				}
				privListVal, _ := types.ListValue(types.StringType, privs)
				privAttrs["privilege"] = privListVal
			}
		}

		privObj, _ := types.ObjectValue(resource_policy.PrivilegesValue{}.AttributeTypes(ctx), privAttrs)
		attributes["privileges"] = privObj
	} else {
		// If privileges are not in API response, set to null
		attributes["privileges"] = types.ObjectNull(resource_policy.PrivilegesValue{}.AttributeTypes(ctx))
	}

	// Create the scope value object
	scopeValue, scopeDiags := resource_policy.NewScopesValue(attributeTypes, attributes)
	diags.Append(scopeDiags...)
	return scopeValue
}

// plannedScopeIdentity builds the identity key for a planned scope. The five fields below form
// a scope's natural key (an entity in a specific schema/table/column context) so a response
// scope can be matched to its plan counterpart regardless of the API's returned order.
func plannedScopeIdentity(s resource_policy.ScopesValue) string {
	return s.EntityId.ValueString() + "\x00" +
		s.EntityKind.ValueString() + "\x00" +
		s.SchemaName.ValueString() + "\x00" +
		s.TableName.ValueString() + "\x00" +
		s.ColumnName.ValueString()
}

// responseScopeIdentity builds the identity key for a scope in the raw API response, matching
// the shape of plannedScopeIdentity.
func responseScopeIdentity(scope map[string]interface{}) string {
	get := func(k string) string {
		if v, ok := scope[k].(string); ok {
			return v
		}
		return ""
	}
	return get("entityId") + "\x00" +
		get("entityKind") + "\x00" +
		get("schemaName") + "\x00" +
		get("tableName") + "\x00" +
		get("columnName")
}

// plannedPrivilegeElements extracts the planned privilege order from a scope's privileges
// object, returning nil if the object or its privilege list is null/unknown. Any diagnostics
// from list extraction are returned to the caller.
func plannedPrivilegeElements(ctx context.Context, privileges types.Object) ([]string, diag.Diagnostics) {
	if privileges.IsNull() || privileges.IsUnknown() {
		return nil, nil
	}
	privilegeAttr, ok := privileges.Attributes()["privilege"]
	if !ok {
		return nil, nil
	}
	privilegeList, ok := privilegeAttr.(basetypes.ListValue)
	if !ok {
		return nil, nil
	}
	return stringListElements(ctx, privilegeList)
}
