package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &BesuNodeResource{}
var _ resource.ResourceWithImportState = &BesuNodeResource{}

func NewBesuNodeResource() resource.Resource {
	return &BesuNodeResource{}
}

// BesuNodeResource defines the resource implementation.
type BesuNodeResource struct {
	client *Client
}

// BesuNodeResourceModel describes the resource data model.
type BesuNodeResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	NetworkID   types.Int64  `tfsdk:"network_id"`
	KeyID       types.Int64  `tfsdk:"key_id"`
	Mode        types.String `tfsdk:"mode"`
	Version     types.String `tfsdk:"version"`
	ExternalIP  types.String `tfsdk:"external_ip"`
	InternalIP  types.String `tfsdk:"internal_ip"`
	P2PHost     types.String `tfsdk:"p2p_host"`
	P2PPort     types.Int64  `tfsdk:"p2p_port"`
	RPCHost     types.String `tfsdk:"rpc_host"`
	RPCPort     types.Int64  `tfsdk:"rpc_port"`
	BootNodes   types.List   `tfsdk:"boot_nodes"`
	MinGasPrice types.Int64  `tfsdk:"min_gas_price"`
	// Permissions
	AccountsAllowList types.List   `tfsdk:"accounts_allow_list"`
	NodesAllowList    types.List   `tfsdk:"nodes_allow_list"`
	HostAllowList     types.String `tfsdk:"host_allow_list"`
	// Metrics
	MetricsEnabled  types.Bool   `tfsdk:"metrics_enabled"`
	MetricsPort     types.Int64  `tfsdk:"metrics_port"`
	MetricsProtocol types.String `tfsdk:"metrics_protocol"`
	// JWT Authentication
	JWTEnabled                 types.Bool   `tfsdk:"jwt_enabled"`
	JWTAuthenticationAlgorithm types.String `tfsdk:"jwt_authentication_algorithm"`
	JWTPublicKeyContent        types.String `tfsdk:"jwt_public_key_content"`
	// Environment variables
	Environment types.Map `tfsdk:"environment"`
	// Computed
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

// BesuNodeResponse represents the API response for a Besu node
type BesuNodeResponse struct {
	ID                         int64             `json:"id"`
	Name                       string            `json:"name"`
	NetworkID                  int64             `json:"networkId"`
	KeyID                      int64             `json:"keyId"`
	Mode                       string            `json:"mode"`
	Version                    string            `json:"version"`
	ExternalIP                 string            `json:"externalIp"`
	InternalIP                 string            `json:"internalIp"`
	P2PHost                    string            `json:"p2pHost"`
	P2PPort                    int64             `json:"p2pPort"`
	RPCHost                    string            `json:"rpcHost"`
	RPCPort                    int64             `json:"rpcPort"`
	BootNodes                  []string          `json:"bootNodes"`
	MinGasPrice                int64             `json:"minGasPrice"`
	AccountsAllowList          []string          `json:"accountsAllowList"`
	NodesAllowList             []string          `json:"nodesAllowList"`
	HostAllowList              string            `json:"hostAllowList"`
	MetricsEnabled             bool              `json:"metricsEnabled"`
	MetricsPort                int64             `json:"metricsPort"`
	MetricsProtocol            string            `json:"metricsProtocol"`
	JWTEnabled                 bool              `json:"jwtEnabled"`
	JWTAuthenticationAlgorithm string            `json:"jwtAuthenticationAlgorithm"`
	JWTPublicKeyContent        string            `json:"jwtPublicKeyContent"`
	Environment                map[string]string `json:"env"`
	Status                     string            `json:"status"`
	CreatedAt                  string            `json:"createdAt"`
	UpdatedAt                  string            `json:"updatedAt"`
}

func (r *BesuNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_besu_node"
}

