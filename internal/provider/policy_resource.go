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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/resource_policy"
)

var _ resource.Resource = (*policyResource)(nil)
var _ resource.ResourceWithConfigure = (*policyResource)(nil)

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
	resp.Schema = resource_policy.PolicyResourceSchema(ctx)
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
	if !model.Expiration.IsNull() && model.Expiration.ValueString() != "" {
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
						if privilegeListVal, ok := privilegeAttr.(basetypes.ListValue); ok && !privilegeListVal.IsNull() && !privilegeListVal.IsUnknown() && len(privilegeListVal.Elements()) > 0 {
							privElements := make([]types.String, 0, len(privilegeListVal.Elements()))
							if elemDiags := privilegeListVal.ElementsAs(ctx, &privElements, false); !elemDiags.HasError() {
								privileges := []string{}
								for _, privElem := range privElements {
									if !privElem.IsNull() && !privElem.IsUnknown() && privElem.ValueString() != "" {
										privileges = append(privileges, privElem.ValueString())
									}
								}
								if len(privileges) > 0 {
									privilegeMap["privilege"] = privileges
									hasPrivileges = true
								}
							} else {
								diags.Append(elemDiags...)
							}
						}
					}

					// Only include privileges if both grantKind and privilege list are explicitly set with valid values
					tflog.Debug(ctx, "Privilege validation", map[string]interface{}{
						"hasGrantKind":  hasGrantKind,
						"hasPrivileges": hasPrivileges,
						"privilegeMap":  privilegeMap,
					})
					if hasGrantKind && hasPrivileges {
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
		model.Expiration = types.StringValue("")
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
	// Convert API response scopes to model format
	scopesList := make([]attr.Value, 0, len(scopesData))

	for _, scopeData := range scopesData {
		if scope, ok := scopeData.(map[string]interface{}); ok {
			scopeValue := r.mapAPIResponseScopeToModelValue(ctx, scope, diags)
			if !diags.HasError() {
				objVal, objDiags := scopeValue.ToObjectValue(ctx)
				diags.Append(objDiags...)
				if !objDiags.HasError() {
					scopesList = append(scopesList, objVal)
				}
			}
		}
	}

	scopesListValue, listDiags := types.ListValue(
		types.ObjectType{AttrTypes: resource_policy.ScopesValue{}.AttributeTypes(ctx)},
		scopesList,
	)
	diags.Append(listDiags...)
	if !listDiags.HasError() {
		model.Scopes = scopesListValue
	}
}

func (r *policyResource) mapAPIResponseScopeToModelValue(ctx context.Context, scope map[string]interface{}, diags *diag.Diagnostics) resource_policy.ScopesValue {
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

	// Handle column_mask_ids list
	columnMaskIds := make([]attr.Value, 0)
	if maskIdsData, ok := scope["columnMaskIds"].([]interface{}); ok {
		for _, maskId := range maskIdsData {
			if maskIdStr, ok := maskId.(string); ok {
				columnMaskIds = append(columnMaskIds, types.StringValue(maskIdStr))
			}
		}
	}
	columnMaskIdsList, _ := types.ListValue(types.StringType, columnMaskIds)
	attributes["column_mask_ids"] = columnMaskIdsList

	// Handle row_filter_ids list
	rowFilterIds := make([]attr.Value, 0)
	if filterIdsData, ok := scope["rowFilterIds"].([]interface{}); ok {
		for _, filterId := range filterIdsData {
			if filterIdStr, ok := filterId.(string); ok {
				rowFilterIds = append(rowFilterIds, types.StringValue(filterIdStr))
			}
		}
	}
	rowFilterIdsList, _ := types.ListValue(types.StringType, rowFilterIds)
	attributes["row_filter_ids"] = rowFilterIdsList

	// Handle privileges object - ensure consistency with configuration
	privAttrs := make(map[string]attr.Value)
	privAttrs["grant_kind"] = types.StringValue("Allow") // Default value
	privAttrs["privilege"] = types.ListValueMust(types.StringType, []attr.Value{})

	if privData, ok := scope["privileges"].(map[string]interface{}); ok {
		if grantKind, ok := privData["grantKind"].(string); ok {
			privAttrs["grant_kind"] = types.StringValue(grantKind)
		}

		if privilegeList, ok := privData["privilege"].([]interface{}); ok && len(privilegeList) > 0 {
			privs := make([]attr.Value, 0, len(privilegeList))
			for _, priv := range privilegeList {
				if privStr, ok := priv.(string); ok {
					privs = append(privs, types.StringValue(privStr))
				}
			}
			privListVal, _ := types.ListValue(types.StringType, privs)
			privAttrs["privilege"] = privListVal
		}
	}

	privObj, _ := types.ObjectValue(resource_policy.PrivilegesValue{}.AttributeTypes(ctx), privAttrs)
	attributes["privileges"] = privObj

	// Create the scope value object
	scopeValue, scopeDiags := resource_policy.NewScopesValue(attributeTypes, attributes)
	diags.Append(scopeDiags...)
	return scopeValue
}
