package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &FabricAnchorPeersResource{}

func NewFabricAnchorPeersResource() resource.Resource {
	return &FabricAnchorPeersResource{}
}

type FabricAnchorPeersResource struct {
	client *Client
}

type FabricAnchorPeersResourceModel struct {
	ID             types.String `tfsdk:"id"`
	NetworkID      types.Int64  `tfsdk:"network_id"`
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	AnchorPeerIDs  types.List   `tfsdk:"anchor_peer_ids"`
	TransactionID  types.String `tfsdk:"transaction_id"`
}

func (r *FabricAnchorPeersResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_anchor_peers"
}

func (r *FabricAnchorPeersResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sets the anchor peers for an organization in a Fabric network/channel. Anchor peers are used for cross-organization gossip communication. " +
			"Note: Deleting this resource only removes it from Terraform state - anchor peers cannot be truly destroyed, only updated. " +
			"To clear anchor peers, update the resource with an empty peer_ids list.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for this resource (format: network_id:organization_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the Fabric network (channel).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"organization_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the organization.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"anchor_peer_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.Int64Type,
				Description: "List of peer node IDs to set as anchor peers for this organization.",
			},
			"transaction_id": schema.StringAttribute{
				Computed:    true,
				Description: "The transaction ID of the anchor peer update.",
			},
		},
	}
}

