package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &FabricOrdererResource{}
var _ resource.ResourceWithImportState = &FabricOrdererResource{}

func NewFabricOrdererResource() resource.Resource {
	return &FabricOrdererResource{}
}

type FabricOrdererResource struct {
	client *Client
}

type FabricOrdererResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	OrganizationID          types.Int64  `tfsdk:"organization_id"`
	MspID                   types.String `tfsdk:"msp_id"`
	Mode                    types.String `tfsdk:"mode"`
	Version                 types.String `tfsdk:"version"`
	ListenAddress           types.String `tfsdk:"listen_address"`
	AdminAddress            types.String `tfsdk:"admin_address"`
	OperationsListenAddress types.String `tfsdk:"operations_listen_address"`
	ExternalEndpoint        types.String `tfsdk:"external_endpoint"`
	DomainNames             types.List   `tfsdk:"domain_names"`
	CertificateExpiration   types.Int64  `tfsdk:"certificate_expiration"`
	AutoRenewalEnabled      types.Bool   `tfsdk:"auto_renewal_enabled"`
	AutoRenewalDays         types.Int64  `tfsdk:"auto_renewal_days"`
	Environment             types.Map    `tfsdk:"environment"`
	Status                  types.String `tfsdk:"status"`
	CreatedAt               types.String `tfsdk:"created_at"`
	UpdatedAt               types.String `tfsdk:"updated_at"`
}

func (r *FabricOrdererResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_orderer"
}

func (r *FabricOrdererResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Hyperledger Fabric orderer node in Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the orderer node.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the orderer node (e.g., orderer0-org1).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the organization that owns this orderer.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"msp_id": schema.StringAttribute{
				Required:    true,
				Description: "The MSP ID for the organization (e.g., OrdererMSP).",
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
				Description: "Listen address for the orderer (e.g., 0.0.0.0:7050).",
			},
			"admin_address": schema.StringAttribute{
				Required:    true,
				Description: "Admin listen address for the orderer (e.g., 0.0.0.0:7053).",
			},
			"operations_listen_address": schema.StringAttribute{
				Required:    true,
				Description: "Operations listen address (e.g., 0.0.0.0:8443).",
			},
			"external_endpoint": schema.StringAttribute{
				Required:    true,
				Description: "External endpoint for the orderer (e.g., orderer0.org1.example.com:7050 or localhost:7050).",
			},
			"domain_names": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Domain names for the orderer.",
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
				Description: "Environment variables for the orderer container.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the orderer node.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the orderer was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the orderer was last updated.",
			},
		},
	}
}

