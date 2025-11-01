package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &FabricPeerResource{}
var _ resource.ResourceWithImportState = &FabricPeerResource{}

func NewFabricPeerResource() resource.Resource {
	return &FabricPeerResource{}
}

type FabricPeerResource struct {
	client *Client
}

type FabricPeerResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	OrganizationID          types.Int64  `tfsdk:"organization_id"`
	MspID                   types.String `tfsdk:"msp_id"`
	Mode                    types.String `tfsdk:"mode"`
	Version                 types.String `tfsdk:"version"`
	ListenAddress           types.String `tfsdk:"listen_address"`
	ChaincodeAddress        types.String `tfsdk:"chaincode_address"`
	EventsAddress           types.String `tfsdk:"events_address"`
	OperationsListenAddress types.String `tfsdk:"operations_listen_address"`
	ExternalEndpoint        types.String `tfsdk:"external_endpoint"`
	AddressOverrides        types.List   `tfsdk:"address_overrides"`
	DomainNames             types.List   `tfsdk:"domain_names"`
	CertificateExpiration   types.Int64  `tfsdk:"certificate_expiration"`
	AutoRenewalEnabled      types.Bool   `tfsdk:"auto_renewal_enabled"`
	AutoRenewalDays         types.Int64  `tfsdk:"auto_renewal_days"`
	Environment             types.Map    `tfsdk:"environment"`
	Status                  types.String `tfsdk:"status"`
	CreatedAt               types.String `tfsdk:"created_at"`
	UpdatedAt               types.String `tfsdk:"updated_at"`
}

type AddressOverrideModel struct {
	From      types.String `tfsdk:"from"`
	To        types.String `tfsdk:"to"`
	TLSCACert types.String `tfsdk:"tls_ca_cert"`
}

func (r *FabricPeerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_peer"
}

func (r *FabricPeerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Hyperledger Fabric peer node in Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the peer node.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the peer node (e.g., peer0-org1).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the organization that owns this peer.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"msp_id": schema.StringAttribute{
				Required:    true,
				Description: "The MSP ID for the organization (e.g., Org1MSP).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mode": schema.StringAttribute{
				Required:    true,
				Description: "The deployment mode: 'docker' or 'service'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Required:    true,
				Description: "Fabric version to use (e.g., 2.2.0, 2.5.0, 2.5.9).",
			},
			"listen_address": schema.StringAttribute{
				Required:    true,
				Description: "Listen address for the peer (e.g., 0.0.0.0:7051).",
			},
			"chaincode_address": schema.StringAttribute{
				Required:    true,
				Description: "Chaincode listen address (e.g., 0.0.0.0:7052).",
			},
			"events_address": schema.StringAttribute{
				Required:    true,
				Description: "Events listen address (e.g., 0.0.0.0:7053).",
			},
			"operations_listen_address": schema.StringAttribute{
				Required:    true,
				Description: "Operations listen address (e.g., 0.0.0.0:9443).",
			},
			"external_endpoint": schema.StringAttribute{
				Required:    true,
				Description: "External endpoint for the peer (e.g., peer0.org1.example.com:7051 or localhost:7051).",
			},
			"address_overrides": schema.ListNestedAttribute{
				Optional:    true,
				Description: "Address overrides for the peer to map external addresses to internal addresses.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"from": schema.StringAttribute{
							Required:    true,
							Description: "The external address to override (e.g., peer0.org1.example.com:7051).",
						},
						"to": schema.StringAttribute{
							Required:    true,
							Description: "The internal address to map to (e.g., peer0-org1:7051).",
						},
						"tls_ca_cert": schema.StringAttribute{
							Optional:    true,
							Description: "Optional TLS CA certificate for the target address.",
						},
					},
				},
			},
			"domain_names": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Domain names for the peer.",
			},
			"certificate_expiration": schema.Int64Attribute{
				Optional:    true,
				Description: "Certificate expiration in days. Defaults to 365.",
			},
			"auto_renewal_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable automatic certificate renewal before expiration. Defaults to false.",
			},
			"auto_renewal_days": schema.Int64Attribute{
				Optional:    true,
				Description: "Days before expiration to trigger auto-renewal. Defaults to 30.",
			},
			"environment": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Environment variables for the peer container.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the peer node.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the peer was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the peer was last updated.",
			},
		},
	}
}

