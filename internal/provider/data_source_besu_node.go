package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &BesuNodeDataSource{}

func NewBesuNodeDataSource() datasource.DataSource {
	return &BesuNodeDataSource{}
}

// BesuNodeDataSource defines the data source implementation.
type BesuNodeDataSource struct {
	client *Client
}

// BesuNodeDataSourceModel describes the data source data model.
type BesuNodeDataSourceModel struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	NetworkID                  types.Int64  `tfsdk:"network_id"`
	KeyID                      types.Int64  `tfsdk:"key_id"`
	Mode                       types.String `tfsdk:"mode"`
	Version                    types.String `tfsdk:"version"`
	ExternalIP                 types.String `tfsdk:"external_ip"`
	InternalIP                 types.String `tfsdk:"internal_ip"`
	P2PHost                    types.String `tfsdk:"p2p_host"`
	P2PPort                    types.Int64  `tfsdk:"p2p_port"`
	RPCHost                    types.String `tfsdk:"rpc_host"`
	RPCPort                    types.Int64  `tfsdk:"rpc_port"`
	MinGasPrice                types.Int64  `tfsdk:"min_gas_price"`
	HostAllowList              types.String `tfsdk:"host_allow_list"`
	MetricsEnabled             types.Bool   `tfsdk:"metrics_enabled"`
	MetricsPort                types.Int64  `tfsdk:"metrics_port"`
	MetricsProtocol            types.String `tfsdk:"metrics_protocol"`
	JWTEnabled                 types.Bool   `tfsdk:"jwt_enabled"`
	JWTAuthenticationAlgorithm types.String `tfsdk:"jwt_authentication_algorithm"`
	Status                     types.String `tfsdk:"status"`
	CreatedAt                  types.String `tfsdk:"created_at"`
	UpdatedAt                  types.String `tfsdk:"updated_at"`
}

func (d *BesuNodeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_besu_node"
}

func (d *BesuNodeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Fetches details about a specific Hyperledger Besu node from Chainlaunch.",
		MarkdownDescription: "Fetches details about a specific Hyperledger Besu node from Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:            true,
				Description:         "The unique identifier of the Besu node.",
				MarkdownDescription: "The unique identifier of the Besu node.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				Description:         "The name of the Besu node.",
				MarkdownDescription: "The name of the Besu node.",
			},
			"network_id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The ID of the Besu network this node belongs to.",
				MarkdownDescription: "The ID of the Besu network this node belongs to.",
			},
			"key_id": schema.Int64Attribute{
				Computed:            true,
				Description:         "The ID of the cryptographic key for this node.",
				MarkdownDescription: "The ID of the cryptographic key for this node.",
			},
			"mode": schema.StringAttribute{
				Computed:            true,
				Description:         "Deployment mode (docker or service).",
				MarkdownDescription: "Deployment mode (docker or service).",
			},
			"version": schema.StringAttribute{
				Computed:            true,
				Description:         "Besu version.",
				MarkdownDescription: "Besu version.",
			},
			"external_ip": schema.StringAttribute{
				Computed:            true,
				Description:         "External IP address for the node.",
				MarkdownDescription: "External IP address for the node.",
			},
			"internal_ip": schema.StringAttribute{
				Computed:            true,
				Description:         "Internal IP address for the node.",
				MarkdownDescription: "Internal IP address for the node.",
			},
			"p2p_host": schema.StringAttribute{
				Computed:            true,
				Description:         "P2P host address.",
				MarkdownDescription: "P2P host address.",
			},
			"p2p_port": schema.Int64Attribute{
				Computed:            true,
				Description:         "P2P port number.",
				MarkdownDescription: "P2P port number.",
			},
			"rpc_host": schema.StringAttribute{
				Computed:            true,
				Description:         "RPC host address.",
				MarkdownDescription: "RPC host address.",
			},
			"rpc_port": schema.Int64Attribute{
				Computed:            true,
				Description:         "RPC port number.",
				MarkdownDescription: "RPC port number.",
			},
			"min_gas_price": schema.Int64Attribute{
				Computed:            true,
				Description:         "Minimum gas price in Wei.",
				MarkdownDescription: "Minimum gas price in Wei.",
			},
			"host_allow_list": schema.StringAttribute{
				Computed:            true,
				Description:         "Comma-separated list of hostnames allowed to access the RPC API.",
				MarkdownDescription: "Comma-separated list of hostnames allowed to access the RPC API.",
			},
			"metrics_enabled": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether metrics are enabled.",
				MarkdownDescription: "Whether metrics are enabled.",
			},
			"metrics_port": schema.Int64Attribute{
				Computed:            true,
				Description:         "Port for metrics endpoint.",
				MarkdownDescription: "Port for metrics endpoint.",
			},
			"metrics_protocol": schema.StringAttribute{
				Computed:            true,
				Description:         "Protocol for metrics.",
				MarkdownDescription: "Protocol for metrics.",
			},
			"jwt_enabled": schema.BoolAttribute{
				Computed:            true,
				Description:         "Whether JWT authentication is enabled.",
				MarkdownDescription: "Whether JWT authentication is enabled.",
			},
			"jwt_authentication_algorithm": schema.StringAttribute{
				Computed:            true,
				Description:         "JWT authentication algorithm.",
				MarkdownDescription: "JWT authentication algorithm.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				Description:         "Current status of the node (RUNNING, CREATING, ERROR, etc.).",
				MarkdownDescription: "Current status of the node (RUNNING, CREATING, ERROR, etc.).",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when the node was created.",
				MarkdownDescription: "The timestamp when the node was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when the node was last updated.",
				MarkdownDescription: "The timestamp when the node was last updated.",
			},
		},
	}
}