func (r *BesuNodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Hyperledger Besu node in Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the Besu node.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Besu node.",
			},
			"network_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the Besu network this node belongs to.",
			},
			"key_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the cryptographic key for this node.",
			},
			"mode": schema.StringAttribute{
				Required:    true,
				Description: "Deployment mode (docker or service).",
			},
			"version": schema.StringAttribute{
				Optional:    true,
				Description: "Besu version (e.g., 24.5.1).",
			},
			"external_ip": schema.StringAttribute{
				Required:    true,
				Description: "External IP address for the node.",
			},
			"internal_ip": schema.StringAttribute{
				Required:    true,
				Description: "Internal IP address for the node.",
			},
			"p2p_host": schema.StringAttribute{
				Required:    true,
				Description: "P2P host address (e.g., 0.0.0.0).",
			},
			"p2p_port": schema.Int64Attribute{
				Required:    true,
				Description: "P2P port number (e.g., 30303).",
			},
			"rpc_host": schema.StringAttribute{
				Required:    true,
				Description: "RPC host address (e.g., 0.0.0.0).",
			},
			"rpc_port": schema.Int64Attribute{
				Required:    true,
				Description: "RPC port number (e.g., 8545).",
			},
			"boot_nodes": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of boot node enode URLs.",
			},
			"min_gas_price": schema.Int64Attribute{
				Optional:    true,
				Description: "Minimum gas price in Wei.",
			},
			"accounts_allow_list": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of accounts allowed to participate (for permissioned networks).",
			},
			"nodes_allow_list": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of node enode URLs allowed to connect (for permissioned networks).",
			},
			"host_allow_list": schema.StringAttribute{
				Optional:    true,
				Description: "Comma-separated list of hostnames allowed to access the RPC API.",
			},
			"metrics_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to enable metrics.",
			},
			"metrics_port": schema.Int64Attribute{
				Optional:    true,
				Description: "Port for metrics endpoint.",
			},
			"metrics_protocol": schema.StringAttribute{
				Optional:    true,
				Description: "Protocol for metrics (e.g., prometheus).",
			},
			"jwt_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to enable JWT authentication for engine API.",
			},
			"jwt_authentication_algorithm": schema.StringAttribute{
				Optional:    true,
				Description: "JWT authentication algorithm (e.g., RS256, HS256).",
			},
			"jwt_public_key_content": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "JWT public key content for verification.",
			},
			"environment": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Environment variables for the Besu node.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the node (RUNNING, CREATING, ERROR, etc.).",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the node was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the node was last updated.",
			},
		},
	}
}

