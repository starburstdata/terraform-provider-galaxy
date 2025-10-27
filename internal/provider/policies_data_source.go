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
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_policies"
)

var _ datasource.DataSource = (*policiesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*policiesDataSource)(nil)

func NewPoliciesDataSource() datasource.DataSource {
	return &policiesDataSource{}
}

type policiesDataSource struct {
	client *client.GalaxyClient
}

func (d *policiesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policies"
}

func (d *policiesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_policies.PoliciesDataSourceSchema(ctx)
}

func (d *policiesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.GalaxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.GalaxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *policiesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_policies.PoliciesModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading policies with automatic pagination")

	// Use automatic pagination to get ALL policies across all pages
	allPolicies, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/policy")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading policies",
			"Could not read policies: "+err.Error(),
		)
		return
	}

	// Convert []interface{} to []map[string]interface{} for mapping
	var policyMaps []map[string]interface{}
	for _, policyInterface := range allPolicies {
		if policyMap, ok := policyInterface.(map[string]interface{}); ok {
			policyMaps = append(policyMaps, policyMap)
		}
	}

	// Map API response to model
	if len(policyMaps) > 0 {
		policies, err := d.mapPoliciesResult(ctx, policyMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping policies response",
				"Could not map policies response: "+err.Error(),
			)
			return
		}
		config.Result = policies
	} else {
		elementType := datasource_policies.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_policies.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_policies.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *policiesDataSource) mapPoliciesResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	policies := make([]datasource_policies.ResultValue, 0, len(result))

	for _, policyMap := range result {
		policy := d.mapSinglePolicy(ctx, policyMap)
		policies = append(policies, policy)
	}

	elementType := datasource_policies.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_policies.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, policies)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *policiesDataSource) mapSinglePolicy(ctx context.Context, policyMap map[string]interface{}) datasource_policies.ResultValue {
	attributeTypes := datasource_policies.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map policy ID
	if policyId, ok := policyMap["policyId"].(string); ok {
		attributes["policy_id"] = types.StringValue(policyId)
	} else {
		attributes["policy_id"] = types.StringNull()
	}

	// Map name
	if name, ok := policyMap["name"].(string); ok {
		attributes["name"] = types.StringValue(name)
	} else {
		attributes["name"] = types.StringNull()
	}

	// Map description
	if description, ok := policyMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	// Map predicate
	if predicate, ok := policyMap["predicate"].(string); ok {
		attributes["predicate"] = types.StringValue(predicate)
	} else {
		attributes["predicate"] = types.StringNull()
	}

	// Map role ID
	if roleId, ok := policyMap["roleId"].(string); ok {
		attributes["role_id"] = types.StringValue(roleId)
	} else {
		attributes["role_id"] = types.StringNull()
	}

	// Map expiration
	if expiration, ok := policyMap["expiration"].(string); ok {
		attributes["expiration"] = types.StringValue(expiration)
	} else {
		attributes["expiration"] = types.StringNull()
	}

	// Map created
	if created, ok := policyMap["created"].(string); ok {
		attributes["created"] = types.StringValue(created)
	} else {
		attributes["created"] = types.StringNull()
	}

	// Map modified
	if modified, ok := policyMap["modified"].(string); ok {
		attributes["modified"] = types.StringValue(modified)
	} else {
		attributes["modified"] = types.StringNull()
	}

	// Handle scopes list - this is a list of complex objects, not strings
	if scopes, ok := policyMap["scopes"].([]interface{}); ok {
		scopesList := make([]datasource_policies.ScopesValue, 0, len(scopes))
		for _, scopeInterface := range scopes {
			if scopeMap, ok := scopeInterface.(map[string]interface{}); ok {
				scope := d.mapSingleScope(ctx, scopeMap)
				scopesList = append(scopesList, scope)
			}
		}

		elementType := datasource_policies.ScopesType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_policies.ScopesValue{}.AttributeTypes(ctx),
			},
		}
		scopesListValue, _ := types.ListValueFrom(ctx, elementType, scopesList)
		attributes["scopes"] = scopesListValue
	} else {
		elementType := datasource_policies.ScopesType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_policies.ScopesValue{}.AttributeTypes(ctx),
			},
		}
		attributes["scopes"] = types.ListNull(elementType)
	}

	// Create the ResultValue using the constructor
	policy, diags := datasource_policies.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating policy ResultValue: %v", diags))
		return datasource_policies.NewResultValueNull()
	}

	return policy
}

