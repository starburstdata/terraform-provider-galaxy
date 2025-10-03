package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/starburstdata/terraform-provider-galaxy/internal/client"
	"github.com/starburstdata/terraform-provider-galaxy/internal/provider/datasource_s3_catalogs"
)

var _ datasource.DataSource = (*s3_catalogsDataSource)(nil)
var _ datasource.DataSourceWithConfigure = (*s3_catalogsDataSource)(nil)

func NewS3CatalogsDataSource() datasource.DataSource {
	return &s3_catalogsDataSource{}
}

type s3_catalogsDataSource struct {
	client *client.GalaxyClient
}

func (d *s3_catalogsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_catalogs"
}

func (d *s3_catalogsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_s3_catalogs.S3CatalogsDataSourceSchema(ctx)
}

func (d *s3_catalogsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *s3_catalogsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config datasource_s3_catalogs.S3CatalogsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "========== STARTING S3_CATALOGS DATA SOURCE READ ==========")
	tflog.Debug(ctx, "Reading s3_catalogs with automatic pagination")

	// Use automatic pagination to get ALL S3 catalogs across all pages
	allCatalogs, err := d.client.GetAllPaginatedResults(ctx, "/public/api/v1/catalogType/s3/catalog")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading s3_catalogs",
			"Could not read s3_catalogs: "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Retrieved %d catalogs from API", len(allCatalogs)))

	// Map API response to model - the API response already has the correct structure
	if len(allCatalogs) > 0 {
		// Create a response map with the "result" field for mapS3CatalogsResult
		response := map[string]interface{}{
			"result": allCatalogs,
		}
		catalogs, err := d.mapS3CatalogsResult(ctx, response)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error mapping s3_catalogs response",
				"Could not map s3_catalogs response: "+err.Error(),
			)
			return
		}
		config.Result = catalogs
	} else {
		elementType := datasource_s3_catalogs.ResultType{
			ObjectType: types.ObjectType{
				AttrTypes: datasource_s3_catalogs.ResultValue{}.AttributeTypes(ctx),
			},
		}
		emptyList, _ := types.ListValueFrom(ctx, elementType, []datasource_s3_catalogs.ResultValue{})
		config.Result = emptyList
	}

	// Return the complete config object that matches the expected schema
	resp.Diagnostics.Append(resp.State.Set(ctx, config)...)
}

func (d *s3_catalogsDataSource) mapS3CatalogsResult(ctx context.Context, response map[string]interface{}) (types.List, error) {
	catalogs := make([]datasource_s3_catalogs.ResultValue, 0)

	var catalogList []interface{}

	tflog.Debug(ctx, fmt.Sprintf("mapS3CatalogsResult input: %+v", response))

	// Handle different response formats
	if items, ok := response["catalogs"].([]interface{}); ok {
		catalogList = items
		tflog.Debug(ctx, fmt.Sprintf("Found %d catalogs in 'catalogs' field", len(items)))
	} else if results, ok := response["results"].([]interface{}); ok {
		catalogList = results
		tflog.Debug(ctx, fmt.Sprintf("Found %d catalogs in 'results' field", len(results)))
	} else if data, ok := response["data"].([]interface{}); ok {
		catalogList = data
		tflog.Debug(ctx, fmt.Sprintf("Found %d catalogs in 'data' field", len(data)))
	} else if result, ok := response["result"].([]interface{}); ok {
		catalogList = result
		tflog.Debug(ctx, fmt.Sprintf("Found %d catalogs in 'result' field", len(result)))
	} else {
		// Debug what's actually in the response
		if resultField, exists := response["result"]; exists {
			tflog.Warn(ctx, fmt.Sprintf("'result' field exists but wrong type: %T, value: %+v", resultField, resultField))
		} else {
			tflog.Warn(ctx, "No 'result' field found in response")
		}
		tflog.Warn(ctx, "No recognized catalog list field found in response")
	}

	for i, catalogInterface := range catalogList {
		if catalogMap, ok := catalogInterface.(map[string]interface{}); ok {
			tflog.Debug(ctx, fmt.Sprintf("Mapping catalog %d: %+v", i, catalogMap))
			catalog := d.mapSingleS3Catalog(ctx, catalogMap)
			catalogs = append(catalogs, catalog)
		} else {
			tflog.Warn(ctx, fmt.Sprintf("Catalog %d is not a map: %T", i, catalogInterface))
		}
	}

	elementType := datasource_s3_catalogs.ResultType{
		ObjectType: types.ObjectType{
			AttrTypes: datasource_s3_catalogs.ResultValue{}.AttributeTypes(ctx),
		},
	}

	listValue, diags := types.ListValueFrom(ctx, elementType, catalogs)
	if diags.HasError() {
		return types.ListNull(elementType), fmt.Errorf("failed to create list value: %v", diags)
	}
	return listValue, nil
}

