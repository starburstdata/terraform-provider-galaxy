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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
)

var _ resource.Resource = (*roleGrantResource)(nil)
var _ resource.ResourceWithConfigure = (*roleGrantResource)(nil)
var _ resource.ResourceWithImportState = (*roleGrantResource)(nil)

func NewRoleGrantResource() resource.Resource {
	return &roleGrantResource{}
}

type roleGrantResource struct {
	client *client.GalaxyClient
}

type roleGrantModel struct {
	RoleId          types.String `tfsdk:"role_id"`
	GrantedRoleId   types.String `tfsdk:"granted_role_id"`
	AdminOption     types.Bool   `tfsdk:"admin_option"`
	GrantedRoleName types.String `tfsdk:"granted_role_name"`
}

type grantMatch struct {
	adminOption     bool
	grantedRoleName string
	found           bool
}

func findGrant(grants []map[string]interface{}, grantedRoleID string) grantMatch {
	for _, g := range grants {
		if getStringFromMap(g, "roleId") == grantedRoleID {
			result := grantMatch{found: true, adminOption: getBoolFromMap(g, "adminOption")}
			if name := getStringFromMap(g, "roleName"); name != "" {
				result.grantedRoleName = name
			}
			return result
		}
	}
	return grantMatch{}
}

func (r *roleGrantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_grant"
}