func (r *BesuNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BesuNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BesuNodeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the besuNode config object (as per API schema)
	besuNodeConfig := map[string]interface{}{
		"type":       "besu",
		"networkId":  data.NetworkID.ValueInt64(),
		"keyId":      data.KeyID.ValueInt64(),
		"mode":       data.Mode.ValueString(),
		"externalIp": data.ExternalIP.ValueString(),
		"internalIp": data.InternalIP.ValueString(),
		"p2pHost":    data.P2PHost.ValueString(),
		"p2pPort":    data.P2PPort.ValueInt64(),
		"rpcHost":    data.RPCHost.ValueString(),
		"rpcPort":    data.RPCPort.ValueInt64(),
	}

	if !data.Version.IsNull() {
		besuNodeConfig["version"] = data.Version.ValueString()
	}

	// Optional fields
	if !data.BootNodes.IsNull() {
		var bootNodes []string
		data.BootNodes.ElementsAs(ctx, &bootNodes, false)
		besuNodeConfig["bootNodes"] = bootNodes
	}

	if !data.MinGasPrice.IsNull() {
		besuNodeConfig["minGasPrice"] = data.MinGasPrice.ValueInt64()
	}

	if !data.AccountsAllowList.IsNull() {
		var accountsAllowList []string
		data.AccountsAllowList.ElementsAs(ctx, &accountsAllowList, false)
		besuNodeConfig["accountsAllowList"] = accountsAllowList
	}

	if !data.NodesAllowList.IsNull() {
		var nodesAllowList []string
		data.NodesAllowList.ElementsAs(ctx, &nodesAllowList, false)
		besuNodeConfig["nodesAllowList"] = nodesAllowList
	}

	if !data.HostAllowList.IsNull() {
		besuNodeConfig["hostAllowList"] = data.HostAllowList.ValueString()
	}

	if !data.MetricsEnabled.IsNull() {
		besuNodeConfig["metricsEnabled"] = data.MetricsEnabled.ValueBool()
	}

	if !data.MetricsPort.IsNull() {
		besuNodeConfig["metricsPort"] = data.MetricsPort.ValueInt64()
	}

	if !data.MetricsProtocol.IsNull() {
		besuNodeConfig["metricsProtocol"] = data.MetricsProtocol.ValueString()
	}

	if !data.JWTEnabled.IsNull() {
		besuNodeConfig["jwtEnabled"] = data.JWTEnabled.ValueBool()
	}

	if !data.JWTAuthenticationAlgorithm.IsNull() {
		besuNodeConfig["jwtAuthenticationAlgorithm"] = data.JWTAuthenticationAlgorithm.ValueString()
	}

	if !data.JWTPublicKeyContent.IsNull() {
		besuNodeConfig["jwtPublicKeyContent"] = data.JWTPublicKeyContent.ValueString()
	}

	if !data.Environment.IsNull() {
		var env map[string]string
		data.Environment.ElementsAs(ctx, &env, false)
		besuNodeConfig["env"] = env
	}

	// Build the create request according to API schema
	// The API expects: blockchainPlatform (required) and besuNode object
	createReq := map[string]interface{}{
		"name":               data.Name.ValueString(),
		"blockchainPlatform": "BESU",
		"besuNode":           besuNodeConfig,
	}

	body, err := r.client.DoRequest("POST", "/nodes", createReq)
	if err != nil {
		// Parse response - even if there's an error, the response might contain node data
		var errorResponse struct {
			Message string `json:"message"`
			Data    struct {
				Node   *BesuNodeResponse `json:"node"`
				NodeID int64             `json:"node_id"`
				Stage  string            `json:"stage"`
			} `json:"data"`
		}

		// Extract the JSON from the error message
		errMsg := err.Error()
		jsonStart := strings.Index(errMsg, "{")
		if jsonStart >= 0 {
			jsonData := errMsg[jsonStart:]
			if unmarshalErr := json.Unmarshal([]byte(jsonData), &errorResponse); unmarshalErr == nil {
				// If error response contains node data, use it
				if errorResponse.Data.Node != nil && errorResponse.Data.Node.ID > 0 {
					nodeResp := errorResponse.Data.Node
					// Save to state even though there was an error
					data.ID = types.StringValue(fmt.Sprintf("%d", nodeResp.ID))
					data.Name = types.StringValue(nodeResp.Name)
					if nodeResp.Status != "" {
						data.Status = types.StringValue(nodeResp.Status)
					}
					if nodeResp.CreatedAt != "" {
						data.CreatedAt = types.StringValue(nodeResp.CreatedAt)
					}
					if nodeResp.UpdatedAt != "" {
						data.UpdatedAt = types.StringValue(nodeResp.UpdatedAt)
					}

					// Save to state
					resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

					// Add a warning instead of error so state is saved
					resp.Diagnostics.AddWarning(
						"Besu Node Created with Errors",
						fmt.Sprintf("Besu node was created (ID: %d) but failed during startup: %s. The node is in state and can be updated or deleted.", nodeResp.ID, errorResponse.Message),
					)
					return
				}
			}
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Besu node, got error: %s", err))
		return
	}

	var nodeResp BesuNodeResponse
	if err := json.Unmarshal(body, &nodeResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse Besu node response, got error: %s", err))
		return
	}

	// Set values from response
	data.ID = types.StringValue(fmt.Sprintf("%d", nodeResp.ID))
	data.Name = types.StringValue(nodeResp.Name)
	if nodeResp.Status != "" {
		data.Status = types.StringValue(nodeResp.Status)
	}
	if nodeResp.CreatedAt != "" {
		data.CreatedAt = types.StringValue(nodeResp.CreatedAt)
	}
	if nodeResp.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(nodeResp.UpdatedAt)
	}

	// Wait for node to reach RUNNING status
	if err := r.waitForNodeRunning(ctx, nodeResp.ID); err != nil {
		resp.Diagnostics.AddWarning(
			"Node Status Check",
			fmt.Sprintf("Besu node created but status check failed: %s", err),
		)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BesuNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BesuNodeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.DoRequest("GET", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), nil)
	if err != nil {
		// Check for 404 - node was deleted outside of Terraform
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Besu node, got error: %s", err))
		return
	}

	var nodeResp BesuNodeResponse
	if err := json.Unmarshal(body, &nodeResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse Besu node response, got error: %s", err))
		return
	}

	// Update state with response data
	data.Name = types.StringValue(nodeResp.Name)
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

func (r *BesuNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BesuNodeResourceModel
	var state BesuNodeResourceModel

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

	// Build update request
	updateReq := map[string]interface{}{
		"name": data.Name.ValueString(),
	}

	body, err := r.client.DoRequest("PUT", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update Besu node, got error: %s", err))
		return
	}

	var nodeResp BesuNodeResponse
	if err := json.Unmarshal(body, &nodeResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse Besu node response, got error: %s", err))
		return
	}

	if nodeResp.Status != "" {
		data.Status = types.StringValue(nodeResp.Status)
	}
	if nodeResp.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(nodeResp.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BesuNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BesuNodeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DoRequest("DELETE", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Besu node, got error: %s", err))
		return
	}
}

func (r *BesuNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// waitForNodeRunning polls the node status until it reaches RUNNING state or timeout
func (r *BesuNodeResource) waitForNodeRunning(ctx context.Context, nodeID int64) error {
	maxAttempts := 30 // 30 attempts
	delaySeconds := 2 // 2 seconds between attempts

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for node to be running")
		default:
		}

		// Call the node status endpoint
		body, err := r.client.DoRequest("GET", fmt.Sprintf("/nodes/%d", nodeID), nil)
		if err != nil {
			// If it's just not ready yet, continue waiting
			if attempt < maxAttempts {
				time.Sleep(time.Duration(delaySeconds) * time.Second)
				continue
			}
			return fmt.Errorf("failed to get node status after %d attempts: %s", maxAttempts, err)
		}

		var nodeResp BesuNodeResponse
		if err := json.Unmarshal(body, &nodeResp); err != nil {
			return fmt.Errorf("failed to parse node status response: %s", err)
		}

		// Check if node is running
		if nodeResp.Status == "RUNNING" {
			return nil // Node is ready!
		}

		// If node is in error state, return error
		if nodeResp.Status == "ERROR" || nodeResp.Status == "FAILED" {
			return fmt.Errorf("node entered error state: %s", nodeResp.Status)
		}

		// Not ready yet, wait and try again
		if attempt < maxAttempts {
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}
	}

	return fmt.Errorf("node did not reach RUNNING state after %d attempts (%d seconds)", maxAttempts, maxAttempts*delaySeconds)
}
