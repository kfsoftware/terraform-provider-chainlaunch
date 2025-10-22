package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &FabricNetworkResource{}
var _ resource.ResourceWithImportState = &FabricNetworkResource{}

func NewFabricNetworkResource() resource.Resource {
	return &FabricNetworkResource{}
}

type FabricNetworkResource struct {
	client *Client
}

type FabricNetworkResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`

	// Organization configuration
	PeerOrganizations    []OrganizationConfigModel `tfsdk:"peer_organizations"`
	OrdererOrganizations []OrganizationConfigModel `tfsdk:"orderer_organizations"`
	ExternalPeerOrgs     []ExternalOrgConfigModel  `tfsdk:"external_peer_orgs"`
	ExternalOrdererOrgs  []ExternalOrgConfigModel  `tfsdk:"external_orderer_orgs"`

	// Consensus configuration
	ConsensusType      types.String             `tfsdk:"consensus_type"`
	SmartBFTOptions    *SmartBFTOptionsModel    `tfsdk:"smartbft_options"`
	SmartBFTConsenters []SmartBFTConsenterModel `tfsdk:"smartbft_consenters"`
	EtcdRaftOptions    *EtcdRaftOptionsModel    `tfsdk:"etcdraft_options"`

	// Capabilities
	ChannelCapabilities     []types.String `tfsdk:"channel_capabilities"`
	ApplicationCapabilities []types.String `tfsdk:"application_capabilities"`
	OrdererCapabilities     []types.String `tfsdk:"orderer_capabilities"`

	// Batch configuration
	BatchSize    *BatchSizeModel `tfsdk:"batch_size"`
	BatchTimeout types.String    `tfsdk:"batch_timeout"`

	// Policies
	ConfigurePolicies   types.Bool                   `tfsdk:"configure_policies"`
	ApplicationPolicies map[string]FabricPolicyModel `tfsdk:"application_policies"`
	OrdererPolicies     map[string]FabricPolicyModel `tfsdk:"orderer_policies"`
	ChannelPolicies     map[string]FabricPolicyModel `tfsdk:"channel_policies"`

	// Computed
	Platform  types.String `tfsdk:"platform"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

type OrganizationConfigModel struct {
	ID      types.Int64   `tfsdk:"id"`
	NodeIDs []types.Int64 `tfsdk:"node_ids"`
}

type ExternalOrgConfigModel struct {
	MSPID      types.String           `tfsdk:"mspid"`
	SignCACert types.String           `tfsdk:"sign_ca_cert"`
	TLSCACert  types.String           `tfsdk:"tls_ca_cert"`
	Consenters []ConsenterConfigModel `tfsdk:"consenters"`
}

type ConsenterConfigModel struct {
	Host    types.String `tfsdk:"host"`
	Port    types.Int64  `tfsdk:"port"`
	TLSCert types.String `tfsdk:"tls_cert"`
}

type BatchSizeModel struct {
	MaxMessageCount   types.Int64 `tfsdk:"max_message_count"`
	AbsoluteMaxBytes  types.Int64 `tfsdk:"absolute_max_bytes"`
	PreferredMaxBytes types.Int64 `tfsdk:"preferred_max_bytes"`
}

type EtcdRaftOptionsModel struct {
	TickInterval         types.String `tfsdk:"tick_interval"`
	ElectionTick         types.Int64  `tfsdk:"election_tick"`
	HeartbeatTick        types.Int64  `tfsdk:"heartbeat_tick"`
	MaxInflightBlocks    types.Int64  `tfsdk:"max_inflight_blocks"`
	SnapshotIntervalSize types.Int64  `tfsdk:"snapshot_interval_size"`
}