func (r *FabricOrdererResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FabricOrdererResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricOrdererResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the FabricOrdererConfig
	ordererConfig := map[string]interface{}{
		"name":           data.Name.ValueString(),
		"organizationId": data.OrganizationID.ValueInt64(),
		"mspId":          data.MspID.ValueString(),
		"type":           "fabric-orderer",
	}

	// Add optional fields
	if !data.Mode.IsNull() {
		ordererConfig["mode"] = data.Mode.ValueString()
	}
	if !data.Version.IsNull() {
		ordererConfig["version"] = data.Version.ValueString()
	}
	if !data.ListenAddress.IsNull() {
		ordererConfig["listenAddress"] = data.ListenAddress.ValueString()
	}
	if !data.AdminAddress.IsNull() {
		ordererConfig["adminAddress"] = data.AdminAddress.ValueString()
	}
	if !data.OperationsListenAddress.IsNull() {
		ordererConfig["operationsListenAddress"] = data.OperationsListenAddress.ValueString()
	}
	if !data.ExternalEndpoint.IsNull() {
		ordererConfig["externalEndpoint"] = data.ExternalEndpoint.ValueString()
	}
	if !data.CertificateExpiration.IsNull() {
		ordererConfig["certificateExpiration"] = data.CertificateExpiration.ValueInt64()
	}
	if !data.AutoRenewalEnabled.IsNull() {
		ordererConfig["autoRenewalEnabled"] = data.AutoRenewalEnabled.ValueBool()
	}
	if !data.AutoRenewalDays.IsNull() {
		ordererConfig["autoRenewalDays"] = data.AutoRenewalDays.ValueInt64()
	}

	// Handle domain names list
	if !data.DomainNames.IsNull() {
		var domainNames []string
		diags := data.DomainNames.ElementsAs(ctx, &domainNames, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ordererConfig["domainNames"] = domainNames
	}

	// Handle environment variables map
	if !data.Environment.IsNull() {
		var envVars map[string]string
		diags := data.Environment.ElementsAs(ctx, &envVars, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ordererConfig["env"] = envVars
	}

	// Build the CreateNodeRequest
	createReq := map[string]interface{}{
		"name":               data.Name.ValueString(),
		"blockchainPlatform": "FABRIC",
		"fabricOrderer":      ordererConfig,
	}

	body, err := r.client.DoRequest("POST", "/nodes", createReq)
	var nodeResp OrdererNodeResponse

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
					"Orderer Creation Failed",
					fmt.Sprintf("Orderer creation failed at stage '%s': %s\nOrderer was not created in the database and will not be saved to state.",
						errResp.Details.Stage, errResp.Message),
				)
				return
			}
			// Node WAS created in database but deployment failed - fetch full node details from API
			resp.Diagnostics.AddWarning(
				"Orderer Partially Created",
				fmt.Sprintf("Orderer was created in database (ID: %d) but deployment failed at stage '%s': %s\nThe orderer will be saved to state so you can manage or delete it.",
					errResp.Details.NodeID, errResp.Details.Stage, errResp.Message),
			)

			// Fetch the full node details from the API
			nodeBody, getErr := r.client.DoRequest("GET", fmt.Sprintf("/nodes/%d", errResp.Details.NodeID), nil)
			if getErr != nil {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Orderer was created (ID: %d) but unable to fetch node details: %s", errResp.Details.NodeID, getErr))
				return
			}
			if unmarshalErr := json.Unmarshal(nodeBody, &nodeResp); unmarshalErr != nil {
				resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse orderer details: %s\nResponse body: %s", unmarshalErr, string(nodeBody)))
				return
			}
		} else {
			// Generic error without node creation details
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create fabric orderer, got error: %s", err))
			return
		}
	} else {
		// Success case - parse the response body as OrdererNodeResponse
		if err := json.Unmarshal(body, &nodeResp); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse orderer response, got error: %s\nResponse body: %s", err, string(body)))
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

	// Wait for orderer to reach RUNNING state
	if err := r.waitForOrdererRunning(ctx, nodeResp.ID); err != nil {
		resp.Diagnostics.AddWarning(
			"Orderer Status Check",
			fmt.Sprintf("Orderer created but did not reach RUNNING state: %s. The orderer may still be starting up.", err),
		)
	} else {
		// Refresh status after waiting
		data.Status = types.StringValue("RUNNING")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricOrdererResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricOrdererResourceModel

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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read fabric orderer, got error: %s", err))
		return
	}

	var nodeResp OrdererNodeResponse
	if err := json.Unmarshal(body, &nodeResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse orderer response, got error: %s", err))
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricOrdererResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FabricOrdererResourceModel
	var state FabricOrdererResourceModel

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

	// Build the FabricOrdererConfig for update
	ordererConfig := map[string]interface{}{
		"name":           data.Name.ValueString(),
		"organizationId": data.OrganizationID.ValueInt64(),
		"mspId":          data.MspID.ValueString(),
		"type":           "fabric-orderer",
	}

	// Add optional fields
	if !data.Mode.IsNull() {
		ordererConfig["mode"] = data.Mode.ValueString()
	}
	if !data.Version.IsNull() {
		ordererConfig["version"] = data.Version.ValueString()
	}
	if !data.ListenAddress.IsNull() {
		ordererConfig["listenAddress"] = data.ListenAddress.ValueString()
	}
	if !data.AdminAddress.IsNull() {
		ordererConfig["adminAddress"] = data.AdminAddress.ValueString()
	}
	if !data.OperationsListenAddress.IsNull() {
		ordererConfig["operationsListenAddress"] = data.OperationsListenAddress.ValueString()
	}
	if !data.ExternalEndpoint.IsNull() {
		ordererConfig["externalEndpoint"] = data.ExternalEndpoint.ValueString()
	}
	if !data.CertificateExpiration.IsNull() {
		ordererConfig["certificateExpiration"] = data.CertificateExpiration.ValueInt64()
	}
	if !data.AutoRenewalEnabled.IsNull() {
		ordererConfig["autoRenewalEnabled"] = data.AutoRenewalEnabled.ValueBool()
	}
	if !data.AutoRenewalDays.IsNull() {
		ordererConfig["autoRenewalDays"] = data.AutoRenewalDays.ValueInt64()
	}

	// Handle domain names list
	if !data.DomainNames.IsNull() {
		var domainNames []string
		diags := data.DomainNames.ElementsAs(ctx, &domainNames, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ordererConfig["domainNames"] = domainNames
	}

	// Handle environment variables map
	if !data.Environment.IsNull() {
		var envVars map[string]string
		diags := data.Environment.ElementsAs(ctx, &envVars, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		ordererConfig["env"] = envVars
	}

	// Build the update request
	updateReq := map[string]interface{}{
		"name":               data.Name.ValueString(),
		"blockchainPlatform": "FABRIC",
		"fabricOrderer":      ordererConfig,
	}

	body, err := r.client.DoRequest("PUT", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update fabric orderer, got error: %s", err))
		return
	}

	var nodeResp OrdererNodeResponse
	if err := json.Unmarshal(body, &nodeResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse orderer response, got error: %s", err))
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

func (r *FabricOrdererResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FabricOrdererResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DoRequest("DELETE", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete fabric orderer, got error: %s", err))
		return
	}
}

func (r *FabricOrdererResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: id or id,organization_id,msp_id
	// We'll support simple ID import and let the Read method populate the rest
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// OrdererNodeResponse represents the API response when creating/reading an orderer node
type OrdererNodeResponse struct {
	ID            int64                  `json:"id"`
	Name          string                 `json:"name"`
	Platform      string                 `json:"blockchainPlatform,omitempty"`
	Status        string                 `json:"status,omitempty"`
	CreatedAt     string                 `json:"createdAt,omitempty"`
	UpdatedAt     string                 `json:"updatedAt,omitempty"`
	Config        map[string]interface{} `json:"config,omitempty"`
	FabricOrderer map[string]interface{} `json:"fabricOrderer,omitempty"`
}

// waitForOrdererRunning polls the node status until it reaches RUNNING state or timeout
func (r *FabricOrdererResource) waitForOrdererRunning(ctx context.Context, ordererID int64) error {
	maxAttempts := 60 // 60 attempts
	delaySeconds := 2 // 2 seconds between attempts (total 120 seconds / 2 minutes)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for orderer to reach RUNNING state")
		default:
		}

		// Get current node status
		body, err := r.client.DoRequest("GET", fmt.Sprintf("/nodes/%d", ordererID), nil)
		if err != nil {
			return fmt.Errorf("failed to check orderer status: %w", err)
		}

		var nodeResp OrdererNodeResponse
		if err := json.Unmarshal(body, &nodeResp); err != nil {
			return fmt.Errorf("failed to parse orderer status response: %w", err)
		}

		// Check if orderer is running
		if nodeResp.Status == "RUNNING" {
			return nil
		}

		// Check if orderer is in ERROR state
		if nodeResp.Status == "ERROR" {
			return fmt.Errorf("orderer entered ERROR state")
		}

		// Not ready yet, wait and try again
		if attempt < maxAttempts {
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}
	}

	return fmt.Errorf("orderer did not reach RUNNING state after %d attempts (%d seconds)", maxAttempts, maxAttempts*delaySeconds)
}
