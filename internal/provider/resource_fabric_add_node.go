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

var _ resource.Resource = &FabricAddNodeResource{}
var _ resource.ResourceWithImportState = &FabricAddNodeResource{}

func NewFabricAddNodeResource() resource.Resource {
	return &FabricAddNodeResource{}
}

type FabricAddNodeResource struct {
	client *Client
}

type FabricAddNodeResourceModel struct {
	ID        types.String `tfsdk:"id"`
	NetworkID types.Int64  `tfsdk:"network_id"`
	NodeID    types.Int64  `tfsdk:"node_id"`
	Role      types.String `tfsdk:"role"`
}

func (r *FabricAddNodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_add_node"
}

func (r *FabricAddNodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Logically associates a Fabric node (peer or orderer) with a Fabric network/channel. This adds the node to the network configuration but does not perform the physical Fabric channel join operation. For peers, use chainlaunch_fabric_join_node after this to perform the actual channel join.",

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
				Description: "The ID of the Fabric network (channel) to add the node to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"node_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the node to add to the network.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "The role of the node in the network (must be 'peer' or 'orderer').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *FabricAddNodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FabricAddNodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricAddNodeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate role
	if data.Role.ValueString() != "peer" && data.Role.ValueString() != "orderer" {
		resp.Diagnostics.AddError("Invalid Role", fmt.Sprintf("Role must be 'peer' or 'orderer', got: %s", data.Role.ValueString()))
		return
	}

	addNodeReq := AddNodeToNetworkRequest{
		NodeID: data.NodeID.ValueInt64(),
		Role:   data.Role.ValueString(),
	}

	endpoint := fmt.Sprintf("/networks/fabric/%d/nodes", data.NetworkID.ValueInt64())

	body, err := r.client.DoRequest("POST", endpoint, addNodeReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to add node to network, got error: %s", err))
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

func (r *FabricAddNodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricAddNodeResourceModel

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
		if node.NodeID == data.NodeID.ValueInt64() && node.Role == data.Role.ValueString() {
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

func (r *FabricAddNodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update is not supported - all changes require replacement
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Updating a node's network association is not supported. Changes require recreating the resource.",
	)
}

func (r *FabricAddNodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FabricAddNodeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine the endpoint based on role
	var endpoint string
	if data.Role.ValueString() == "peer" {
		endpoint = fmt.Sprintf("/networks/fabric/%d/peers/%d", data.NetworkID.ValueInt64(), data.NodeID.ValueInt64())
	} else if data.Role.ValueString() == "orderer" {
		endpoint = fmt.Sprintf("/networks/fabric/%d/orderers/%d", data.NetworkID.ValueInt64(), data.NodeID.ValueInt64())
	} else {
		resp.Diagnostics.AddError("Invalid Role", fmt.Sprintf("Unknown role: %s", data.Role.ValueString()))
		return
	}

	_, err := r.client.DoRequest("DELETE", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to remove node from network, got error: %s", err))
		return
	}
}

func (r *FabricAddNodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: network_id:node_id:role
	parts := strings.Split(req.ID, ":")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format 'network_id:node_id:role', got: %s", req.ID),
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

	role := parts[2]
	if role != "peer" && role != "orderer" {
		resp.Diagnostics.AddError("Invalid Role", fmt.Sprintf("Role must be 'peer' or 'orderer', got: %s", role))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("network_id"), networkID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("node_id"), nodeID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("role"), role)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
