package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ConnectedPeersDataSource{}

func NewConnectedPeersDataSource() datasource.DataSource {
	return &ConnectedPeersDataSource{}
}

type ConnectedPeersDataSource struct {
	client *Client
}

type ConnectedPeersDataSourceModel struct {
	ConnectedPeers []ConnectedPeerModel `tfsdk:"connected_peers"`
}

type ConnectedPeerModel struct {
	ID                types.Int64  `tfsdk:"id"`
	NodeID            types.String `tfsdk:"node_id"`
	PeerEndpoint      types.String `tfsdk:"peer_endpoint"`
	PeerPublicKey     types.String `tfsdk:"peer_public_key"`
	Status            types.String `tfsdk:"status"`
	ConnectedAt       types.String `tfsdk:"connected_at"`
	AnalyticsEnabled  types.Bool   `tfsdk:"analytics_enabled"`
	MetricsFederation types.Bool   `tfsdk:"metrics_federation"`
}

func (d *ConnectedPeersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connected_peers"
}

func (d *ConnectedPeersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Retrieve all connected peer nodes (remote Chainlaunch instances).

This data source returns a list of all peer-to-peer connections between Chainlaunch instances,
including connection IDs needed for network sharing operations.

**Use cases**:
- Get connection IDs for ` + "`chainlaunch_network_share`" + ` recipients
- List all peer-to-peer connections between Chainlaunch instances
- Monitor connection status across the consortium

**Prerequisites**:
- Pro features enabled on the Chainlaunch instance
- Peer connections established via ` + "`chainlaunch_node_invitation`" + ` and ` + "`chainlaunch_node_accept_invitation`" + ``,

		Attributes: map[string]schema.Attribute{
			"connected_peers": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of connected peer nodes",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Connection ID. Use this ID as recipient in `chainlaunch_network_share`.",
						},
						"node_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Remote node's unique identifier",
						},
						"peer_endpoint": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Remote node's endpoint URL",
						},
						"peer_public_key": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Remote node's public key for secure communication",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Connection status (e.g., 'connected', 'disconnected')",
						},
						"connected_at": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Timestamp when the connection was established",
						},
						"analytics_enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether analytics sharing is enabled for this connection",
						},
						"metrics_federation": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether metrics federation is enabled for this connection",
						},
					},
				},
			},
		},
	}
}

func (d *ConnectedPeersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ConnectedPeersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConnectedPeersDataSourceModel

	// Call the /node/connected-peers endpoint
	body, err := d.client.DoRequest("GET", "/node/connected-peers", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read connected peers, got error: %s", err))
		return
	}

	var apiResp ConnectedPeersResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse connected peers response: %s", err))
		return
	}

	// Parse connected peers
	data.ConnectedPeers = make([]ConnectedPeerModel, 0, len(apiResp.ConnectedPeers))
	for _, peer := range apiResp.ConnectedPeers {
		peerModel := ConnectedPeerModel{
			ID:                types.Int64Value(peer.ID),
			NodeID:            types.StringValue(peer.NodeID),
			PeerEndpoint:      types.StringValue(peer.PeerEndpoint),
			PeerPublicKey:     types.StringValue(peer.PeerPublicKey),
			Status:            types.StringValue(peer.Status),
			ConnectedAt:       types.StringValue(peer.ConnectedAt),
			AnalyticsEnabled:  types.BoolValue(peer.AnalyticsEnabled),
			MetricsFederation: types.BoolValue(peer.MetricsFederation),
		}
		data.ConnectedPeers = append(data.ConnectedPeers, peerModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
