package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SharedNetworkAcceptResource{}

func NewSharedNetworkAcceptResource() resource.Resource {
	return &SharedNetworkAcceptResource{}
}

type SharedNetworkAcceptResource struct {
	client *Client
}

type SharedNetworkAcceptResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ShareID   types.String `tfsdk:"share_id"`
	NetworkID types.Int64  `tfsdk:"network_id"`
	Status    types.String `tfsdk:"status"`
}

func (r *SharedNetworkAcceptResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_shared_network_accept"
}

func (r *SharedNetworkAcceptResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Accept a network that has been shared with this Chainlaunch instance.

When another Chainlaunch instance shares a network (Fabric or Besu) with you, use this resource to accept it and import the network configuration locally.

**Important**: This is a Pro-only feature.

**Workflow**:
1. Connect to peer nodes using ` + "`chainlaunch_node_invitation`" + ` and ` + "`chainlaunch_node_accept_invitation`" + `
2. Peers share networks with you using ` + "`chainlaunch_network_share`" + `
3. Query shared networks using ` + "`data.chainlaunch_shared_networks`" + `
4. Accept the network using this resource

**What happens on accept**:
- The network genesis block and configuration are imported locally
- A new network record is created in your Chainlaunch instance
- You can then join your nodes to this network

**Rejection**:
To reject a shared network instead, use the API directly: ` + "`POST /pro/shared-networks/{shareId}/reject`" + ``,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for this acceptance (same as share_id)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"share_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the shared network to accept (from `data.chainlaunch_shared_networks`)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "ID of the locally created network after acceptance",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the acceptance (accepted)",
			},
		},
	}
}

func (r *SharedNetworkAcceptResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SharedNetworkAcceptResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SharedNetworkAcceptResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	shareID := data.ShareID.ValueString()

	// Call the accept endpoint
	endpoint := fmt.Sprintf("/pro/shared-networks/%s/accept", shareID)

	body, err := r.client.DoRequest("POST", endpoint, map[string]interface{}{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to accept shared network, got error: %s", err),
		)
		return
	}

	// Parse response
	var acceptResp struct {
		NetworkID int64  `json:"networkId"`
		Status    string `json:"status"`
	}

	if err := json.Unmarshal(body, &acceptResp); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse accept response: %s", err),
		)
		return
	}

	// Set state
	data.ID = types.StringValue(shareID)
	data.NetworkID = types.Int64Value(acceptResp.NetworkID)
	data.Status = types.StringValue(acceptResp.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SharedNetworkAcceptResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SharedNetworkAcceptResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Accepted shares are permanent - the network is now local
	// We just preserve the state since the acceptance is complete
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SharedNetworkAcceptResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// share_id requires replacement, so updates should not occur
	var data SharedNetworkAcceptResourceModel
	var state SharedNetworkAcceptResourceModel

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
	data.NetworkID = state.NetworkID
	data.Status = state.Status

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SharedNetworkAcceptResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Accepting a shared network imports it locally - we can't "un-accept" it
	// The network would need to be deleted separately if desired
	// Just remove from Terraform state

	// Note: If you want to reject future shares, use the reject API endpoint directly
}