func (d *policiesDataSource) mapSingleScope(ctx context.Context, scopeMap map[string]interface{}) datasource_policies.ScopesValue {
	attributeTypes := datasource_policies.ScopesValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map column_mask_id
	if columnMaskId, ok := scopeMap["columnMaskId"].(string); ok {
		attributes["column_mask_id"] = types.StringValue(columnMaskId)
	} else {
		attributes["column_mask_id"] = types.StringNull()
	}

	// Map column_name
	if columnName, ok := scopeMap["columnName"].(string); ok {
		attributes["column_name"] = types.StringValue(columnName)
	} else {
		attributes["column_name"] = types.StringNull()
	}

	// Map entity_id
	if entityId, ok := scopeMap["entityId"].(string); ok {
		attributes["entity_id"] = types.StringValue(entityId)
	} else {
		attributes["entity_id"] = types.StringNull()
	}

	// Map entity_kind
	if entityKind, ok := scopeMap["entityKind"].(string); ok {
		attributes["entity_kind"] = types.StringValue(entityKind)
	} else {
		attributes["entity_kind"] = types.StringNull()
	}

	// Map privileges (complex nested object)
	if privilegesMap, ok := scopeMap["privileges"].(map[string]interface{}); ok {
		privileges := d.mapPrivileges(ctx, privilegesMap)
		attributes["privileges"] = privileges
	} else {
		privilegeTypes := datasource_policies.PrivilegesValue{}.AttributeTypes(ctx)
		attributes["privileges"] = types.ObjectNull(privilegeTypes)
	}

	// Map row_filter_ids
	if rowFilterIds, ok := scopeMap["rowFilterIds"].([]interface{}); ok {
		filterIds := make([]attr.Value, 0, len(rowFilterIds))
		for _, filterInterface := range rowFilterIds {
			if filterId, ok := filterInterface.(string); ok {
				filterIds = append(filterIds, types.StringValue(filterId))
			}
		}
		filterIdsListValue, _ := types.ListValue(types.StringType, filterIds)
		attributes["row_filter_ids"] = filterIdsListValue
	} else {
		attributes["row_filter_ids"] = types.ListNull(types.StringType)
	}

	// Map schema_name
	if schemaName, ok := scopeMap["schemaName"].(string); ok {
		attributes["schema_name"] = types.StringValue(schemaName)
	} else {
		attributes["schema_name"] = types.StringNull()
	}

	// Map table_name
	if tableName, ok := scopeMap["tableName"].(string); ok {
		attributes["table_name"] = types.StringValue(tableName)
	} else {
		attributes["table_name"] = types.StringNull()
	}

	// Create the ScopesValue using the constructor
	scope, diags := datasource_policies.NewScopesValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating scope ScopesValue: %v", diags))
		return datasource_policies.NewScopesValueNull()
	}

	return scope
}

func (d *policiesDataSource) mapPrivileges(ctx context.Context, privilegesMap map[string]interface{}) basetypes.ObjectValue {
	attributeTypes := datasource_policies.PrivilegesValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map grant_kind
	if grantKind, ok := privilegesMap["grantKind"].(string); ok {
		attributes["grant_kind"] = types.StringValue(grantKind)
	} else {
		attributes["grant_kind"] = types.StringNull()
	}

	// Map privilege list
	if privileges, ok := privilegesMap["privilege"].([]interface{}); ok {
		privilegesList := make([]attr.Value, 0, len(privileges))
		for _, privilegeInterface := range privileges {
			if privilege, ok := privilegeInterface.(string); ok {
				privilegesList = append(privilegesList, types.StringValue(privilege))
			}
		}
		privilegesListValue, _ := types.ListValue(types.StringType, privilegesList)
		attributes["privilege"] = privilegesListValue
	} else {
		attributes["privilege"] = types.ListNull(types.StringType)
	}

	// Create ObjectValue
	objValue, diags := types.ObjectValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating privileges object: %v", diags))
		return types.ObjectNull(attributeTypes)
	}

	return objValue
}
