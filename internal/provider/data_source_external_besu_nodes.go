package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ExternalBesuNodesDataSource{}

func NewExternalBesuNodesDataSource() datasource.DataSource {
	return &ExternalBesuNodesDataSource{}
}

type ExternalBesuNodesDataSource struct {
	client *Client
}

type ExternalBesuNodesDataSourceModel struct {
	Nodes []ExternalBesuNodeModel `tfsdk:"nodes"`
}

type ExternalBesuNodeModel struct {
	ID             types.Int64  `tfsdk:"id"`
	ExternalNodeID types.Int64  `tfsdk:"external_node_id"`
	Name           types.String `tfsdk:"name"`
	EnodeURL       types.String `tfsdk:"enode_url"`
	P2PHost        types.String `tfsdk:"p2p_host"`
	P2PPort        types.Int64  `tfsdk:"p2p_port"`
	Version        types.String `tfsdk:"version"`
	MetricsEnabled types.Bool   `tfsdk:"metrics_enabled"`
	MetricsPort    types.Int64  `tfsdk:"metrics_port"`
}

func (d *ExternalBesuNodesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_besu_nodes"
}

func (d *ExternalBesuNodesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "List all external Besu nodes that have been synced from remote Chainlaunch instances.",
		MarkdownDescription: "List all external Besu nodes that have been synced from remote Chainlaunch instances.",

		Attributes: map[string]schema.Attribute{
			"nodes": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "List of external Besu nodes",
				MarkdownDescription: "List of external Besu nodes",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:            true,
							Description:         "External Besu node ID",
							MarkdownDescription: "External Besu node ID",
						},
						"external_node_id": schema.Int64Attribute{
							Computed:            true,
							Description:         "ID of the external node this Besu node belongs to",
							MarkdownDescription: "ID of the external node this Besu node belongs to",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							Description:         "Node name",
							MarkdownDescription: "Node name",
						},
						"enode_url": schema.StringAttribute{
							Computed:            true,
							Description:         "Enode URL for P2P connections",
							MarkdownDescription: "Enode URL for P2P connections",
						},
						"p2p_host": schema.StringAttribute{
							Computed:            true,
							Description:         "P2P host address",
							MarkdownDescription: "P2P host address",
						},
						"p2p_port": schema.Int64Attribute{
							Computed:            true,
							Description:         "P2P port number",
							MarkdownDescription: "P2P port number",
						},
						"version": schema.StringAttribute{
							Computed:            true,
							Description:         "Besu version",
							MarkdownDescription: "Besu version",
						},
						"metrics_enabled": schema.BoolAttribute{
							Computed:            true,
							Description:         "Whether metrics are enabled",
							MarkdownDescription: "Whether metrics are enabled",
						},
						"metrics_port": schema.Int64Attribute{
							Computed:            true,
							Description:         "Metrics port number",
							MarkdownDescription: "Metrics port number",
						},
					},
				},
			},
		},
	}
}

func (d *ExternalBesuNodesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ExternalBesuNodesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExternalBesuNodesDataSourceModel

	body, err := d.client.DoRequest("GET", "/external-nodes/besu-nodes", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read external Besu nodes, got error: %s", err))
		return
	}

	var nodes []map[string]interface{}
	if err := json.Unmarshal(body, &nodes); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse Besu nodes: %s", err))
		return
	}

	data.Nodes = make([]ExternalBesuNodeModel, 0, len(nodes))
	for _, node := range nodes {
		nodeModel := ExternalBesuNodeModel{}

		if id, ok := node["id"].(float64); ok {
			nodeModel.ID = types.Int64Value(int64(id))
		}

		if externalNodeID, ok := node["externalNodeId"].(float64); ok {
			nodeModel.ExternalNodeID = types.Int64Value(int64(externalNodeID))
		}

		if name, ok := node["name"].(string); ok {
			nodeModel.Name = types.StringValue(name)
		}

		if enodeURL, ok := node["enodeUrl"].(string); ok {
			nodeModel.EnodeURL = types.StringValue(enodeURL)
		}

		if p2pHost, ok := node["p2pHost"].(string); ok {
			nodeModel.P2PHost = types.StringValue(p2pHost)
		}

		if p2pPort, ok := node["p2pPort"].(float64); ok {
			nodeModel.P2PPort = types.Int64Value(int64(p2pPort))
		}

		if version, ok := node["version"].(string); ok {
			nodeModel.Version = types.StringValue(version)
		}

		if metricsEnabled, ok := node["metricsEnabled"].(bool); ok {
			nodeModel.MetricsEnabled = types.BoolValue(metricsEnabled)
		}

		if metricsPort, ok := node["metricsPort"].(float64); ok {
			nodeModel.MetricsPort = types.Int64Value(int64(metricsPort))
		}

		data.Nodes = append(data.Nodes, nodeModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
