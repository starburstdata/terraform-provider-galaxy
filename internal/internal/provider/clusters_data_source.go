package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_clusters"
)

var _ datasource.DataSource = (*clustersDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*clustersDataSource)(nil)

func NewClustersDataSource() datasource.DataSource {
	return &clustersDataSource{}
}

type clustersDataSource struct {
	client *client.GalaxyClient
}

func (d *clustersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clusters"
}

func (d *clustersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_clusters.ClustersDataSourceSchema(ctx)
}

func (d *clustersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *clustersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_clusters.ClustersModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading clusters with automatic pagination")

	// Use automatic pagination to get ALL clusters across all pages
	allClusters, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/cluster")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading clusters",
			"Could not read clusters: "+err.Error(),
		)
		return
	}

	// Convert []interface{} to []map[string]interface{} for mapping
	var clusterMaps []map[string]interface{}
	for _, clusterInterface := range allClusters {
		if clusterMap, ok := clusterInterface.(map[string]interface{}); ok {
			clusterMaps = append(clusterMaps, clusterMap)
		}
	}

	// Map API response to model
	if len(clusterMaps) > 0 {
		clusters, err := d.mapClustersResult(ctx, clusterMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping clusters response",
				"Could not map clusters response: "+err.Error(),
			)
			return
		}
		config.Result = clusters
	} else {
		elementType := datasource_clusters.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_clusters.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_clusters.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *clustersDataSource) mapClustersResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	clusters := make([]datasource_clusters.ResultValue, 0, len(result))

	for _, clusterMap := range result {
		cluster := d.mapSingleCluster(ctx, clusterMap)
		clusters = append(clusters, cluster)
	}

	elementType := datasource_clusters.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_clusters.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, clusters)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *clustersDataSource) mapSingleCluster(ctx context.Context, clusterMap map[string]interface{}) datasource_clusters.ResultValue {
	attributeTypes := datasource_clusters.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map cluster ID
	if clusterId, ok := clusterMap["clusterId"].(string); ok {
		attributes["cluster_id"] = types.StringValue(clusterId)
	} else {
		attributes["cluster_id"] = types.StringNull()
	}

	// Map name
	if name, ok := clusterMap["name"].(string); ok {
		attributes["name"] = types.StringValue(name)
	} else {
		attributes["name"] = types.StringNull()
	}

	// Map cloud region ID
	if cloudRegionId, ok := clusterMap["cloudRegionId"].(string); ok {
		attributes["cloud_region_id"] = types.StringValue(cloudRegionId)
	} else {
		attributes["cloud_region_id"] = types.StringNull()
	}

	// Map cluster state
	if clusterState, ok := clusterMap["clusterState"].(string); ok {
		attributes["cluster_state"] = types.StringValue(clusterState)
	} else {
		attributes["cluster_state"] = types.StringNull()
	}

	// Map min workers
	if minWorkers, ok := clusterMap["minWorkers"].(float64); ok {
		attributes["min_workers"] = types.Int64Value(int64(minWorkers))
	} else {
		attributes["min_workers"] = types.Int64Null()
	}

	// Map max workers
	if maxWorkers, ok := clusterMap["maxWorkers"].(float64); ok {
		attributes["max_workers"] = types.Int64Value(int64(maxWorkers))
	} else {
		attributes["max_workers"] = types.Int64Null()
	}

	// Map idle stop minutes
	if idleStopMinutes, ok := clusterMap["idleStopMinutes"].(float64); ok {
		attributes["idle_stop_minutes"] = types.Int64Value(int64(idleStopMinutes))
	} else {
		attributes["idle_stop_minutes"] = types.Int64Null()
	}

	// Map trino URI
	if trinoUri, ok := clusterMap["trinoUri"].(string); ok {
		attributes["trino_uri"] = types.StringValue(trinoUri)
	} else {
		attributes["trino_uri"] = types.StringNull()
	}

	// Map batch cluster flag
	if batchCluster, ok := clusterMap["batchCluster"].(bool); ok {
		attributes["batch_cluster"] = types.BoolValue(batchCluster)
	} else {
		attributes["batch_cluster"] = types.BoolNull()
	}

	// Map warp speed cluster flag
	if warpSpeedCluster, ok := clusterMap["warpSpeedCluster"].(bool); ok {
		attributes["warp_speed_cluster"] = types.BoolValue(warpSpeedCluster)
	} else {
		attributes["warp_speed_cluster"] = types.BoolNull()
	}

	// Map catalog refs list
	attributes["catalog_refs"] = d.mapCatalogRefs(ctx, clusterMap)

	// Create the ResultValue using the constructor
	cluster, diags := datasource_clusters.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating cluster ResultValue: %v", diags))
		return datasource_clusters.NewResultValueNull()
	}

	return cluster
}

func (d *clustersDataSource) mapCatalogRefs(ctx context.Context, clusterMap map[string]interface{}) types.List {
	if catalogRefs, ok := clusterMap["catalogRefs"].([]interface{}); ok {
		catalogList := make([]attr.Value, 0, len(catalogRefs))
		for _, ref := range catalogRefs {
			if refStr, ok := ref.(string); ok {
				catalogList = append(catalogList, types.StringValue(refStr))
			}
		}
		catalogListValue, _ := types.ListValue(types.StringType, catalogList)
		return catalogListValue
	} else {
		return types.ListNull(types.StringType)
	}
}