func (d *s3_catalogsDataSource) mapSingleS3Catalog(ctx context.Context, catalogMap map[string]interface{}) datasource_s3_catalogs.ResultValue {
	tflog.Debug(ctx, fmt.Sprintf("mapSingleS3Catalog input: %+v", catalogMap))

	// Use the generated AttributeTypes method
	attributeTypes := datasource_s3_catalogs.ResultValue{}.AttributeTypes(ctx)
	attributes := map[string]attr.Value{}

	// Map catalog_id
	if catalogId, ok := catalogMap["catalogId"].(string); ok {
		attributes["catalog_id"] = types.StringValue(catalogId)
	} else {
		attributes["catalog_id"] = types.StringNull()
	}

	// Map name
	if name, ok := catalogMap["name"].(string); ok {
		attributes["name"] = types.StringValue(name)
	} else {
		attributes["name"] = types.StringNull()
	}

	// Map description
	if description, ok := catalogMap["description"].(string); ok {
		attributes["description"] = types.StringValue(description)
	} else {
		attributes["description"] = types.StringNull()
	}

	// Map access_key
	if accessKey, ok := catalogMap["accessKey"].(string); ok {
		attributes["access_key"] = types.StringValue(accessKey)
	} else {
		attributes["access_key"] = types.StringNull()
	}

	// Map secret_key
	if secretKey, ok := catalogMap["secretKey"].(string); ok {
		attributes["secret_key"] = types.StringValue(secretKey)
	} else {
		attributes["secret_key"] = types.StringNull()
	}

	// Map region
	if region, ok := catalogMap["region"].(string); ok {
		attributes["region"] = types.StringValue(region)
	} else {
		attributes["region"] = types.StringNull()
	}

	// Map default_bucket
	if defaultBucket, ok := catalogMap["defaultBucket"].(string); ok {
		attributes["default_bucket"] = types.StringValue(defaultBucket)
	} else {
		attributes["default_bucket"] = types.StringNull()
	}

	// Map default_data_location
	if defaultDataLocation, ok := catalogMap["defaultDataLocation"].(string); ok {
		attributes["default_data_location"] = types.StringValue(defaultDataLocation)
	} else {
		attributes["default_data_location"] = types.StringNull()
	}

	// Map metastore_type
	if metastoreType, ok := catalogMap["metastoreType"].(string); ok {
		attributes["metastore_type"] = types.StringValue(metastoreType)
	} else {
		attributes["metastore_type"] = types.StringNull()
	}

	// Map role_arn
	if roleArn, ok := catalogMap["roleArn"].(string); ok {
		attributes["role_arn"] = types.StringValue(roleArn)
	} else {
		attributes["role_arn"] = types.StringNull()
	}

	// Map default_table_format
	if defaultTableFormat, ok := catalogMap["defaultTableFormat"].(string); ok {
		attributes["default_table_format"] = types.StringValue(defaultTableFormat)
	} else {
		attributes["default_table_format"] = types.StringNull()
	}

	// Map external_table_creation_enabled
	if externalTableCreationEnabled, ok := catalogMap["externalTableCreationEnabled"].(bool); ok {
		attributes["external_table_creation_enabled"] = types.BoolValue(externalTableCreationEnabled)
	} else {
		attributes["external_table_creation_enabled"] = types.BoolNull()
	}

	// Map external_table_writes_enabled
	if externalTableWritesEnabled, ok := catalogMap["externalTableWritesEnabled"].(bool); ok {
		attributes["external_table_writes_enabled"] = types.BoolValue(externalTableWritesEnabled)
	} else {
		attributes["external_table_writes_enabled"] = types.BoolNull()
	}

	// Map read_only
	if readOnly, ok := catalogMap["readOnly"].(bool); ok {
		attributes["read_only"] = types.BoolValue(readOnly)
	} else {
		attributes["read_only"] = types.BoolNull()
	}

	// Map glue_access_key
	if glueAccessKey, ok := catalogMap["glueAccessKey"].(string); ok {
		attributes["glue_access_key"] = types.StringValue(glueAccessKey)
	} else {
		attributes["glue_access_key"] = types.StringNull()
	}

	// Map glue_secret_key
	if glueSecretKey, ok := catalogMap["glueSecretKey"].(string); ok {
		attributes["glue_secret_key"] = types.StringValue(glueSecretKey)
	} else {
		attributes["glue_secret_key"] = types.StringNull()
	}

	// Map glue_role_arn
	if glueRoleArn, ok := catalogMap["glueRoleArn"].(string); ok {
		attributes["glue_role_arn"] = types.StringValue(glueRoleArn)
	} else {
		attributes["glue_role_arn"] = types.StringNull()
	}

	// Map hive_metastore_host
	if hiveMetastoreHost, ok := catalogMap["hiveMetastoreHost"].(string); ok {
		attributes["hive_metastore_host"] = types.StringValue(hiveMetastoreHost)
	} else {
		attributes["hive_metastore_host"] = types.StringNull()
	}

	// Map hive_metastore_port
	if hiveMetastorePort, ok := catalogMap["hiveMetastorePort"].(float64); ok {
		attributes["hive_metastore_port"] = types.Int64Value(int64(hiveMetastorePort))
	} else if hiveMetastorePort, ok := catalogMap["hiveMetastorePort"].(int64); ok {
		attributes["hive_metastore_port"] = types.Int64Value(hiveMetastorePort)
	} else {
		attributes["hive_metastore_port"] = types.Int64Null()
	}

	// Map ssh_tunnel_id
	if sshTunnelId, ok := catalogMap["sshTunnelId"].(string); ok {
		attributes["ssh_tunnel_id"] = types.StringValue(sshTunnelId)
	} else {
		attributes["ssh_tunnel_id"] = types.StringNull()
	}

	// Handle validate field
	if validate, ok := catalogMap["validate"].(bool); ok {
		attributes["validate"] = types.BoolValue(validate)
	} else {
		attributes["validate"] = types.BoolNull()
	}
	// Create the ResultValue using the constructor
	catalog, diags := datasource_s3_catalogs.NewResultValue(attributeTypes, attributes)
	if diags.HasError() {
		tflog.Error(ctx, fmt.Sprintf("Error creating ResultValue: %v", diags))
		return datasource_s3_catalogs.NewResultValueNull()
	}

	tflog.Debug(ctx, "mapSingleS3Catalog result created successfully")
	return catalog
}