func (r *FabricPeerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FabricPeerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricPeerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the FabricPeerConfig
	peerConfig := map[string]interface{}{
		"name":           data.Name.ValueString(),
		"organizationId": data.OrganizationID.ValueInt64(),
		"mspId":          data.MspID.ValueString(),
		"type":           "fabric-peer",
	}

	// Add optional fields
	if !data.Mode.IsNull() {
		peerConfig["mode"] = data.Mode.ValueString()
	}
	if !data.Version.IsNull() {
		peerConfig["version"] = data.Version.ValueString()
	}
	if !data.ListenAddress.IsNull() {
		peerConfig["listenAddress"] = data.ListenAddress.ValueString()
	}
	if !data.ChaincodeAddress.IsNull() {
		peerConfig["chaincodeAddress"] = data.ChaincodeAddress.ValueString()
	}
	if !data.EventsAddress.IsNull() {
		peerConfig["eventsAddress"] = data.EventsAddress.ValueString()
	}
	if !data.OperationsListenAddress.IsNull() {
		peerConfig["operationsListenAddress"] = data.OperationsListenAddress.ValueString()
	}
	if !data.ExternalEndpoint.IsNull() {
		peerConfig["externalEndpoint"] = data.ExternalEndpoint.ValueString()
	}
	if !data.CertificateExpiration.IsNull() {
		peerConfig["certificateExpiration"] = data.CertificateExpiration.ValueInt64()
	}
	if !data.AutoRenewalEnabled.IsNull() {
		peerConfig["autoRenewalEnabled"] = data.AutoRenewalEnabled.ValueBool()
	}
	if !data.AutoRenewalDays.IsNull() {
		peerConfig["autoRenewalDays"] = data.AutoRenewalDays.ValueInt64()
	}

	// Handle address overrides list
	if !data.AddressOverrides.IsNull() {
		var addressOverrides []AddressOverrideModel
		diags := data.AddressOverrides.ElementsAs(ctx, &addressOverrides, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Convert to API format
		overrides := make([]map[string]interface{}, len(addressOverrides))
		for i, override := range addressOverrides {
			overrides[i] = map[string]interface{}{
				"from": override.From.ValueString(),
				"to":   override.To.ValueString(),
			}
			if !override.TLSCACert.IsNull() {
				overrides[i]["tlsCACert"] = override.TLSCACert.ValueString()
			}
		}
		peerConfig["addressOverrides"] = overrides
	}

	// Handle domain names list
	if !data.DomainNames.IsNull() {
		var domainNames []string
		diags := data.DomainNames.ElementsAs(ctx, &domainNames, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		peerConfig["domainNames"] = domainNames
	}

	// Handle environment variables map
	if !data.Environment.IsNull() {
		var envVars map[string]string
		diags := data.Environment.ElementsAs(ctx, &envVars, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		peerConfig["env"] = envVars
	}

	// Build the CreateNodeRequest
	createReq := map[string]interface{}{
		"name":               data.Name.ValueString(),
		"blockchainPlatform": "FABRIC",
		"fabricPeer":         peerConfig,
	}

	body, err := r.client.DoRequest("POST", "/nodes", createReq)
	var nodeResp NodeResponse

	if err != nil {
		// Try to parse as NodeCreationErrorResponse to check if node was created in DB
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
			Details struct {
				NodeCreated bool        `json:"node_created"`
				NodeID      int64       `json:"node_id"`
				Stage       string      `json:"stage"`
				Node        interface{} `json:"node"` // Contains the actual node data
			} `json:"details"`
		}

		// If we can parse the error response and node was NOT created, don't save to state
		if parseErr := json.Unmarshal(body, &errResp); parseErr == nil && errResp.Details.Stage != "" {
			if !errResp.Details.NodeCreated {
				// Node was NOT created in database - don't save to state
				resp.Diagnostics.AddError(
					"Peer Creation Failed",
					fmt.Sprintf("Peer creation failed at stage '%s': %s\nPeer was not created in the database and will not be saved to state.",
						errResp.Details.Stage, errResp.Message),
				)
				return
			}
			// Node WAS created in database but deployment failed - fetch full node details from API
			resp.Diagnostics.AddWarning(
				"Peer Partially Created",
				fmt.Sprintf("Peer was created in database (ID: %d) but deployment failed at stage '%s': %s\nThe peer will be saved to state so you can manage or delete it.",
					errResp.Details.NodeID, errResp.Details.Stage, errResp.Message),
			)

			// Fetch the full node details from the API
			nodeBody, getErr := r.client.DoRequest("GET", fmt.Sprintf("/nodes/%d", errResp.Details.NodeID), nil)
			if getErr != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Peer was created (ID: %d) but unable to fetch node details: %s", errResp.Details.NodeID, getErr))
				return
			}
			if unmarshalErr := json.Unmarshal(nodeBody, &nodeResp); unmarshalErr != nil {
				resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse peer details: %s\nResponse body: %s", unmarshalErr, string(nodeBody)))
				return
			}
		} else {
			// Generic error without node creation details
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create fabric peer, got error: %s", err))
			return
		}
	} else {
		// Success case - parse the response body as NodeResponse
		if err := json.Unmarshal(body, &nodeResp); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse peer response, got error: %s\nResponse body: %s", err, string(body)))
			return
		}
	}

	// Set state from response
	data.ID = types.StringValue(fmt.Sprintf("%d", nodeResp.ID))
	data.Name = types.StringValue(data.Name.ValueString()) // Preserve from plan
	if nodeResp.Status != "" {
		data.Status = types.StringValue(nodeResp.Status)
	}
	if nodeResp.CreatedAt != "" {
		data.CreatedAt = types.StringValue(nodeResp.CreatedAt)
	}
	if nodeResp.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(nodeResp.UpdatedAt)
	}

	// Wait for peer to reach RUNNING state
	if err := r.waitForPeerRunning(ctx, nodeResp.ID); err != nil {
		resp.Diagnostics.AddWarning(
			"Peer Status Check",
			fmt.Sprintf("Peer created but did not reach RUNNING state: %s. The peer may still be starting up.", err),
		)
	} else {
		// Refresh status after waiting
		data.Status = types.StringValue("RUNNING")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricPeerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricPeerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.DoRequest("GET", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), nil)
	if err != nil {
		// If resource not found (404), remove from state
		if IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read fabric peer, got error: %s", err))
		return
	}

	var nodeResp NodeResponse
	if err := json.Unmarshal(body, &nodeResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse peer response, got error: %s", err))
		return
	}

	// Update state from API response
	// Note: Some fields might not be returned by API, preserve them from state
	if nodeResp.Status != "" {
		data.Status = types.StringValue(nodeResp.Status)
	}
	if nodeResp.CreatedAt != "" {
		data.CreatedAt = types.StringValue(nodeResp.CreatedAt)
	}
	if nodeResp.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(nodeResp.UpdatedAt)
	}

	// Try to parse address_overrides from the API response if available
	if nodeResp.FabricPeer != nil {
		if overrides, ok := nodeResp.FabricPeer["addressOverrides"].([]interface{}); ok && len(overrides) > 0 {
			// Parse address overrides from API response
			addressOverrides := make([]AddressOverrideModel, len(overrides))
			for i, override := range overrides {
				if overrideMap, ok := override.(map[string]interface{}); ok {
					if from, ok := overrideMap["from"].(string); ok {
						addressOverrides[i].From = types.StringValue(from)
					}
					if to, ok := overrideMap["to"].(string); ok {
						addressOverrides[i].To = types.StringValue(to)
					}
					if tlsCACert, ok := overrideMap["tlsCACert"].(string); ok && tlsCACert != "" {
						addressOverrides[i].TLSCACert = types.StringValue(tlsCACert)
					} else {
						addressOverrides[i].TLSCACert = types.StringNull()
					}
				}
			}
			// Convert to types.List
			addressOverridesList, diags := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"from":        types.StringType,
					"to":          types.StringType,
					"tls_ca_cert": types.StringType,
				},
			}, addressOverrides)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				data.AddressOverrides = addressOverridesList
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricPeerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FabricPeerResourceModel
	var state FabricPeerResourceModel

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

	// Build the FabricPeerConfig for update
	peerConfig := map[string]interface{}{
		"name":           data.Name.ValueString(),
		"organizationId": data.OrganizationID.ValueInt64(),
		"mspId":          data.MspID.ValueString(),
		"type":           "fabric-peer",
	}

	// Add optional fields
	if !data.Mode.IsNull() {
		peerConfig["mode"] = data.Mode.ValueString()
	}
	if !data.Version.IsNull() {
		peerConfig["version"] = data.Version.ValueString()
	}
	if !data.ListenAddress.IsNull() {
		peerConfig["listenAddress"] = data.ListenAddress.ValueString()
	}
	if !data.ChaincodeAddress.IsNull() {
		peerConfig["chaincodeAddress"] = data.ChaincodeAddress.ValueString()
	}
	if !data.EventsAddress.IsNull() {
		peerConfig["eventsAddress"] = data.EventsAddress.ValueString()
	}
	if !data.OperationsListenAddress.IsNull() {
		peerConfig["operationsListenAddress"] = data.OperationsListenAddress.ValueString()
	}
	if !data.ExternalEndpoint.IsNull() {
		peerConfig["externalEndpoint"] = data.ExternalEndpoint.ValueString()
	}
	if !data.CertificateExpiration.IsNull() {
		peerConfig["certificateExpiration"] = data.CertificateExpiration.ValueInt64()
	}
	if !data.AutoRenewalEnabled.IsNull() {
		peerConfig["autoRenewalEnabled"] = data.AutoRenewalEnabled.ValueBool()
	}
	if !data.AutoRenewalDays.IsNull() {
		peerConfig["autoRenewalDays"] = data.AutoRenewalDays.ValueInt64()
	}

	// Handle address overrides list
	if !data.AddressOverrides.IsNull() {
		var addressOverrides []AddressOverrideModel
		diags := data.AddressOverrides.ElementsAs(ctx, &addressOverrides, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Convert to API format
		overrides := make([]map[string]interface{}, len(addressOverrides))
		for i, override := range addressOverrides {
			overrides[i] = map[string]interface{}{
				"from": override.From.ValueString(),
				"to":   override.To.ValueString(),
			}
			if !override.TLSCACert.IsNull() {
				overrides[i]["tlsCACert"] = override.TLSCACert.ValueString()
			}
		}
		peerConfig["addressOverrides"] = overrides
	}

	// Handle domain names list
	if !data.DomainNames.IsNull() {
		var domainNames []string
		diags := data.DomainNames.ElementsAs(ctx, &domainNames, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		peerConfig["domainNames"] = domainNames
	}

	// Handle environment variables map
	if !data.Environment.IsNull() {
		var envVars map[string]string
		diags := data.Environment.ElementsAs(ctx, &envVars, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		peerConfig["env"] = envVars
	}

	// Build the update request
	updateReq := map[string]interface{}{
		"name":               data.Name.ValueString(),
		"blockchainPlatform": "FABRIC",
		"fabricPeer":         peerConfig,
	}

	body, err := r.client.DoRequest("PUT", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update fabric peer, got error: %s", err))
		return
	}

	var nodeResp NodeResponse
	if err := json.Unmarshal(body, &nodeResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse peer response, got error: %s", err))
		return
	}

	// Update state from response
	if nodeResp.Status != "" {
		data.Status = types.StringValue(nodeResp.Status)
	}
	if nodeResp.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(nodeResp.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricPeerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FabricPeerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DoRequest("DELETE", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete fabric peer, got error: %s", err))
		return
	}
}

func (r *FabricPeerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: id or id,organization_id,msp_id
	// We'll support simple ID import and let the Read method populate the rest
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// NodeResponse represents the API response when creating/reading a node
type NodeResponse struct {
	ID         int64                  `json:"id"`
	Name       string                 `json:"name"`
	Platform   string                 `json:"blockchainPlatform,omitempty"`
	Status     string                 `json:"status,omitempty"`
	CreatedAt  string                 `json:"createdAt,omitempty"`
	UpdatedAt  string                 `json:"updatedAt,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
	FabricPeer map[string]interface{} `json:"fabricPeer,omitempty"`
}

// waitForPeerRunning polls the node status until it reaches RUNNING state or timeout
func (r *FabricPeerResource) waitForPeerRunning(ctx context.Context, peerID int64) error {
	maxAttempts := 60 // 60 attempts
	delaySeconds := 2 // 2 seconds between attempts (total 120 seconds / 2 minutes)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for peer to reach RUNNING state")
		default:
		}

		// Get current node status
		body, err := r.client.DoRequest("GET", fmt.Sprintf("/nodes/%d", peerID), nil)
		if err != nil {
			return fmt.Errorf("failed to check peer status: %w", err)
		}

		var nodeResp NodeResponse
		if err := json.Unmarshal(body, &nodeResp); err != nil {
			return fmt.Errorf("failed to parse peer status response: %w", err)
		}

		// Check if peer is running
		if nodeResp.Status == "RUNNING" {
			return nil
		}

		// Check if peer is in ERROR state
		if nodeResp.Status == "ERROR" {
			return fmt.Errorf("peer entered ERROR state")
		}

		// Not ready yet, wait and try again
		if attempt < maxAttempts {
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}
	}

	return fmt.Errorf("peer did not reach RUNNING state after %d attempts (%d seconds)", maxAttempts, maxAttempts*delaySeconds)
}
