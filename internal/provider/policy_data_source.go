package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_policy"
)

var _ datasource.DataSource = (*policyDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*policyDataSource)(nil)

func NewPolicyDataSource() datasource.DataSource {
	return &policyDataSource{}
}

type policyDataSource struct {
	client *client.GalaxyClient
}

func (d *policyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy"
}

func (d *policyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_policy.PolicyDataSourceSchema(ctx)
}

func (d *policyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *policyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_policy.PolicyModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading policy", map[string]interface{}{"id": id})

	response, err := d.client.GetPolicy(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading policy",
			"Could not read policy "+id+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Policy API response", map[string]interface{}{"response": response})

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *policyDataSource) updateModelFromResponse(ctx context.Context, model *datasource_policy.PolicyModel, response map[string]interface{}) {
	// Map response fields to model - API returns camelCase field names
	if policyId, ok := response["policyId"].(string); ok {
		model.PolicyId = types.StringValue(policyId)
		// Also set the Id field to the policyId for proper identification
		model.Id = types.StringValue(policyId)
	}

	if roleId, ok := response["roleId"].(string); ok {
		model.RoleId = types.StringValue(roleId)
	}

	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}

	if predicate, ok := response["predicate"].(string); ok {
		model.Predicate = types.StringValue(predicate)
	}

	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	}

	if expiration, ok := response["expiration"].(string); ok && expiration != "" {
		model.Expiration = types.StringValue(expiration)
	} else {
		// Set to empty string instead of null when expiration is not set or empty
		model.Expiration = types.StringValue("")
	}

	if created, ok := response["created"].(string); ok {
		model.Created = types.StringValue(created)
	}

	if modified, ok := response["modified"].(string); ok {
		model.Modified = types.StringValue(modified)
	}

	// Map scopes - always provide a valid list to avoid null values
	scopesList := make([]attr.Value, 0)
	if scopesData, ok := response["scopes"].([]interface{}); ok {
		for _, scopeData := range scopesData {
			if scope, ok := scopeData.(map[string]interface{}); ok {
				scopeValue := d.mapScopeToValue(ctx, scope)
				scopesList = append(scopesList, scopeValue)
			}
		}
	}

	// Create scopes list without conversion to avoid tolist() display
	if len(scopesList) > 0 {
		model.Scopes = types.ListValueMust(
			types.ObjectType{AttrTypes: datasource_policy.ScopesValue{}.AttributeTypes(ctx)},
			scopesList,
		)
	} else {
		model.Scopes = types.ListValueMust(
			types.ObjectType{AttrTypes: datasource_policy.ScopesValue{}.AttributeTypes(ctx)},
			[]attr.Value{},
		)
	}
}

func (d *policyDataSource) mapScopeToValue(ctx context.Context, scope map[string]interface{}) attr.Value {
	attributeTypes := datasource_policy.ScopesValue{}.AttributeTypes(ctx)
	attributes := make(map[string]attr.Value)

	// Handle column_mask_ids field - convert from columnMaskIds array
	var columnMaskIdsList types.List
	if columnMaskIds, ok := scope["columnMaskIds"].([]interface{}); ok && len(columnMaskIds) > 0 {
		maskIds := make([]attr.Value, 0, len(columnMaskIds))
		for _, id := range columnMaskIds {
			if idStr, ok := id.(string); ok {
				maskIds = append(maskIds, types.StringValue(idStr))
			}
		}
		columnMaskIdsList = types.ListValueMust(types.StringType, maskIds)
	} else {
		// Create empty list without conversion
		columnMaskIdsList = types.ListValueMust(types.StringType, []attr.Value{})
	}
	attributes["column_mask_ids"] = columnMaskIdsList

	// Set default values for optional fields, then override with actual values if present
	attributes["column_name"] = types.StringValue("")
	if val, ok := scope["columnName"].(string); ok {
		attributes["column_name"] = types.StringValue(val)
	}

	attributes["entity_id"] = types.StringValue("")
	if val, ok := scope["entityId"].(string); ok {
		attributes["entity_id"] = types.StringValue(val)
	}

	attributes["entity_kind"] = types.StringValue("")
	if val, ok := scope["entityKind"].(string); ok {
		attributes["entity_kind"] = types.StringValue(val)
	}

	attributes["schema_name"] = types.StringValue("")
	if val, ok := scope["schemaName"].(string); ok {
		attributes["schema_name"] = types.StringValue(val)
	}

	attributes["table_name"] = types.StringValue("")
	if val, ok := scope["tableName"].(string); ok {
		attributes["table_name"] = types.StringValue(val)
	}

	// Handle row_filter_ids list - create proper empty list to avoid tolist() display
	var rowFilterIdsList types.List
	if rowFilterIds, ok := scope["rowFilterIds"].([]interface{}); ok && len(rowFilterIds) > 0 {
		filterIds := make([]attr.Value, 0, len(rowFilterIds))
		for _, id := range rowFilterIds {
			if idStr, ok := id.(string); ok {
				filterIds = append(filterIds, types.StringValue(idStr))
			}
		}
		rowFilterIdsList = types.ListValueMust(types.StringType, filterIds)
	} else {
		// Create empty list without conversion
		rowFilterIdsList = types.ListValueMust(types.StringType, []attr.Value{})
	}
	attributes["row_filter_ids"] = rowFilterIdsList

	// Handle privileges object - create proper lists to avoid tolist() display
	privAttrs := make(map[string]attr.Value)
	privAttrs["grant_kind"] = types.StringValue("")

	var privilegeList types.List
	if privData, ok := scope["privileges"].(map[string]interface{}); ok {
		if grantKind, ok := privData["grantKind"].(string); ok && grantKind != "" {
			privAttrs["grant_kind"] = types.StringValue(grantKind)
		}

		if privilegeData, ok := privData["privilege"].([]interface{}); ok && len(privilegeData) > 0 {
			privs := make([]attr.Value, 0, len(privilegeData))
			for _, priv := range privilegeData {
				if privStr, ok := priv.(string); ok {
					privs = append(privs, types.StringValue(privStr))
				}
			}
			privilegeList = types.ListValueMust(types.StringType, privs)
		} else {
			privilegeList = types.ListValueMust(types.StringType, []attr.Value{})
		}
	} else {
		privilegeList = types.ListValueMust(types.StringType, []attr.Value{})
	}
	privAttrs["privilege"] = privilegeList

	privObj, _ := types.ObjectValue(datasource_policy.PrivilegesValue{}.AttributeTypes(ctx), privAttrs)
	attributes["privileges"] = privObj

	// Create the object value with custom handling to avoid null conversions
	return d.createNonNullScopeValue(ctx, attributeTypes, attributes)
}

// createNonNullScopeValue creates a scope value ensuring no null values are displayed as functions
func (d *policyDataSource) createNonNullScopeValue(ctx context.Context, attributeTypes map[string]attr.Type, attributes map[string]attr.Value) attr.Value {
	// Create a custom ScopeValue that ensures proper display formatting
	objVal, _ := types.ObjectValue(attributeTypes, attributes)
	return objVal
}
