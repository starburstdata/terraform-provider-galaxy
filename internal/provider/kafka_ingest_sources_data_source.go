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
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_kafka_ingest_sources"
)

var _ datasource.DataSource = (*kafkaIngestSourcesDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*kafkaIngestSourcesDataSource)(nil)

func NewKafkaIngestSourcesDataSource() datasource.DataSource {
	return &kafkaIngestSourcesDataSource{}
}

type kafkaIngestSourcesDataSource struct {
	client *client.GalaxyClient
}

func (d *kafkaIngestSourcesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_ingest_sources"
}

func (d *kafkaIngestSourcesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_kafka_ingest_sources.KafkaIngestSourcesDataSourceSchema(ctx)
}

func (d *kafkaIngestSourcesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *kafkaIngestSourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_kafka_ingest_sources.KafkaIngestSourcesModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading Kafka ingest sources with automatic pagination")

	// Use automatic pagination to get ALL Kafka ingest sources across all pages
	allSources, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/kafkaIngestSource")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Kafka ingest sources",
			"Could not read Kafka ingest sources: "+err.Error(),
		)
		return
	}

	// Convert []interface{} to []map[string]interface{} for mapping
	var sourceMaps []map[string]interface{}
	for _, sourceInterface := range allSources {
		if sourceMap, ok := sourceInterface.(map[string]interface{}); ok {
			sourceMaps = append(sourceMaps, sourceMap)
		}
	}

	// Map API response to model
	if len(sourceMaps) > 0 {
		sources, err := d.mapKafkaIngestSourcesResult(ctx, sourceMaps)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping Kafka ingest sources response",
				"Could not map Kafka ingest sources response: "+err.Error(),
			)
			return
		}
		config.Result = sources
	} else {
		elementType := datasource_kafka_ingest_sources.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_kafka_ingest_sources.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_kafka_ingest_sources.ResultValue{})
		config.Result = emptyList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

func (d *kafkaIngestSourcesDataSource) mapKafkaIngestSourcesResult(ctx context.Context, result []map[string]interface{}) (types.List, error) {
	sources := make([]datasource_kafka_ingest_sources.ResultValue, 0)

	for _, sourceMap := range result {
		source := d.mapSingleKafkaIngestSource(ctx, sourceMap)
		sources = append(sources, source)
	}

	elementType := datasource_kafka_ingest_sources.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_kafka_ingest_sources.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, sources)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *kafkaIngestSourcesDataSource) mapSingleKafkaIngestSource(ctx context.Context, sourceMap map[string]interface{}) datasource_kafka_ingest_sources.ResultValue {
	attributeTypes := datasource_kafka_ingest_sources.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map kafka_ingest_source_id
	if id, ok := sourceMap["kafkaIngestSourceId"].(string); ok {
		attributes["kafka_ingest_source_id"] = types.StringValue(id)
	} else {
		attributes["kafka_ingest_source_id"] = types.StringNull()
	}

	// Map name
	if name, ok := sourceMap["name"].(string); ok {
		attributes["name"] = types.StringValue(name)
	} else {
		attributes["name"] = types.StringNull()
	}

	// Map description
	if description, ok := sourceMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	// Map kafka_brokers - it's a list of strings
	var kafkaBrokers []string
	if brokers, ok := sourceMap["kafkaBrokers"].([]interface{}); ok {
		for _, broker := range brokers {
			if brokerStr, ok := broker.(string); ok {
				kafkaBrokers = append(kafkaBrokers, brokerStr)
			}
		}
	}

	if len(kafkaBrokers) > 0 {
		brokerValues := make([]attr.Value, len(kafkaBrokers))
		for i, broker := range kafkaBrokers {
			brokerValues[i] = types.StringValue(broker)
		}
		listValue, _ := types.ListValue(types.StringType, brokerValues)
		attributes["kafka_brokers"] = listValue
	} else {
		attributes["kafka_brokers"] = types.ListNull(types.StringType)
	}

	// Create the ResultValue using the constructor
	source, diags := datasource_kafka_ingest_sources.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating Kafka ingest source ResultValue: %v", diags))
		return datasource_kafka_ingest_sources.NewResultValueNull()
	}

	return source
}