type SmartBFTOptionsModel struct {
	RequestBatchMaxCount      types.Int64  `tfsdk:"request_batch_max_count"`
	RequestBatchMaxBytes      types.Int64  `tfsdk:"request_batch_max_bytes"`
	RequestBatchMaxInterval   types.String `tfsdk:"request_batch_max_interval"`
	RequestMaxBytes           types.Int64  `tfsdk:"request_max_bytes"`
	IncomingMessageBufferSize types.Int64  `tfsdk:"incoming_message_buffer_size"`
	RequestPoolSize           types.Int64  `tfsdk:"request_pool_size"`
	ViewChangeResendInterval  types.String `tfsdk:"view_change_resend_interval"`
	ViewChangeTimeout         types.String `tfsdk:"view_change_timeout"`
	LeaderHeartbeatCount      types.Int64  `tfsdk:"leader_heartbeat_count"`
	LeaderHeartbeatTimeout    types.String `tfsdk:"leader_heartbeat_timeout"`
	CollectTimeout            types.String `tfsdk:"collect_timeout"`
	SyncOnStart               types.Bool   `tfsdk:"sync_on_start"`
	SpeedUpViewChange         types.Bool   `tfsdk:"speed_up_view_change"`
	LeaderRotation            types.String `tfsdk:"leader_rotation"`
	DecisionsPerLeader        types.Int64  `tfsdk:"decisions_per_leader"`
	RequestComplainTimeout    types.String `tfsdk:"request_complain_timeout"`
	RequestAutoRemoveTimeout  types.String `tfsdk:"request_auto_remove_timeout"`
	RequestForwardTimeout     types.String `tfsdk:"request_forward_timeout"`
}

type SmartBFTConsenterModel struct {
	ID            types.Int64    `tfsdk:"id"`
	MSPID         types.String   `tfsdk:"mspid"`
	Identity      types.String   `tfsdk:"identity"`
	ClientTLSCert types.String   `tfsdk:"client_tls_cert"`
	ServerTLSCert types.String   `tfsdk:"server_tls_cert"`
	Address       *HostPortModel `tfsdk:"address"`
}

type HostPortModel struct {
	Host types.String `tfsdk:"host"`
	Port types.Int64  `tfsdk:"port"`
}

type FabricPolicyModel struct {
	Type              types.String   `tfsdk:"type"`
	Rule              types.String   `tfsdk:"rule"`
	Organizations     []types.String `tfsdk:"organizations"`
	SignatureOperator types.String   `tfsdk:"signature_operator"`
	SignatureN        types.Int64    `tfsdk:"signature_n"`
}

func (r *FabricNetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_network"
}

