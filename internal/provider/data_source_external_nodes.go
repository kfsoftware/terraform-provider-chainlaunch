package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ExternalNodesDataSource{}

func NewExternalNodesDataSource() datasource.DataSource {
	return &ExternalNodesDataSource{}
}

type ExternalNodesDataSource struct {
	client *Client
}

type ExternalNodesDataSourceModel struct {
	FabricPeers    []ExternalFabricPeerModel    `tfsdk:"fabric_peers"`
	FabricOrderers []ExternalFabricOrdererModel `tfsdk:"fabric_orderers"`
	BesuNodes      []ExternalBesuNodeModel      `tfsdk:"besu_nodes"`
}

// Model types are defined in their respective data source files:
// - ExternalFabricPeerModel in data_source_external_fabric_peers.go
// - ExternalFabricOrdererModel in data_source_external_fabric_orderers.go
// - ExternalBesuNodeModel in data_source_external_besu_nodes.go

func (d *ExternalNodesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_nodes"
}

func (d *ExternalNodesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Retrieve all external nodes (Fabric peers, orderers, and Besu nodes) that have been synced from remote Chainlaunch instances.

This data source provides a comprehensive view of all external nodes in a single API call, including:
- Fabric peers with certificates and endpoints
- Fabric orderers with certificates and endpoints
- Besu nodes with enode URLs and P2P configuration

Use this data source after running ` + "`chainlaunch_sync_all_external_nodes`" + ` or ` + "`chainlaunch_external_nodes_sync`" + ` to retrieve the complete list of synced nodes.

**Benefits over individual data sources**:
- Single API call for all node types
- More efficient than querying each type separately
- Comprehensive view of the entire external node landscape`,

		Attributes: map[string]schema.Attribute{
			"fabric_peers": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of external Fabric peer nodes",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "External peer ID",
						},
						"external_node_id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "ID of the external node this peer belongs to",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Peer name",
						},
						"msp_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "MSP ID of the peer's organization",
						},
						"external_endpoint": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "External endpoint address (e.g., 'peer0.org1.example.com:7051')",
						},
						"version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Hyperledger Fabric version",
						},
						"sign_certificate": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "PEM-encoded signing certificate",
						},
						"tls_certificate": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "PEM-encoded TLS certificate",
						},
					},
				},
			},
			"fabric_orderers": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of external Fabric orderer nodes",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "External orderer ID",
						},
						"external_node_id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "ID of the external node this orderer belongs to",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Orderer name",
						},
						"msp_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "MSP ID of the orderer's organization",
						},
						"external_endpoint": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "External endpoint address (e.g., 'orderer0.org1.example.com:7050')",
						},
						"version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Hyperledger Fabric version",
						},
						"sign_certificate": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "PEM-encoded signing certificate",
						},
						"tls_certificate": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "PEM-encoded TLS certificate",
						},
					},
				},
			},
			"besu_nodes": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of external Hyperledger Besu nodes",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "External Besu node ID",
						},
						"external_node_id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "ID of the external node this Besu node belongs to",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Besu node name",
						},
						"enode_url": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Enode URL for peer-to-peer connections (e.g., 'enode://publickey@host:port')",
						},
						"p2p_host": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "P2P host address",
						},
						"p2p_port": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "P2P port number",
						},
						"version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Hyperledger Besu version",
						},
						"metrics_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether Prometheus metrics are enabled",
						},
						"metrics_port": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Metrics endpoint port (if metrics enabled)",
						},
					},
				},
			},
		},
	}
}

func (d *ExternalNodesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ExternalNodesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExternalNodesDataSourceModel

	// Call the /external-nodes endpoint which returns all node types in one call
	body, err := d.client.DoRequest("GET", "/external-nodes", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read external nodes, got error: %s", err))
		return
	}

	var apiResp struct {
		FabricPeers    []map[string]interface{} `json:"fabric_peers"`
		FabricOrderers []map[string]interface{} `json:"fabric_orderers"`
		BesuNodes      []map[string]interface{} `json:"besu_nodes"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse external nodes response: %s", err))
		return
	}

	// Parse Fabric Peers
	data.FabricPeers = make([]ExternalFabricPeerModel, 0, len(apiResp.FabricPeers))
	for _, peer := range apiResp.FabricPeers {
		peerModel := ExternalFabricPeerModel{}

		if id, ok := peer["id"].(float64); ok {
			peerModel.ID = types.Int64Value(int64(id))
		}
		if externalNodeID, ok := peer["externalNodeId"].(float64); ok {
			peerModel.ExternalNodeID = types.Int64Value(int64(externalNodeID))
		}
		if name, ok := peer["name"].(string); ok {
			peerModel.Name = types.StringValue(name)
		}
		if mspID, ok := peer["mspId"].(string); ok {
			peerModel.MSPID = types.StringValue(mspID)
		}
		if endpoint, ok := peer["externalEndpoint"].(string); ok {
			peerModel.ExternalEndpoint = types.StringValue(endpoint)
		}
		if version, ok := peer["version"].(string); ok {
			peerModel.Version = types.StringValue(version)
		}
		if signCert, ok := peer["signCertificate"].(string); ok {
			peerModel.SignCertificate = types.StringValue(signCert)
		}
		if tlsCert, ok := peer["tlsCertificate"].(string); ok {
			peerModel.TLSCertificate = types.StringValue(tlsCert)
		}

		data.FabricPeers = append(data.FabricPeers, peerModel)
	}

	// Parse Fabric Orderers
	data.FabricOrderers = make([]ExternalFabricOrdererModel, 0, len(apiResp.FabricOrderers))
	for _, orderer := range apiResp.FabricOrderers {
		ordererModel := ExternalFabricOrdererModel{}

		if id, ok := orderer["id"].(float64); ok {
			ordererModel.ID = types.Int64Value(int64(id))
		}
		if externalNodeID, ok := orderer["externalNodeId"].(float64); ok {
			ordererModel.ExternalNodeID = types.Int64Value(int64(externalNodeID))
		}
		if name, ok := orderer["name"].(string); ok {
			ordererModel.Name = types.StringValue(name)
		}
		if mspID, ok := orderer["mspId"].(string); ok {
			ordererModel.MSPID = types.StringValue(mspID)
		}
		if endpoint, ok := orderer["externalEndpoint"].(string); ok {
			ordererModel.ExternalEndpoint = types.StringValue(endpoint)
		}
		if version, ok := orderer["version"].(string); ok {
			ordererModel.Version = types.StringValue(version)
		}
		if signCert, ok := orderer["signCertificate"].(string); ok {
			ordererModel.SignCertificate = types.StringValue(signCert)
		}
		if tlsCert, ok := orderer["tlsCertificate"].(string); ok {
			ordererModel.TLSCertificate = types.StringValue(tlsCert)
		}

		data.FabricOrderers = append(data.FabricOrderers, ordererModel)
	}

	// Parse Besu Nodes
	data.BesuNodes = make([]ExternalBesuNodeModel, 0, len(apiResp.BesuNodes))
	for _, besu := range apiResp.BesuNodes {
		besuModel := ExternalBesuNodeModel{}

		if id, ok := besu["id"].(float64); ok {
			besuModel.ID = types.Int64Value(int64(id))
		}
		if externalNodeID, ok := besu["externalNodeId"].(float64); ok {
			besuModel.ExternalNodeID = types.Int64Value(int64(externalNodeID))
		}
		if name, ok := besu["name"].(string); ok {
			besuModel.Name = types.StringValue(name)
		}
		if enodeURL, ok := besu["enodeUrl"].(string); ok {
			besuModel.EnodeURL = types.StringValue(enodeURL)
		}
		if p2pHost, ok := besu["p2pHost"].(string); ok {
			besuModel.P2PHost = types.StringValue(p2pHost)
		}
		if p2pPort, ok := besu["p2pPort"].(float64); ok {
			besuModel.P2PPort = types.Int64Value(int64(p2pPort))
		}
		if version, ok := besu["version"].(string); ok {
			besuModel.Version = types.StringValue(version)
		}
		if metricsEnabled, ok := besu["metricsEnabled"].(bool); ok {
			besuModel.MetricsEnabled = types.BoolValue(metricsEnabled)
		}
		if metricsPort, ok := besu["metricsPort"].(float64); ok {
			besuModel.MetricsPort = types.Int64Value(int64(metricsPort))
		}

		data.BesuNodes = append(data.BesuNodes, besuModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
