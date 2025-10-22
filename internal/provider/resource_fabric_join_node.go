package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &FabricJoinNodeResource{}
var _ resource.ResourceWithImportState = &FabricJoinNodeResource{}

func NewFabricJoinNodeResource() resource.Resource {
	return &FabricJoinNodeResource{}
}

type FabricJoinNodeResource struct {
	client *Client
}

type FabricJoinNodeResourceModel struct {
	ID        types.String `tfsdk:"id"`
	NetworkID types.Int64  `tfsdk:"network_id"`
	NodeID    types.Int64  `tfsdk:"node_id"`
	Role      types.String `tfsdk:"role"`
}

func (r *FabricJoinNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_join_node"
}

func (r *FabricJoinNodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Performs the physical join operation for a Fabric node (peer or orderer) to join a channel. The node must already be logically associated with the network through the network's peer_organizations or orderer_organizations configuration before using this resource.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for this resource (format: network_id:node_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"network_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the Fabric network (channel) to join the node to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"node_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the node to join to the network.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "The role of the node in the network (e.g., 'peer', 'orderer').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *FabricJoinNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FabricJoinNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricJoinNodeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the role-specific join endpoint
	var endpoint string
	if data.Role.ValueString() == "peer" {
		endpoint = fmt.Sprintf("/networks/fabric/%d/peers/%d/join", data.NetworkID.ValueInt64(), data.NodeID.ValueInt64())
	} else if data.Role.ValueString() == "orderer" {
		endpoint = fmt.Sprintf("/networks/fabric/%d/orderers/%d/join", data.NetworkID.ValueInt64(), data.NodeID.ValueInt64())
	} else {
		resp.Diagnostics.AddError("Invalid Role", fmt.Sprintf("Role must be 'peer' or 'orderer', got: %s", data.Role.ValueString()))
		return
	}

	body, err := r.client.DoRequest("POST", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to join node to channel, got error: %s", err))
		return
	}

	// Parse response to verify success
	var networkResp FabricNetworkResponse
	if err := json.Unmarshal(body, &networkResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse network response: %s", err))
		return
	}

	// Set the composite ID
	data.ID = types.StringValue(fmt.Sprintf("%d:%d", data.NetworkID.ValueInt64(), data.NodeID.ValueInt64()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricJoinNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricJoinNodeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the list of nodes in the network
	endpoint := fmt.Sprintf("/networks/fabric/%d/nodes", data.NetworkID.ValueInt64())

	body, err := r.client.DoRequest("GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read network nodes, got error: %s", err))
		return
	}

	var nodesResp GetNetworkNodesResponse
	if err := json.Unmarshal(body, &nodesResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse nodes response: %s", err))
		return
	}

	// Check if the node is still in the network
	found := false
	for _, node := range nodesResp.Nodes {
		if node.NodeID == data.NodeID.ValueInt64() {
			found = true
			break
		}
	}

	if !found {
		// Node is no longer in the network, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Node still exists, keep state as is
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricJoinNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is not supported - all changes require replacement
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Updating a node's network membership is not supported. Changes require recreating the resource.",
	)
}

func (r *FabricJoinNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FabricJoinNodeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the unjoin endpoint to remove node from channel
	var endpoint string
	if data.Role.ValueString() == "peer" {
		endpoint = fmt.Sprintf("/networks/fabric/%d/peers/%d/unjoin", data.NetworkID.ValueInt64(), data.NodeID.ValueInt64())
	} else if data.Role.ValueString() == "orderer" {
		endpoint = fmt.Sprintf("/networks/fabric/%d/orderers/%d/unjoin", data.NetworkID.ValueInt64(), data.NodeID.ValueInt64())
	} else {
		resp.Diagnostics.AddError("Invalid Role", fmt.Sprintf("Unknown role: %s", data.Role.ValueString()))
		return
	}

	_, err := r.client.DoRequest("POST", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to unjoin node from channel, got error: %s", err))
		return
	}
}

func (r *FabricJoinNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: network_id:node_id
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'network_id:node_id', got: %s", req.ID),
		)
		return
	}

	networkID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Network ID", fmt.Sprintf("Unable to parse network ID: %s", err))
		return
	}

	nodeID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Node ID", fmt.Sprintf("Unable to parse node ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("network_id"), networkID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("node_id"), nodeID)...)
	// Note: Role needs to be determined from the network configuration
	resp.Diagnostics.AddWarning(
		"Role Not Determined",
		"The 'role' field was not set during import. You may need to update your configuration to specify the role.",
	)
}