func (r *roleGrantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an individual role-to-role grant. This is a non-authoritative resource that manages a single grant entry in a role's directlyGrantedRoles list. When multiple Terraform configurations or external systems manage grants on the same role, changes may conflict.",
		Attributes: map[string]schema.Attribute{
			"role_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the role receiving the grant.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"granted_role_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the role being granted.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"admin_option": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the grant includes the WITH ADMIN OPTION privilege.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"granted_role_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the granted role, populated from the API.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *roleGrantResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.GalaxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.GalaxyClient, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *roleGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan roleGrantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := plan.RoleId.ValueString()
	grantedRoleID := plan.GrantedRoleId.ValueString()
	adminOption := plan.AdminOption.ValueBool()

	tflog.Debug(ctx, "Creating role_grant", map[string]interface{}{
		"roleId":        roleID,
		"grantedRoleId": grantedRoleID,
	})

	// Look up the granted role to get its name (required by the API)
	grantedRole, err := r.client.GetRole(ctx, grantedRoleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading granted role",
			"Could not read role "+grantedRoleID+": "+err.Error(),
		)
		return
	}
	grantedRoleName := ""
	if name, ok := grantedRole["roleName"].(string); ok {
		grantedRoleName = name
	}

	// Read current grants
	grants, err := r.client.GetRoleGrants(ctx, roleID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading role grants",
			"Could not read current grants for role "+roleID+": "+err.Error(),
		)
		return
	}

	// Check for duplicate
	for _, g := range grants {
		if getStringFromMap(g, "roleId") == grantedRoleID {
			resp.Diagnostics.AddError(
				"Duplicate role grant",
				fmt.Sprintf("Role %s is already granted to role %s.", grantedRoleID, roleID),
			)
			return
		}
	}

	// Append new grant (roleName is required by the PATCH API)
	newGrant := map[string]interface{}{
		"roleId":      grantedRoleID,
		"roleName":    grantedRoleName,
		"adminOption": adminOption,
	}
	grants = append(grants, newGrant)

	// PATCH the full list
	_, err = r.client.UpdateRoleGrants(ctx, roleID, grants)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating role grant",
			"Could not update role grants: "+err.Error(),
		)
		return
	}

	// Re-read to get the granted role name
	r.readAndUpdateModel(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state roleGrantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := state.RoleId.ValueString()
	grantedRoleID := state.GrantedRoleId.ValueString()

	tflog.Debug(ctx, "Reading role_grant", map[string]interface{}{
		"roleId":        roleID,
		"grantedRoleId": grantedRoleID,
	})

	grants, err := r.client.GetRoleGrants(ctx, roleID)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Role not found, removing role_grant from state", map[string]interface{}{"roleId": roleID})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading role grants",
			"Could not read grants for role "+roleID+": "+err.Error(),
		)
		return
	}

	// Find the specific grant
	match := findGrant(grants, grantedRoleID)
	if match.found {
		state.AdminOption = types.BoolValue(match.adminOption)
		if match.grantedRoleName != "" {
			state.GrantedRoleName = types.StringValue(match.grantedRoleName)
		} else {
			state.GrantedRoleName = types.StringNull()
		}
	}

	if !match.found {
		tflog.Warn(ctx, "Role grant not found, removing from state", map[string]interface{}{
			"roleId":        roleID,
			"grantedRoleId": grantedRoleID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *roleGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes use RequiresReplace, so Update should never be called.
	// If it is, just read the current state.
	var plan roleGrantModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *roleGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state roleGrantModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	roleID := state.RoleId.ValueString()
	grantedRoleID := state.GrantedRoleId.ValueString()

	tflog.Debug(ctx, "Deleting role_grant", map[string]interface{}{
		"roleId":        roleID,
		"grantedRoleId": grantedRoleID,
	})

	// Read current grants
	grants, err := r.client.GetRoleGrants(ctx, roleID)
	if err != nil {
		if client.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error reading role grants",
			"Could not read current grants for role "+roleID+": "+err.Error(),
		)
		return
	}

	// Filter out the grant being deleted
	filtered := make([]map[string]interface{}, 0, len(grants))
	for _, g := range grants {
		if getStringFromMap(g, "roleId") != grantedRoleID {
			filtered = append(filtered, g)
		}
	}

	// PATCH with filtered list
	_, err = r.client.UpdateRoleGrants(ctx, roleID, filtered)
	if err != nil {
		if client.IsNotFound(err) {
			tflog.Warn(ctx, "Role not found during role_grant delete; treating as already deleted", map[string]interface{}{
				"roleId":        roleID,
				"grantedRoleId": grantedRoleID,
			})
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting role grant",
			"Could not update role grants: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Deleted role_grant", map[string]interface{}{
		"roleId":        roleID,
		"grantedRoleId": grantedRoleID,
	})
}

func (r *roleGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format role_id/granted_role_id",
		)
		return
	}

	state := roleGrantModel{
		RoleId:          types.StringValue(parts[0]),
		GrantedRoleId:   types.StringValue(parts[1]),
		AdminOption:     types.BoolNull(),
		GrantedRoleName: types.StringNull(),
	}

	// Read from API to populate admin_option and granted_role_name
	grants, err := r.client.GetRoleGrants(ctx, parts[0])
	if err != nil {
		resp.Diagnostics.AddError(
			"Error importing role grant",
			"Could not read grants for role "+parts[0]+": "+err.Error(),
		)
		return
	}

	match := findGrant(grants, parts[1])
	if !match.found {
		resp.Diagnostics.AddError(
			"Role grant not found",
			fmt.Sprintf("Role %s does not have a grant for role %s.", parts[0], parts[1]),
		)
		return
	}

	state.AdminOption = types.BoolValue(match.adminOption)
	if match.grantedRoleName != "" {
		state.GrantedRoleName = types.StringValue(match.grantedRoleName)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// readAndUpdateModel reads the current grants from the API and updates the model
func (r *roleGrantResource) readAndUpdateModel(ctx context.Context, model *roleGrantModel, diags *diag.Diagnostics) {
	grants, err := r.client.GetRoleGrants(ctx, model.RoleId.ValueString())
	if err != nil {
		diags.AddError(
			"Error reading role grants after create",
			"Could not verify role grant was created: "+err.Error(),
		)
		return
	}

	grantedRoleID := model.GrantedRoleId.ValueString()
	match := findGrant(grants, grantedRoleID)
	if match.found {
		model.AdminOption = types.BoolValue(match.adminOption)
		if match.grantedRoleName != "" {
			model.GrantedRoleName = types.StringValue(match.grantedRoleName)
		} else {
			model.GrantedRoleName = types.StringNull()
		}
		return
	}

	// Grant wasn't found after create - this shouldn't happen but handle gracefully
	tflog.Warn(ctx, "Grant not found in API response after create", map[string]interface{}{
		"roleId":        model.RoleId.ValueString(),
		"grantedRoleId": grantedRoleID,
	})
	model.GrantedRoleName = types.StringNull()
}
