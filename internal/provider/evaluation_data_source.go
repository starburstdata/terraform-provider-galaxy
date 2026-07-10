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
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_evaluation"
)

var _ datasource.DataSource = (*evaluationDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*evaluationDataSource)(nil)

func NewEvaluationDataSource() datasource.DataSource {
	return &evaluationDataSource{}
}

type evaluationDataSource struct {
	client *client.GalaxyClient
}

func (d *evaluationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_quality_evaluation"
}

func (d *evaluationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_evaluation.EvaluationDataSourceSchema(ctx)
}

func (d *evaluationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *evaluationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError(
			"Unconfigured Provider",
			"The provider has not been properly configured. Please ensure the provider credentials are set.",
		)
		return
	}

	var config datasource_evaluation.EvaluationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	checkId := config.DataQualityCheckId.ValueString()
	tflog.Debug(ctx, "Reading evaluation", map[string]interface{}{"data_quality_check_id": checkId})

	response, err := d.client.GetEvaluation(ctx, checkId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading evaluation",
			"Could not read evaluation for check "+checkId+": "+err.Error(),
		)
		return
	}

	updateDiags := d.updateModelFromResponse(ctx, &config, response)
	resp.Diagnostics.Append(updateDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *evaluationDataSource) updateModelFromResponse(ctx context.Context, model *datasource_evaluation.EvaluationModel, response map[string]interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	if catalogId, ok := response["catalogId"].(string); ok {
		model.CatalogId = types.StringValue(catalogId)
	}
	if category, ok := response["category"].(string); ok {
		model.Category = types.StringValue(category)
	}
	if dataQualityCheckId, ok := response["dataQualityCheckId"].(string); ok {
		model.DataQualityCheckId = types.StringValue(dataQualityCheckId)
	}
	if description, ok := response["description"].(string); ok {
		model.Description = types.StringValue(description)
	}
	if kind, ok := response["kind"].(string); ok {
		model.Kind = types.StringValue(kind)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if query, ok := response["query"].(string); ok {
		model.Query = types.StringValue(query)
	}
	if schemaId, ok := response["schemaId"].(string); ok {
		model.SchemaId = types.StringValue(schemaId)
	}
	if severity, ok := response["severity"].(string); ok {
		model.Severity = types.StringValue(severity)
	}
	if tableId, ok := response["tableId"].(string); ok {
		model.TableId = types.StringValue(tableId)
	}

	// Map evaluations list
	evaluationsType := datasource_evaluation.EvaluationsType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_evaluation.EvaluationsValue{}.AttributeTypes(ctx),
		},
	}
	if evaluations, ok := response["evaluations"].([]interface{}); ok {
		evalsList := make([]datasource_evaluation.EvaluationsValue, 0, len(evaluations))
		for _, evalInterface := range evaluations {
			if evalMap, ok := evalInterface.(map[string]interface{}); ok {
				attributeTypes := datasource_evaluation.EvaluationsValue{}.AttributeTypes(ctx)
				attributes := map[string]attr.Value{}

				if basedOnStatsAt, ok := evalMap["basedOnStatsAt"].(string); ok {
					attributes["based_on_stats_at"] = types.StringValue(basedOnStatsAt)
				} else {
					attributes["based_on_stats_at"] = types.StringNull()
				}
				if evaluatedAt, ok := evalMap["evaluatedAt"].(string); ok {
					attributes["evaluated_at"] = types.StringValue(evaluatedAt)
				} else {
					attributes["evaluated_at"] = types.StringNull()
				}
				if predicate, ok := evalMap["predicate"].(string); ok {
					attributes["predicate"] = types.StringValue(predicate)
				} else {
					attributes["predicate"] = types.StringNull()
				}
				if status, ok := evalMap["status"].(string); ok {
					attributes["status"] = types.StringValue(status)
				} else {
					attributes["status"] = types.StringNull()
				}

				evalValue, d := datasource_evaluation.NewEvaluationsValue(attributeTypes, attributes)
				if d.HasError() {
					diags.Append(d...)
					continue
				}
				evalsList = append(evalsList, evalValue)
			}
		}
		listValue, d := types.ListValueFrom(ctx, evaluationsType, evalsList)
		if d.HasError() {
			diags.Append(d...)
			model.Evaluations = types.ListNull(evaluationsType)
		} else {
			model.Evaluations = listValue
		}
	} else {
		emptyList, _ := types.ListValueFrom(ctx, evaluationsType, []datasource_evaluation.EvaluationsValue{})
		model.Evaluations = emptyList
	}

	return diags
}
