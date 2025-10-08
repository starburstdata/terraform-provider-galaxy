package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_cluster"
)

var _ datasource.DataSource = (*clusterDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*clusterDataSource)(nil)

func NewClusterDataSource() datasource.DataSource {
	return &clusterDataSource{}
}

type clusterDataSource struct {
	client *client.GalaxyClient
}

func (d *clusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (d *clusterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_cluster.ClusterDataSourceSchema(ctx)
}

func (d *clusterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *clusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_cluster.ClusterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.Id.ValueString()
	tflog.Debug(ctx, "Reading cluster", map[string]interface{}{"id": id})

	response, err := d.client.GetCluster(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading cluster",
			"Could not read cluster "+id+": "+err.Error(),
		)
		return
	}

	// Map response to model
	d.updateModelFromResponse(ctx, &config, response)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *clusterDataSource) updateModelFromResponse(ctx context.Context, model *datasource_cluster.ClusterModel, response map[string]interface{}) {
	// Map response fields to model based on actual API response structure
	if clusterId, ok := response["clusterId"].(string); ok {
		model.Id = types.StringValue(clusterId)
		model.ClusterId = types.StringValue(clusterId)
	}
	if name, ok := response["name"].(string); ok {
		model.Name = types.StringValue(name)
	}
	if cloudRegionId, ok := response["cloudRegionId"].(string); ok {
		model.CloudRegionId = types.StringValue(cloudRegionId)
	}
	if clusterState, ok := response["clusterState"].(string); ok {
		model.ClusterState = types.StringValue(clusterState)
	}
	if minWorkers, ok := response["minWorkers"].(float64); ok {
		model.MinWorkers = types.Int64Value(int64(minWorkers))
	}
	if maxWorkers, ok := response["maxWorkers"].(float64); ok {
		model.MaxWorkers = types.Int64Value(int64(maxWorkers))
	}
	if idleStopMinutes, ok := response["idleStopMinutes"].(float64); ok {
		model.IdleStopMinutes = types.Int64Value(int64(idleStopMinutes))
	}
	if trinoUri, ok := response["trinoUri"].(string); ok {
		model.TrinoUri = types.StringValue(trinoUri)
	} else {
		// trino_uri is null when cluster is in DISABLED state (data sources cannot return unknown values)
		model.TrinoUri = types.StringNull()
	}
	if batchCluster, ok := response["batchCluster"].(bool); ok {
		model.BatchCluster = types.BoolValue(batchCluster)
	}
	if warpSpeedCluster, ok := response["warpSpeedCluster"].(bool); ok {
		model.WarpSpeedCluster = types.BoolValue(warpSpeedCluster)
	}
	if replicas, ok := response["replicas"].(float64); ok {
		model.Replicas = types.Int64Value(int64(replicas))
	}
	if catalogRefs, ok := response["catalogRefs"].([]interface{}); ok {
		catalogList := make([]types.String, 0, len(catalogRefs))
		for _, ref := range catalogRefs {
			if refStr, ok := ref.(string); ok {
				catalogList = append(catalogList, types.StringValue(refStr))
			}
		}
		catalogListValue, _ := types.ListValueFrom(ctx, types.StringType, catalogList)
		model.CatalogRefs = catalogListValue
	} else {
		model.CatalogRefs = types.ListNull(types.StringType)
	}
}
