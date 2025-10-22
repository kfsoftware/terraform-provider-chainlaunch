package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ExternalFabricPeersDataSource{}

func NewExternalFabricPeersDataSource() datasource.DataSource {
	return &ExternalFabricPeersDataSource{}
}

type ExternalFabricPeersDataSource struct {
	client *Client
}

type ExternalFabricPeersDataSourceModel struct {
	Peers []ExternalFabricPeerModel `tfsdk:"peers"`
}

type ExternalFabricPeerModel struct {
	ID               types.Int64  `tfsdk:"id"`
	ExternalNodeID   types.Int64  `tfsdk:"external_node_id"`
	Name             types.String `tfsdk:"name"`
	MSPID            types.String `tfsdk:"msp_id"`
	ExternalEndpoint types.String `tfsdk:"external_endpoint"`
	Version          types.String `tfsdk:"version"`
	SignCertificate  types.String `tfsdk:"sign_certificate"`
	TLSCertificate   types.String `tfsdk:"tls_certificate"`
}

func (d *ExternalFabricPeersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_fabric_peers"
}

func (d *ExternalFabricPeersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "List all external Fabric peers that have been synced from remote Chainlaunch instances.",

		Attributes: map[string]schema.Attribute{
			"peers": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of external Fabric peers",
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
							MarkdownDescription: "External endpoint address",
						},
						"version": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Fabric version",
						},
						"sign_certificate": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Signing certificate",
						},
						"tls_certificate": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "TLS certificate",
						},
					},
				},
			},
		},
	}
}

func (d *ExternalFabricPeersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ExternalFabricPeersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExternalFabricPeersDataSourceModel

	body, err := d.client.DoRequest("GET", "/external-nodes/fabric-peers", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read external Fabric peers, got error: %s", err))
		return
	}

	var peers []map[string]interface{}
	if err := json.Unmarshal(body, &peers); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse peers: %s", err))
		return
	}

	data.Peers = make([]ExternalFabricPeerModel, 0, len(peers))
	for _, peer := range peers {
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

		data.Peers = append(data.Peers, peerModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
