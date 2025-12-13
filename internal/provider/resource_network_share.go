package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &NetworkShareResource{}

func NewNetworkShareResource() resource.Resource {
	return &NetworkShareResource{}
}

type NetworkShareResource struct {
	client *Client
}

type NetworkShareResourceModel struct {
	ID           types.String `tfsdk:"id"`
	NetworkID    types.String `tfsdk:"network_id"`
	NetworkType  types.String `tfsdk:"network_type"`
	Recipients   types.List   `tfsdk:"recipients"`
	Metadata     types.Map    `tfsdk:"metadata"`
	Status       types.String `tfsdk:"status"`
	SharedBy     types.String `tfsdk:"shared_by"`
	SharedByNode types.String `tfsdk:"shared_by_node"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

func (r *NetworkShareResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_share"
}

func (r *NetworkShareResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Share a network (Fabric or Besu) with other Chainlaunch peer nodes.

This resource allows you to share network configuration (genesis block, etc.) with specific peer nodes that have been connected via node invitations.

**Important**: This is a Pro-only feature. Both the sharing and receiving nodes must have Pro features enabled.

**Workflow**:
1. Connect to peer nodes using ` + "`chainlaunch_node_invitation`" + ` and ` + "`chainlaunch_node_accept_invitation`" + `
2. Create the network locally (` + "`chainlaunch_fabric_network`" + ` or ` + "`chainlaunch_besu_network`" + `)
3. Use this resource to share the network with connected peers
4. Peers can accept the shared network on their end

**Network Types**:
- ` + "`fabric`" + ` - Share a Hyperledger Fabric network/channel
- ` + "`besu`" + ` - Share a Hyperledger Besu network`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for this share",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the network to share (Fabric or Besu network ID)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of network to share: `fabric` or `besu`",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"recipients": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "List of peer node IDs (connection IDs) to share the network with. These must be connected peers from accepted invitations.",
			},
			"metadata": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Optional metadata to include with the share (key-value pairs)",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the share (e.g., 'pending', 'accepted', 'rejected')",
			},
			"shared_by": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Description of who shared this network",
			},
			"shared_by_node": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Node ID of the sharer",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the share was created",
			},
		},
	}
}

func (r *NetworkShareResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *NetworkShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NetworkShareResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate network type
	networkType := data.NetworkType.ValueString()
	if networkType != "fabric" && networkType != "besu" {
		resp.Diagnostics.AddError(
			"Invalid Network Type",
			fmt.Sprintf("network_type must be either 'fabric' or 'besu', got: %s", networkType),
		)
		return
	}

	// Parse network ID
	networkID, err := strconv.ParseInt(data.NetworkID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Network ID",
			fmt.Sprintf("Unable to parse network_id: %s", err),
		)
		return
	}

	// Parse recipients
	var recipients []string
	resp.Diagnostics.Append(data.Recipients.ElementsAs(ctx, &recipients, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert recipient IDs to integers
	var recipientIDs []int64
	for _, r := range recipients {
		recipientID, err := strconv.ParseInt(r, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Recipient ID",
				fmt.Sprintf("Unable to parse recipient ID '%s': %s", r, err),
			)
			return
		}
		recipientIDs = append(recipientIDs, recipientID)
	}

	if len(recipientIDs) == 0 {
		resp.Diagnostics.AddError(
			"No Recipients",
			"At least one recipient peer node ID must be specified",
		)
		return
	}

	// Build request payload
	payload := map[string]interface{}{
		"networkId":  networkID,
		"recipients": recipientIDs,
	}

	// Add metadata if provided
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		var metadata map[string]string
		resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadata, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload["metadata"] = metadata
	}

	// Determine endpoint based on network type
	var endpoint string
	if networkType == "fabric" {
		endpoint = "/pro/sharing/network"
	} else {
		endpoint = "/pro/sharing/besu-network"
	}

	// Make API call
	body, err := r.client.DoRequest("POST", endpoint, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to share network, got error: %s", err),
		)
		return
	}

	// Parse response
	var shareResp NetworkShareResponse
	if err := json.Unmarshal(body, &shareResp); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse share response: %s", err),
		)
		return
	}

	// Set state
	data.ID = types.StringValue(fmt.Sprintf("%d", shareResp.NetworkID))
	data.Status = types.StringValue("pending")
	data.SharedBy = types.StringValue("")
	data.SharedByNode = types.StringValue("")
	data.CreatedAt = types.StringValue("")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NetworkShareResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Network shares are ephemeral - once created, they're sent to recipients
	// We can't query individual share status, so we just keep the existing state
	// The receiving end can accept/reject via their own resources

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NetworkShareResourceModel
	var state NetworkShareResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve computed fields from state
	data.ID = state.ID
	data.Status = state.Status
	data.SharedBy = state.SharedBy
	data.SharedByNode = state.SharedByNode
	data.CreatedAt = state.CreatedAt

	// Update recipients by re-sharing
	// Parse network ID
	networkID, err := strconv.ParseInt(data.NetworkID.ValueString(), 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Network ID",
			fmt.Sprintf("Unable to parse network_id: %s", err),
		)
		return
	}

	// Parse recipients
	var recipients []string
	resp.Diagnostics.Append(data.Recipients.ElementsAs(ctx, &recipients, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert recipient IDs to integers
	var recipientIDs []int64
	for _, r := range recipients {
		recipientID, err := strconv.ParseInt(r, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Recipient ID",
				fmt.Sprintf("Unable to parse recipient ID '%s': %s", r, err),
			)
			return
		}
		recipientIDs = append(recipientIDs, recipientID)
	}

	// Build request payload
	payload := map[string]interface{}{
		"networkId":  networkID,
		"recipients": recipientIDs,
	}

	// Add metadata if provided
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		var metadata map[string]string
		resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadata, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload["metadata"] = metadata
	}

	// Determine endpoint based on network type
	networkType := data.NetworkType.ValueString()
	var endpoint string
	if networkType == "fabric" {
		endpoint = "/pro/sharing/network"
	} else {
		endpoint = "/pro/sharing/besu-network"
	}

	// Make API call
	_, err = r.client.DoRequest("POST", endpoint, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update network share, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Network shares can't be "deleted" - once sent, they're on the recipient's end
	// The recipient can reject them if needed
	// Just remove from Terraform state
}