func (d *BesuNodeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *BesuNodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BesuNodeDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.DoRequest("GET", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Besu node, got error: %s", err))
		return
	}

	var nodeResp BesuNodeResponse
	if err := json.Unmarshal(body, &nodeResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse Besu node response, got error: %s", err))
		return
	}

	data.Name = types.StringValue(nodeResp.Name)
	if nodeResp.NetworkID > 0 {
		data.NetworkID = types.Int64Value(nodeResp.NetworkID)
	}
	if nodeResp.KeyID > 0 {
		data.KeyID = types.Int64Value(nodeResp.KeyID)
	}
	if nodeResp.Mode != "" {
		data.Mode = types.StringValue(nodeResp.Mode)
	}
	if nodeResp.Version != "" {
		data.Version = types.StringValue(nodeResp.Version)
	}
	if nodeResp.ExternalIP != "" {
		data.ExternalIP = types.StringValue(nodeResp.ExternalIP)
	}
	if nodeResp.InternalIP != "" {
		data.InternalIP = types.StringValue(nodeResp.InternalIP)
	}
	if nodeResp.P2PHost != "" {
		data.P2PHost = types.StringValue(nodeResp.P2PHost)
	}
	if nodeResp.P2PPort > 0 {
		data.P2PPort = types.Int64Value(nodeResp.P2PPort)
	}
	if nodeResp.RPCHost != "" {
		data.RPCHost = types.StringValue(nodeResp.RPCHost)
	}
	if nodeResp.RPCPort > 0 {
		data.RPCPort = types.Int64Value(nodeResp.RPCPort)
	}
	if nodeResp.MinGasPrice > 0 {
		data.MinGasPrice = types.Int64Value(nodeResp.MinGasPrice)
	}
	if nodeResp.HostAllowList != "" {
		data.HostAllowList = types.StringValue(nodeResp.HostAllowList)
	}
	data.MetricsEnabled = types.BoolValue(nodeResp.MetricsEnabled)
	if nodeResp.MetricsPort > 0 {
		data.MetricsPort = types.Int64Value(nodeResp.MetricsPort)
	}
	if nodeResp.MetricsProtocol != "" {
		data.MetricsProtocol = types.StringValue(nodeResp.MetricsProtocol)
	}
	data.JWTEnabled = types.BoolValue(nodeResp.JWTEnabled)
	if nodeResp.JWTAuthenticationAlgorithm != "" {
		data.JWTAuthenticationAlgorithm = types.StringValue(nodeResp.JWTAuthenticationAlgorithm)
	}
	if nodeResp.Status != "" {
		data.Status = types.StringValue(nodeResp.Status)
	}
	if nodeResp.CreatedAt != "" {
		data.CreatedAt = types.StringValue(nodeResp.CreatedAt)
	}
	if nodeResp.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(nodeResp.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