func (r *FabricNetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Hyperledger Fabric network (channel) in Chainlaunch with comprehensive configuration options.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier of the network.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the network (channel name).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the network.",
			},

			// Organizations
			"peer_organizations": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of peer organizations to include in the network.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Required:    true,
							Description: "Organization ID.",
						},
						"node_ids": schema.ListAttribute{
							Required:    true,
							ElementType: types.Int64Type,
							Description: "List of peer node IDs for this organization.",
						},
					},
				},
			},
			"orderer_organizations": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of orderer organizations to include in the network.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Required:    true,
							Description: "Organization ID.",
						},
						"node_ids": schema.ListAttribute{
							Required:    true,
							ElementType: types.Int64Type,
							Description: "List of orderer node IDs (consenters) for this organization.",
						},
					},
				},
			},
			"external_peer_orgs": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of external peer organizations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"mspid": schema.StringAttribute{
							Required:    true,
							Description: "MSP ID of the external organization.",
						},
						"sign_ca_cert": schema.StringAttribute{
							Required:    true,
							Description: "Signing CA certificate (PEM format).",
						},
						"tls_ca_cert": schema.StringAttribute{
							Required:    true,
							Description: "TLS CA certificate (PEM format).",
						},
						"consenters": schema.ListNestedAttribute{
							Optional:    true,
							Description: "List of consenters for external orderer organizations.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"host": schema.StringAttribute{
										Required:    true,
										Description: "Consenter hostname.",
									},
									"port": schema.Int64Attribute{
										Required:    true,
										Description: "Consenter port.",
									},
									"tls_cert": schema.StringAttribute{
										Optional:    true,
										Description: "TLS certificate for the consenter.",
									},
								},
							},
						},
					},
				},
			},
			"external_orderer_orgs": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of external orderer organizations.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"mspid": schema.StringAttribute{
							Required:    true,
							Description: "MSP ID of the external organization.",
						},
						"sign_ca_cert": schema.StringAttribute{
							Required:    true,
							Description: "Signing CA certificate (PEM format).",
						},
						"tls_ca_cert": schema.StringAttribute{
							Required:    true,
							Description: "TLS CA certificate (PEM format).",
						},
						"consenters": schema.ListNestedAttribute{
							Required:    true,
							Description: "List of consenters for external orderer organizations.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"host": schema.StringAttribute{
										Required:    true,
										Description: "Consenter hostname.",
									},
									"port": schema.Int64Attribute{
										Required:    true,
										Description: "Consenter port.",
									},
									"tls_cert": schema.StringAttribute{
										Optional:    true,
										Description: "TLS certificate for the consenter.",
									},
								},
							},
						},
					},
				},
			},

			// Consensus
			"consensus_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Consensus type: 'etcdraft' or 'smartbft'. Default: 'etcdraft'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"etcdraft_options": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Etcd Raft consensus options.",
				Attributes: map[string]schema.Attribute{
					"tick_interval": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Tick interval (e.g., '500ms'). Default: '500ms'.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"election_tick": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Election tick. Default: 10.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"heartbeat_tick": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Heartbeat tick. Default: 1.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"max_inflight_blocks": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Maximum in-flight blocks. Default: 5.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"snapshot_interval_size": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Snapshot interval size in bytes. Default: 20971520 (20MB).",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"smartbft_options": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "SmartBFT consensus options (required if consensus_type is 'smartbft').",
				Attributes: map[string]schema.Attribute{
					"request_batch_max_count": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Maximum number of requests in a batch. Default: 100.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"request_batch_max_bytes": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Maximum bytes in a batch. Default: 10485760 (10MB).",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"request_batch_max_interval": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Maximum batch interval (e.g., '50ms'). Default: '50ms'.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"request_max_bytes": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Maximum request bytes. Default: 10485760 (10MB).",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"incoming_message_buffer_size": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Incoming message buffer size. Default: 200.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"request_pool_size": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Request pool size. Default: 400.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"view_change_resend_interval": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "View change resend interval. Default: '5s'.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"view_change_timeout": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "View change timeout. Default: '20s'.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"leader_heartbeat_count": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Leader heartbeat count. Default: 10.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"leader_heartbeat_timeout": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Leader heartbeat timeout. Default: '10s'.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"collect_timeout": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Collect timeout. Default: '1s'.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"sync_on_start": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Sync on start. Default: false.",
					},
					"speed_up_view_change": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Speed up view change. Default: false.",
					},
					"leader_rotation": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Leader rotation setting. Default: 'ROTATION_UNSPECIFIED'.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"decisions_per_leader": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Decisions per leader. Default: 1000.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"request_complain_timeout": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Request complain timeout. Default: '20s'.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"request_auto_remove_timeout": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Request auto remove timeout. Default: '3m'.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"request_forward_timeout": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Request forward timeout. Default: '2s'.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"smartbft_consenters": schema.ListNestedAttribute{
				Optional:    true,
				Description: "SmartBFT consenters (required if consensus_type is 'smartbft').",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Required:    true,
							Description: "Node ID.",
						},
						"mspid": schema.StringAttribute{
							Required:    true,
							Description: "MSP ID.",
						},
						"identity": schema.StringAttribute{
							Required:    true,
							Description: "Identity certificate (PEM format).",
						},
						"client_tls_cert": schema.StringAttribute{
							Required:    true,
							Description: "Client TLS certificate (PEM format).",
						},
						"server_tls_cert": schema.StringAttribute{
							Required:    true,
							Description: "Server TLS certificate (PEM format).",
						},
						"address": schema.SingleNestedAttribute{
							Required:    true,
							Description: "Consenter address.",
							Attributes: map[string]schema.Attribute{
								"host": schema.StringAttribute{
									Required:    true,
									Description: "Hostname.",
								},
								"port": schema.Int64Attribute{
									Required:    true,
									Description: "Port number.",
								},
							},
						},
					},
				},
			},

			// Capabilities
			"channel_capabilities": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Channel capabilities (e.g., ['V2_0']).",
			},
			"application_capabilities": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Application capabilities (e.g., ['V2_0']).",
			},
			"orderer_capabilities": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Orderer capabilities (e.g., ['V2_0']).",
			},

			// Batch configuration
			"batch_size": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Batch size configuration.",
				Attributes: map[string]schema.Attribute{
					"max_message_count": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Maximum messages per batch. Default: 500.",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"absolute_max_bytes": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Absolute maximum batch size in bytes. Default: 103809024 (99MB).",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"preferred_max_bytes": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Description: "Preferred maximum batch size in bytes. Default: 524288 (512KB).",
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"batch_timeout": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Batch timeout (e.g., '2s'). Default: '2s'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			// Policies
			"configure_policies": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether to configure custom policies. Default: false.",
			},
			"application_policies": schema.MapNestedAttribute{
				Optional:    true,
				Description: "Application-level policies.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Policy type: 'ImplicitMeta' or 'Signature'.",
						},
						"rule": schema.StringAttribute{
							Required:    true,
							Description: "Policy rule (e.g., 'MAJORITY Endorsement' for ImplicitMeta).",
						},
						"organizations": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Organizations involved in signature policies.",
						},
						"signature_operator": schema.StringAttribute{
							Optional:    true,
							Description: "Signature operator: 'OR', 'AND', or 'OUTOF'.",
						},
						"signature_n": schema.Int64Attribute{
							Optional:    true,
							Description: "N value for OUTOF operator.",
						},
					},
				},
			},
			"orderer_policies": schema.MapNestedAttribute{
				Optional:    true,
				Description: "Orderer-level policies.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Policy type: 'ImplicitMeta' or 'Signature'.",
						},
						"rule": schema.StringAttribute{
							Required:    true,
							Description: "Policy rule.",
						},
						"organizations": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Organizations involved in signature policies.",
						},
						"signature_operator": schema.StringAttribute{
							Optional:    true,
							Description: "Signature operator: 'OR', 'AND', or 'OUTOF'.",
						},
						"signature_n": schema.Int64Attribute{
							Optional:    true,
							Description: "N value for OUTOF operator.",
						},
					},
				},
			},
			"channel_policies": schema.MapNestedAttribute{
				Optional:    true,
				Description: "Channel-level policies.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required:    true,
							Description: "Policy type: 'ImplicitMeta' or 'Signature'.",
						},
						"rule": schema.StringAttribute{
							Required:    true,
							Description: "Policy rule.",
						},
						"organizations": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Organizations involved in signature policies.",
						},
						"signature_operator": schema.StringAttribute{
							Optional:    true,
							Description: "Signature operator: 'OR', 'AND', or 'OUTOF'.",
						},
						"signature_n": schema.Int64Attribute{
							Optional:    true,
							Description: "N value for OUTOF operator.",
						},
					},
				},
			},

			// Computed fields
			"platform": schema.StringAttribute{
				Computed:    true,
				Description: "Blockchain platform (always 'fabric').",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the network.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the network was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the network was last updated.",
			},
		},
	}
}

