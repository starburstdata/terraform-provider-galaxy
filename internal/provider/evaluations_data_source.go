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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_evaluations"
)

var _ datasource.DataSource = (*evaluationsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*evaluationsDataSource)(nil)

func NewEvaluationsDataSource() datasource.DataSource {
	return &evaluationsDataSource{}
}

type evaluationsDataSource struct {
	client *client.GalaxyClient
}

func (d *evaluationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_evaluations"
}

func (d *evaluationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_evaluations.EvaluationsDataSourceSchema(ctx)
}

func (d *evaluationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *evaluationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_evaluations.EvaluationsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkID := config.DataQualityCheckId.ValueString()
	tflog.Debug(ctx, "Reading evaluations", map[string]interface{}{"data_quality_check_id": checkID})

	response, err := d.client.GetEvaluation(ctx, checkID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading evaluations",
			"Could not read evaluations for data quality check "+checkID+": "+err.Error(),
		)
		return
	}

	// Map top-level scalar fields
	if v, ok := response["dataQualityCheckId"].(string); ok {
		config.DataQualityCheckId = types.StringValue(v)
	}
	if v, ok := response["catalogId"].(string); ok {
		config.CatalogId = types.StringValue(v)
	} else {
		config.CatalogId = types.StringNull()
	}
	if v, ok := response["schemaId"].(string); ok {
		config.SchemaId = types.StringValue(v)
	} else {
		config.SchemaId = types.StringNull()
	}
	if v, ok := response["tableId"].(string); ok {
		config.TableId = types.StringValue(v)
	} else {
		config.TableId = types.StringNull()
	}
	if v, ok := response["name"].(string); ok {
		config.Name = types.StringValue(v)
	} else {
		config.Name = types.StringNull()
	}
	if v, ok := response["description"].(string); ok {
		config.Description = types.StringValue(v)
	} else {
		config.Description = types.StringNull()
	}
	if v, ok := response["kind"].(string); ok {
		config.Kind = types.StringValue(v)
	} else {
		config.Kind = types.StringNull()
	}
	if v, ok := response["category"].(string); ok {
		config.Category = types.StringValue(v)
	} else {
		config.Category = types.StringNull()
	}
	if v, ok := response["severity"].(string); ok {
		config.Severity = types.StringValue(v)
	} else {
		config.Severity = types.StringNull()
	}
	if v, ok := response["query"].(string); ok {
		config.Query = types.StringValue(v)
	} else {
		config.Query = types.StringNull()
	}

	// Map evaluations list
	evaluationsElementType := datasource_evaluations.EvaluationsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_evaluations.EvaluationsValue{}.AttributeTypes(ctx),
		},
	}

	var evaluationItems []interface{}
	if items, ok := response["evaluations"].([]interface{}); ok {
		evaluationItems = items
	} else if items, ok := response["items"].([]interface{}); ok {
		evaluationItems = items
	} else if items, ok := response["result"].([]interface{}); ok {
		evaluationItems = items
	}

	if len(evaluationItems) > 0 {
		evaluationsAttrTypes := datasource_evaluations.EvaluationsValue{}.AttributeTypes(ctx)
		var evalElements []attr.Value
		for _, itemRaw := range evaluationItems {
			if item, ok := itemRaw.(map[string]interface{}); ok {
				evalAttrs := map[string]attr.Value{}
				if v, ok := item["basedOnStatsAt"].(string); ok {
					evalAttrs["based_on_stats_at"] = types.StringValue(v)
				} else {
					evalAttrs["based_on_stats_at"] = types.StringNull()
				}
				if v, ok := item["evaluatedAt"].(string); ok {
					evalAttrs["evaluated_at"] = types.StringValue(v)
				} else {
					evalAttrs["evaluated_at"] = types.StringNull()
				}
				if v, ok := item["predicate"].(string); ok {
					evalAttrs["predicate"] = types.StringValue(v)
				} else {
					evalAttrs["predicate"] = types.StringNull()
				}
				if v, ok := item["status"].(string); ok {
					evalAttrs["status"] = types.StringValue(v)
				} else {
					evalAttrs["status"] = types.StringNull()
				}
				evalValue, d2 := datasource_evaluations.NewEvaluationsValue(evaluationsAttrTypes, evalAttrs)
				resp.Diagnostics.Append(d2...)
				if !d2.HasError() {
					evalElements = append(evalElements, evalValue)
				}
			}
		}
		var ld diag.Diagnostics
		config.Evaluations, ld = types.ListValue(evaluationsElementType, evalElements)
		resp.Diagnostics.Append(ld...)
	} else {
		var ld diag.Diagnostics
		config.Evaluations, ld = types.ListValue(evaluationsElementType, []attr.Value{})
		resp.Diagnostics.Append(ld...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
