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

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_kafka_ingest_source"
)

var _ datasource.DataSource = (*kafkaIngestSourceDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*kafkaIngestSourceDataSource)(nil)

func NewKafkaIngestSourceDataSource() datasource.DataSource {
	return &kafkaIngestSourceDataSource{}
}

type kafkaIngestSourceDataSource struct {
	client *client.GalaxyClient
}

func (d *kafkaIngestSourceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_ingest_source"
}

func (d *kafkaIngestSourceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_kafka_ingest_source.KafkaIngestSourceDataSourceSchema(ctx)
}

func (d *kafkaIngestSourceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.GalaxyClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.GalaxyClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = c
}

func (d *kafkaIngestSourceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_kafka_ingest_source.KafkaIngestSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := config.KafkaIngestSourceId.ValueString()
	response, err := d.client.GetKafkaIngestSource(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading kafka_ingest_source",
			"Could not read kafka_ingest_source "+id+": "+err.Error(),
		)
		return
	}

	if name, ok := response["name"].(string); ok {
		config.Name = types.StringValue(name)
	} else {
		config.Name = types.StringNull()
	}
	if description, ok := response["description"].(string); ok {
		config.Description = types.StringValue(description)
	} else {
		config.Description = types.StringNull()
	}

	if brokers, ok := response["kafkaBrokers"].([]interface{}); ok {
		elems := make([]attr.Value, 0, len(brokers))
		for _, b := range brokers {
			if s, ok := b.(string); ok {
				elems = append(elems, types.StringValue(s))
			}
		}
		listValue, dd := types.ListValue(types.StringType, elems)
		resp.Diagnostics.Append(dd...)
		config.KafkaBrokers = listValue
	} else {
		config.KafkaBrokers = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