func (r *FabricAnchorPeersResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// setAnchorPeersWithRetry performs the anchor peers API call with exponential backoff
// This is necessary because consenters may not be available immediately after network creation
func (r *FabricAnchorPeersResource) setAnchorPeersWithRetry(ctx context.Context, networkID int64, organizationID int64, anchorPeers []interface{}) (string, error) {
	// AnchorPeer type defined for documentation purposes
	// nolint:unused
	type AnchorPeer struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	setAnchorPeersReq := struct {
		OrganizationID int64       `json:"organizationId"`
		AnchorPeers    interface{} `json:"anchorPeers"`
	}{
		OrganizationID: organizationID,
		AnchorPeers:    anchorPeers,
	}

	endpoint := fmt.Sprintf("/networks/fabric/%d/anchor-peers", networkID)

	// Exponential backoff: 500ms, 1s, 2s, 4s, 5s (max), 5s, ...
	// Total attempts: up to 10 retries over ~30 seconds
	maxRetries := 10
	baseDelay := 500 * time.Millisecond
	maxDelay := 5 * time.Second

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		body, err := r.client.DoRequest("POST", endpoint, setAnchorPeersReq)
		if err == nil {
			// Success - parse and return transaction ID
			var setResp struct {
				TransactionID string `json:"transactionId"`
			}
			if err := json.Unmarshal(body, &setResp); err != nil {
				return "", fmt.Errorf("unable to parse response: %s", err)
			}
			if attempt > 0 {
				// Log that we succeeded after retries
				return setResp.TransactionID, nil
			}
			return setResp.TransactionID, nil
		}

		lastErr = err

		// Don't retry on last attempt
		if attempt == maxRetries {
			break
		}

		// Calculate exponential backoff with cap at maxDelay
		delay := baseDelay * (1 << attempt) // 500ms * 2^attempt
		if delay > maxDelay {
			delay = maxDelay
		}

		// Wait before next attempt
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return "", fmt.Errorf("failed after %d attempts (last error: %s)", maxRetries+1, lastErr)
}

func (r *FabricAnchorPeersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricAnchorPeersResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert anchor_peer_ids list to slice
	var anchorPeerIDs []int64
	resp.Diagnostics.Append(data.AnchorPeerIDs.ElementsAs(ctx, &anchorPeerIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the request - need to get peer host:port from the API
	// First, fetch each peer to get their endpoints
	type AnchorPeer struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	var anchorPeers []AnchorPeer
	for _, peerID := range anchorPeerIDs {
		// Get peer details
		peerBody, err := r.client.DoRequest("GET", fmt.Sprintf("/nodes/%d", peerID), nil)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read peer %d: %s", peerID, err))
			return
		}

		var peerResp struct {
			FabricPeer map[string]interface{} `json:"fabricPeer"`
		}
		if err := json.Unmarshal(peerBody, &peerResp); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse peer %d response: %s", peerID, err))
			return
		}

		// Extract external endpoint
		if externalEndpoint, ok := peerResp.FabricPeer["externalEndpoint"].(string); ok {
			// Parse host:port using strings.Split
			parts := strings.Split(externalEndpoint, ":")
			if len(parts) != 2 {
				resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Invalid endpoint format %s (expected host:port)", externalEndpoint))
				return
			}

			port, err := strconv.Atoi(parts[1])
			if err != nil {
				resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse port in endpoint %s: %s", externalEndpoint, err))
				return
			}

			anchorPeers = append(anchorPeers, AnchorPeer{
				Host: parts[0],
				Port: port,
			})
		} else {
			resp.Diagnostics.AddError("Missing Field", fmt.Sprintf("Peer %d does not have externalEndpoint", peerID))
			return
		}
	}

	// Convert to []interface{} for the retry function
	anchorPeersInterface := make([]interface{}, len(anchorPeers))
	for i, ap := range anchorPeers {
		anchorPeersInterface[i] = map[string]interface{}{
			"host": ap.Host,
			"port": ap.Port,
		}
	}

	// Use retry function with exponential backoff
	// This handles the case where consenters are not available immediately after network creation
	transactionID, err := r.setAnchorPeersWithRetry(ctx, data.NetworkID.ValueInt64(), data.OrganizationID.ValueInt64(), anchorPeersInterface)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set anchor peers: %s", err))
		return
	}

	// Set computed fields
	data.ID = types.StringValue(fmt.Sprintf("%d:%d", data.NetworkID.ValueInt64(), data.OrganizationID.ValueInt64()))
	data.TransactionID = types.StringValue(transactionID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricAnchorPeersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricAnchorPeersResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API doesn't provide a GET endpoint for anchor peers, so we just keep the state as-is
	// In a real implementation, you might query the network config block to verify
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricAnchorPeersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FabricAnchorPeersResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert anchor_peer_ids list to slice
	var anchorPeerIDs []int64
	resp.Diagnostics.Append(data.AnchorPeerIDs.ElementsAs(ctx, &anchorPeerIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the request - need to get peer host:port from the API
	type AnchorPeer struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	var anchorPeers []AnchorPeer
	for _, peerID := range anchorPeerIDs {
		// Get peer details
		peerBody, err := r.client.DoRequest("GET", fmt.Sprintf("/nodes/%d", peerID), nil)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read peer %d: %s", peerID, err))
			return
		}

		var peerResp struct {
			FabricPeer map[string]interface{} `json:"fabricPeer"`
		}
		if err := json.Unmarshal(peerBody, &peerResp); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse peer %d response: %s", peerID, err))
			return
		}

		// Extract external endpoint
		if externalEndpoint, ok := peerResp.FabricPeer["externalEndpoint"].(string); ok {
			// Parse host:port using strings.Split
			parts := strings.Split(externalEndpoint, ":")
			if len(parts) != 2 {
				resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Invalid endpoint format %s (expected host:port)", externalEndpoint))
				return
			}

			port, err := strconv.Atoi(parts[1])
			if err != nil {
				resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse port in endpoint %s: %s", externalEndpoint, err))
				return
			}

			anchorPeers = append(anchorPeers, AnchorPeer{
				Host: parts[0],
				Port: port,
			})
		} else {
			resp.Diagnostics.AddError("Missing Field", fmt.Sprintf("Peer %d does not have externalEndpoint", peerID))
			return
		}
	}

	// Convert to []interface{} for the retry function
	anchorPeersInterface := make([]interface{}, len(anchorPeers))
	for i, ap := range anchorPeers {
		anchorPeersInterface[i] = map[string]interface{}{
			"host": ap.Host,
			"port": ap.Port,
		}
	}

	// Use retry function with exponential backoff
	transactionID, err := r.setAnchorPeersWithRetry(ctx, data.NetworkID.ValueInt64(), data.OrganizationID.ValueInt64(), anchorPeersInterface)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to set anchor peers: %s", err))
		return
	}

	// Set computed fields
	data.TransactionID = types.StringValue(transactionID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricAnchorPeersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Anchor peers are a configuration setting in Fabric that cannot be truly "destroyed"
	// They can only be updated to a different list of peers
	// Deletion just removes from Terraform state without making any API calls
	// If you want to clear anchor peers, update the resource with an empty peer_ids list instead
}
