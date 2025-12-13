package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SharedNetworksDataSource{}

func NewSharedNetworksDataSource() datasource.DataSource {
	return &SharedNetworksDataSource{}
}

type SharedNetworksDataSource struct {
	client *Client
}

type SharedNetworkModel struct {
	ID             types.String `tfsdk:"id"`
	ResourceID     types.String `tfsdk:"resource_id"`
	ResourceType   types.String `tfsdk:"resource_type"`
	ConnectionID   types.Int64  `tfsdk:"connection_id"`
	SharedBy       types.String `tfsdk:"shared_by"`
	SharedByNodeID types.String `tfsdk:"shared_by_node_id"`
	Status         types.String `tfsdk:"status"`
	CreatedAt      types.String `tfsdk:"created_at"`
	Data           types.Map    `tfsdk:"data"`
}

type SharedNetworksDataSourceModel struct {
	Name     types.String         `tfsdk:"name"`
	Networks []SharedNetworkModel `tfsdk:"networks"`
}

func (d *SharedNetworksDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shared_networks"
}

func (d *SharedNetworksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Query networks that have been shared with this Chainlaunch instance.

This data source retrieves shared networks (Fabric or Besu) that have been shared by other connected Chainlaunch instances.
Use optional name filter to narrow down results.

**Important**: This is a Pro-only feature.

**Workflow**:
1. Connect to peer nodes using ` + "`chainlaunch_node_invitation`" + ` and ` + "`chainlaunch_node_accept_invitation`" + `
2. Peers share networks with you using ` + "`chainlaunch_network_share`" + `
3. Use this data source to list all shared networks
4. Accept networks using ` + "`chainlaunch_shared_network_accept`" + ``,

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Filter by network name (exact match, case-sensitive)",
			},
			"networks": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of shared networks matching the filter criteria",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Unique share ID (use this to accept the share)",
						},
						"resource_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "ID of the network resource",
						},
						"resource_type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Type of resource (fabric_network or besu_network)",
						},
						"connection_id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "ID of the connection used for this share",
						},
						"shared_by": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Description of who shared this network",
						},
						"shared_by_node_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Node ID of who shared this network",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Status of the share (pending, accepted, rejected)",
						},
						"created_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Timestamp when the share was created",
						},
						"data": schema.MapAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Network data as map (contains network configuration, genesis block, name, etc.)",
						},
					},
				},
			},
		},
	}
}

func (d *SharedNetworksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *SharedNetworksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SharedNetworksDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the /pro/shared-networks endpoint
	body, err := d.client.DoRequest("GET", "/pro/shared-networks", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read shared networks, got error: %s", err))
		return
	}

	var apiResp struct {
		Networks []struct {
			ID             string                 `json:"id"`
			ResourceID     string                 `json:"resourceId"`
			ResourceType   string                 `json:"resourceType"`
			ConnectionID   int64                  `json:"connectionId"`
			SharedBy       string                 `json:"sharedBy"`
			SharedByNodeID string                 `json:"sharedByNodeId"`
			Status         string                 `json:"status"`
			CreatedAt      string                 `json:"createdAt"`
			DataJSON       map[string]interface{} `json:"dataJson"`
		} `json:"networks"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse shared networks response: %s", err))
		return
	}

	// Filter and map results
	data.Networks = make([]SharedNetworkModel, 0)

	nameFilter := data.Name.ValueString()

	for _, network := range apiResp.Networks {
		// Apply name filter if specified (check in dataJson for name field)
		if nameFilter != "" {
			networkName := ""
			if dataJSON := network.DataJSON; dataJSON != nil {
				if name, ok := dataJSON["name"].(string); ok {
					networkName = name
				}
			}
			// Skip if name doesn't match exactly (case-sensitive)
			if networkName != nameFilter {
				continue
			}
		}

		// Convert DataJSON map to types.Map with string values
		dataMap := types.MapNull(types.StringType)
		if network.DataJSON != nil {
			mapValues := make(map[string]types.String)
			for key, value := range network.DataJSON {
				// Convert all values to strings
				var strValue string
				switch v := value.(type) {
				case string:
					strValue = v
				case float64:
					strValue = fmt.Sprintf("%v", v)
				case int:
					strValue = fmt.Sprintf("%d", v)
				case bool:
					strValue = fmt.Sprintf("%t", v)
				default:
					// For complex types (objects, arrays), marshal to JSON
					jsonBytes, err := json.Marshal(v)
					if err == nil {
						strValue = string(jsonBytes)
					}
				}
				mapValues[key] = types.StringValue(strValue)
			}
			var diags []error
			dataMap, _ = types.MapValueFrom(ctx, types.StringType, mapValues)
			if len(diags) > 0 {
				dataMap = types.MapNull(types.StringType)
			}
		}

		networkModel := SharedNetworkModel{
			ID:             types.StringValue(network.ID),
			ResourceID:     types.StringValue(network.ResourceID),
			ResourceType:   types.StringValue(network.ResourceType),
			ConnectionID:   types.Int64Value(network.ConnectionID),
			SharedBy:       types.StringValue(network.SharedBy),
			SharedByNodeID: types.StringValue(network.SharedByNodeID),
			Status:         types.StringValue(network.Status),
			CreatedAt:      types.StringValue(network.CreatedAt),
			Data:           dataMap,
		}

		data.Networks = append(data.Networks, networkModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