func (r *FabricNetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *FabricNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricNetworkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the config object
	config := r.buildFabricNetworkConfig(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := CreateFabricNetworkRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Config:      config,
	}

	body, err := r.client.DoRequest("POST", "/networks/fabric", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Fabric network, got error: %s", err))
		return
	}

	var networkResp FabricNetworkResponse
	if err := json.Unmarshal(body, &networkResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse network response: %s", err))
		return
	}

	// Set computed values
	data.ID = types.Int64Value(networkResp.ID)
	data.Platform = types.StringValue("fabric")
	if networkResp.Status != "" {
		data.Status = types.StringValue(networkResp.Status)
	}
	if networkResp.CreatedAt != "" {
		data.CreatedAt = types.StringValue(networkResp.CreatedAt)
	}

	// Set updated_at - if API doesn't provide it, use created_at
	if networkResp.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(networkResp.UpdatedAt)
	} else if networkResp.CreatedAt != "" {
		// API doesn't always return updated_at during creation, use created_at as fallback
		data.UpdatedAt = types.StringValue(networkResp.CreatedAt)
	}

	// Apply defaults if not set
	r.applyDefaults(&data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricNetworkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/networks/fabric/%d", data.ID.ValueInt64())

	body, err := r.client.DoRequest("GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Fabric network, got error: %s", err))
		return
	}

	var networkResp FabricNetworkResponse
	if err := json.Unmarshal(body, &networkResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse network response: %s", err))
		return
	}

	// Update computed fields
	// Note: API may not return all config fields, preserve from state
	if networkResp.Status != "" {
		data.Status = types.StringValue(networkResp.Status)
	}
	if networkResp.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(networkResp.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricNetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FabricNetworkResourceModel
	var state FabricNetworkResourceModel

	// Get current state to preserve computed fields
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve created_at from state (it's a computed field that never changes)
	data.CreatedAt = state.CreatedAt

	// Build the config object
	config := r.buildFabricNetworkConfig(ctx, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := UpdateFabricNetworkRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Config:      config,
	}

	endpoint := fmt.Sprintf("/networks/fabric/%d", data.ID.ValueInt64())

	body, err := r.client.DoRequest("PUT", endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Fabric network, got error: %s", err))
		return
	}

	var networkResp FabricNetworkResponse
	if err := json.Unmarshal(body, &networkResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse network response: %s", err))
		return
	}

	if networkResp.Status != "" {
		data.Status = types.StringValue(networkResp.Status)
	}
	if networkResp.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(networkResp.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FabricNetworkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/networks/fabric/%d", data.ID.ValueInt64())

	_, err := r.client.DoRequest("DELETE", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Fabric network, got error: %s", err))
		return
	}
}

func (r *FabricNetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to build the config from the model
func (r *FabricNetworkResource) buildFabricNetworkConfig(ctx context.Context, data *FabricNetworkResourceModel, diags *diag.Diagnostics) FabricNetworkConfig {
	config := FabricNetworkConfig{}

	// Organizations
	if len(data.PeerOrganizations) > 0 {
		config.PeerOrganizations = make([]OrganizationConfig, len(data.PeerOrganizations))
		for i, org := range data.PeerOrganizations {
			nodeIDs := make([]int64, len(org.NodeIDs))
			for j, nid := range org.NodeIDs {
				nodeIDs[j] = nid.ValueInt64()
			}
			config.PeerOrganizations[i] = OrganizationConfig{
				ID:      org.ID.ValueInt64(),
				NodeIDs: nodeIDs,
			}
		}
	}

	if len(data.OrdererOrganizations) > 0 {
		config.OrdererOrganizations = make([]OrganizationConfig, len(data.OrdererOrganizations))
		for i, org := range data.OrdererOrganizations {
			nodeIDs := make([]int64, len(org.NodeIDs))
			for j, nid := range org.NodeIDs {
				nodeIDs[j] = nid.ValueInt64()
			}
			config.OrdererOrganizations[i] = OrganizationConfig{
				ID:      org.ID.ValueInt64(),
				NodeIDs: nodeIDs,
			}
		}
	}

	if len(data.ExternalPeerOrgs) > 0 {
		config.ExternalPeerOrgs = make([]ExternalOrgConfig, len(data.ExternalPeerOrgs))
		for i, org := range data.ExternalPeerOrgs {
			extOrg := ExternalOrgConfig{
				MSPID:      org.MSPID.ValueString(),
				SignCACert: org.SignCACert.ValueString(),
				TLSCACert:  org.TLSCACert.ValueString(),
			}
			if len(org.Consenters) > 0 {
				extOrg.Consenters = make([]ConsenterConfig, len(org.Consenters))
				for j, c := range org.Consenters {
					extOrg.Consenters[j] = ConsenterConfig{
						Host:    c.Host.ValueString(),
						Port:    c.Port.ValueInt64(),
						TLSCert: c.TLSCert.ValueString(),
					}
				}
			}
			config.ExternalPeerOrgs[i] = extOrg
		}
	}

	if len(data.ExternalOrdererOrgs) > 0 {
		config.ExternalOrdererOrgs = make([]ExternalOrgConfig, len(data.ExternalOrdererOrgs))
		for i, org := range data.ExternalOrdererOrgs {
			extOrg := ExternalOrgConfig{
				MSPID:      org.MSPID.ValueString(),
				SignCACert: org.SignCACert.ValueString(),
				TLSCACert:  org.TLSCACert.ValueString(),
			}
			if len(org.Consenters) > 0 {
				extOrg.Consenters = make([]ConsenterConfig, len(org.Consenters))
				for j, c := range org.Consenters {
					extOrg.Consenters[j] = ConsenterConfig{
						Host:    c.Host.ValueString(),
						Port:    c.Port.ValueInt64(),
						TLSCert: c.TLSCert.ValueString(),
					}
				}
			}
			config.ExternalOrdererOrgs[i] = extOrg
		}
	}

	// Consensus
	if !data.ConsensusType.IsNull() {
		config.ConsensusType = data.ConsensusType.ValueString()
	}

	if data.EtcdRaftOptions != nil {
		opts := EtcdRaftOptions{}
		if !data.EtcdRaftOptions.TickInterval.IsNull() {
			opts.TickInterval = data.EtcdRaftOptions.TickInterval.ValueString()
		}
		if !data.EtcdRaftOptions.ElectionTick.IsNull() {
			opts.ElectionTick = int(data.EtcdRaftOptions.ElectionTick.ValueInt64())
		}
		if !data.EtcdRaftOptions.HeartbeatTick.IsNull() {
			opts.HeartbeatTick = int(data.EtcdRaftOptions.HeartbeatTick.ValueInt64())
		}
		if !data.EtcdRaftOptions.MaxInflightBlocks.IsNull() {
			opts.MaxInflightBlocks = int(data.EtcdRaftOptions.MaxInflightBlocks.ValueInt64())
		}
		if !data.EtcdRaftOptions.SnapshotIntervalSize.IsNull() {
			opts.SnapshotIntervalSize = int(data.EtcdRaftOptions.SnapshotIntervalSize.ValueInt64())
		}
		config.EtcdRaftOptions = &opts
	}

	if data.SmartBFTOptions != nil {
		opts := SmartBFTOptions{}
		if !data.SmartBFTOptions.RequestBatchMaxCount.IsNull() {
			opts.RequestBatchMaxCount = int(data.SmartBFTOptions.RequestBatchMaxCount.ValueInt64())
		}
		if !data.SmartBFTOptions.RequestBatchMaxBytes.IsNull() {
			opts.RequestBatchMaxBytes = int(data.SmartBFTOptions.RequestBatchMaxBytes.ValueInt64())
		}
		if !data.SmartBFTOptions.RequestBatchMaxInterval.IsNull() {
			opts.RequestBatchMaxInterval = data.SmartBFTOptions.RequestBatchMaxInterval.ValueString()
		}
		if !data.SmartBFTOptions.RequestMaxBytes.IsNull() {
			opts.RequestMaxBytes = int(data.SmartBFTOptions.RequestMaxBytes.ValueInt64())
		}
		if !data.SmartBFTOptions.IncomingMessageBufferSize.IsNull() {
			opts.IncomingMessageBufferSize = int(data.SmartBFTOptions.IncomingMessageBufferSize.ValueInt64())
		}
		if !data.SmartBFTOptions.RequestPoolSize.IsNull() {
			opts.RequestPoolSize = int(data.SmartBFTOptions.RequestPoolSize.ValueInt64())
		}
		if !data.SmartBFTOptions.ViewChangeResendInterval.IsNull() {
			opts.ViewChangeResendInterval = data.SmartBFTOptions.ViewChangeResendInterval.ValueString()
		}
		if !data.SmartBFTOptions.ViewChangeTimeout.IsNull() {
			opts.ViewChangeTimeout = data.SmartBFTOptions.ViewChangeTimeout.ValueString()
		}
		if !data.SmartBFTOptions.LeaderHeartbeatCount.IsNull() {
			opts.LeaderHeartbeatCount = int(data.SmartBFTOptions.LeaderHeartbeatCount.ValueInt64())
		}
		if !data.SmartBFTOptions.LeaderHeartbeatTimeout.IsNull() {
			opts.LeaderHeartbeatTimeout = data.SmartBFTOptions.LeaderHeartbeatTimeout.ValueString()
		}
		if !data.SmartBFTOptions.CollectTimeout.IsNull() {
			opts.CollectTimeout = data.SmartBFTOptions.CollectTimeout.ValueString()
		}
		if !data.SmartBFTOptions.SyncOnStart.IsNull() {
			opts.SyncOnStart = data.SmartBFTOptions.SyncOnStart.ValueBool()
		}
		if !data.SmartBFTOptions.SpeedUpViewChange.IsNull() {
			opts.SpeedUpViewChange = data.SmartBFTOptions.SpeedUpViewChange.ValueBool()
		}
		if !data.SmartBFTOptions.LeaderRotation.IsNull() {
			opts.LeaderRotation = data.SmartBFTOptions.LeaderRotation.ValueString()
		}
		if !data.SmartBFTOptions.DecisionsPerLeader.IsNull() {
			opts.DecisionsPerLeader = int(data.SmartBFTOptions.DecisionsPerLeader.ValueInt64())
		}
		if !data.SmartBFTOptions.RequestComplainTimeout.IsNull() {
			opts.RequestComplainTimeout = data.SmartBFTOptions.RequestComplainTimeout.ValueString()
		}
		if !data.SmartBFTOptions.RequestAutoRemoveTimeout.IsNull() {
			opts.RequestAutoRemoveTimeout = data.SmartBFTOptions.RequestAutoRemoveTimeout.ValueString()
		}
		if !data.SmartBFTOptions.RequestForwardTimeout.IsNull() {
			opts.RequestForwardTimeout = data.SmartBFTOptions.RequestForwardTimeout.ValueString()
		}
		config.SmartBFTOptions = &opts
	}

	if len(data.SmartBFTConsenters) > 0 {
		config.SmartBFTConsenters = make([]SmartBFTConsenter, len(data.SmartBFTConsenters))
		for i, c := range data.SmartBFTConsenters {
			consenter := SmartBFTConsenter{
				ID:            c.ID.ValueInt64(),
				MSPID:         c.MSPID.ValueString(),
				Identity:      c.Identity.ValueString(),
				ClientTLSCert: c.ClientTLSCert.ValueString(),
				ServerTLSCert: c.ServerTLSCert.ValueString(),
			}
			if c.Address != nil {
				consenter.Address = HostPort{
					Host: c.Address.Host.ValueString(),
					Port: int(c.Address.Port.ValueInt64()),
				}
			}
			config.SmartBFTConsenters[i] = consenter
		}
	}

	// Capabilities
	if len(data.ChannelCapabilities) > 0 {
		config.ChannelCapabilities = make([]string, len(data.ChannelCapabilities))
		for i, c := range data.ChannelCapabilities {
			config.ChannelCapabilities[i] = c.ValueString()
		}
	}
	if len(data.ApplicationCapabilities) > 0 {
		config.ApplicationCapabilities = make([]string, len(data.ApplicationCapabilities))
		for i, c := range data.ApplicationCapabilities {
			config.ApplicationCapabilities[i] = c.ValueString()
		}
	}
	if len(data.OrdererCapabilities) > 0 {
		config.OrdererCapabilities = make([]string, len(data.OrdererCapabilities))
		for i, c := range data.OrdererCapabilities {
			config.OrdererCapabilities[i] = c.ValueString()
		}
	}

	// Batch
	if data.BatchSize != nil {
		batchSize := BatchSize{}
		if !data.BatchSize.MaxMessageCount.IsNull() {
			batchSize.MaxMessageCount = int(data.BatchSize.MaxMessageCount.ValueInt64())
		}
		if !data.BatchSize.AbsoluteMaxBytes.IsNull() {
			batchSize.AbsoluteMaxBytes = int(data.BatchSize.AbsoluteMaxBytes.ValueInt64())
		}
		if !data.BatchSize.PreferredMaxBytes.IsNull() {
			batchSize.PreferredMaxBytes = int(data.BatchSize.PreferredMaxBytes.ValueInt64())
		}
		config.BatchSize = &batchSize
	}
	if !data.BatchTimeout.IsNull() {
		config.BatchTimeout = data.BatchTimeout.ValueString()
	}

	// Policies
	if len(data.ApplicationPolicies) > 0 {
		config.ApplicationPolicies = make(map[string]FabricPolicy)
		for k, v := range data.ApplicationPolicies {
			policy := FabricPolicy{
				Type: v.Type.ValueString(),
				Rule: v.Rule.ValueString(),
			}
			config.ApplicationPolicies[k] = policy
		}
	}
	if len(data.OrdererPolicies) > 0 {
		config.OrdererPolicies = make(map[string]FabricPolicy)
		for k, v := range data.OrdererPolicies {
			policy := FabricPolicy{
				Type: v.Type.ValueString(),
				Rule: v.Rule.ValueString(),
			}
			config.OrdererPolicies[k] = policy
		}
	}
	if len(data.ChannelPolicies) > 0 {
		config.ChannelPolicies = make(map[string]FabricPolicy)
		for k, v := range data.ChannelPolicies {
			policy := FabricPolicy{
				Type: v.Type.ValueString(),
				Rule: v.Rule.ValueString(),
			}
			config.ChannelPolicies[k] = policy
		}
	}

	return config
}

// Apply defaults for computed fields
func (r *FabricNetworkResource) applyDefaults(data *FabricNetworkResourceModel) {
	if data.ConsensusType.IsNull() {
		data.ConsensusType = types.StringValue("etcdraft")
	}

	if data.BatchTimeout.IsNull() {
		data.BatchTimeout = types.StringValue("2s")
	}

	if data.BatchSize == nil {
		data.BatchSize = &BatchSizeModel{
			MaxMessageCount:   types.Int64Value(500),
			AbsoluteMaxBytes:  types.Int64Value(103809024),
			PreferredMaxBytes: types.Int64Value(524288),
		}
	} else {
		if data.BatchSize.MaxMessageCount.IsNull() {
			data.BatchSize.MaxMessageCount = types.Int64Value(500)
		}
		if data.BatchSize.AbsoluteMaxBytes.IsNull() {
			data.BatchSize.AbsoluteMaxBytes = types.Int64Value(103809024)
		}
		if data.BatchSize.PreferredMaxBytes.IsNull() {
			data.BatchSize.PreferredMaxBytes = types.Int64Value(524288)
		}
	}

	if data.EtcdRaftOptions != nil {
		if data.EtcdRaftOptions.TickInterval.IsNull() {
			data.EtcdRaftOptions.TickInterval = types.StringValue("500ms")
		}
		if data.EtcdRaftOptions.ElectionTick.IsNull() {
			data.EtcdRaftOptions.ElectionTick = types.Int64Value(10)
		}
		if data.EtcdRaftOptions.HeartbeatTick.IsNull() {
			data.EtcdRaftOptions.HeartbeatTick = types.Int64Value(1)
		}
		if data.EtcdRaftOptions.MaxInflightBlocks.IsNull() {
			data.EtcdRaftOptions.MaxInflightBlocks = types.Int64Value(5)
		}
		if data.EtcdRaftOptions.SnapshotIntervalSize.IsNull() {
			data.EtcdRaftOptions.SnapshotIntervalSize = types.Int64Value(20971520)
		}
	}

	// Apply SmartBFT defaults if SmartBFT is configured
	if data.SmartBFTOptions != nil {
		if data.SmartBFTOptions.RequestBatchMaxCount.IsNull() {
			data.SmartBFTOptions.RequestBatchMaxCount = types.Int64Value(100)
		}
		if data.SmartBFTOptions.RequestBatchMaxBytes.IsNull() {
			data.SmartBFTOptions.RequestBatchMaxBytes = types.Int64Value(10485760)
		}
		if data.SmartBFTOptions.RequestBatchMaxInterval.IsNull() {
			data.SmartBFTOptions.RequestBatchMaxInterval = types.StringValue("50ms")
		}
		if data.SmartBFTOptions.RequestMaxBytes.IsNull() {
			data.SmartBFTOptions.RequestMaxBytes = types.Int64Value(10485760)
		}
		if data.SmartBFTOptions.IncomingMessageBufferSize.IsNull() {
			data.SmartBFTOptions.IncomingMessageBufferSize = types.Int64Value(200)
		}
		if data.SmartBFTOptions.RequestPoolSize.IsNull() {
			data.SmartBFTOptions.RequestPoolSize = types.Int64Value(400)
		}
		if data.SmartBFTOptions.ViewChangeResendInterval.IsNull() {
			data.SmartBFTOptions.ViewChangeResendInterval = types.StringValue("5s")
		}
		if data.SmartBFTOptions.ViewChangeTimeout.IsNull() {
			data.SmartBFTOptions.ViewChangeTimeout = types.StringValue("20s")
		}
		if data.SmartBFTOptions.LeaderHeartbeatCount.IsNull() {
			data.SmartBFTOptions.LeaderHeartbeatCount = types.Int64Value(10)
		}
		if data.SmartBFTOptions.LeaderHeartbeatTimeout.IsNull() {
			data.SmartBFTOptions.LeaderHeartbeatTimeout = types.StringValue("10s")
		}
		if data.SmartBFTOptions.CollectTimeout.IsNull() {
			data.SmartBFTOptions.CollectTimeout = types.StringValue("1s")
		}
		if data.SmartBFTOptions.LeaderRotation.IsNull() {
			data.SmartBFTOptions.LeaderRotation = types.StringValue("ROTATION_UNSPECIFIED")
		}
		if data.SmartBFTOptions.DecisionsPerLeader.IsNull() {
			data.SmartBFTOptions.DecisionsPerLeader = types.Int64Value(1000)
		}
		if data.SmartBFTOptions.RequestComplainTimeout.IsNull() {
			data.SmartBFTOptions.RequestComplainTimeout = types.StringValue("20s")
		}
		if data.SmartBFTOptions.RequestAutoRemoveTimeout.IsNull() {
			data.SmartBFTOptions.RequestAutoRemoveTimeout = types.StringValue("3m")
		}
		if data.SmartBFTOptions.RequestForwardTimeout.IsNull() {
			data.SmartBFTOptions.RequestForwardTimeout = types.StringValue("2s")
		}
	}

	if data.ConfigurePolicies.IsNull() {
		data.ConfigurePolicies = types.BoolValue(false)
	}
}
