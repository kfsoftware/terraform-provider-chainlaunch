package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &FabricChaincodeResource{}

func NewFabricChaincodeResource() resource.Resource {
	return &FabricChaincodeResource{}
}

type FabricChaincodeResource struct {
	client *Client
}

type FabricChaincodeResourceModel struct {
	ID              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	NetworkID       types.Int64  `tfsdk:"network_id"`
	NetworkName     types.String `tfsdk:"network_name"`
	NetworkPlatform types.String `tfsdk:"network_platform"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

func (r *FabricChaincodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_chaincode"
}

func (r *FabricChaincodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fabric chaincode (smart contract) package. This resource creates a chaincode record associated with a network/channel. After creating the chaincode, you can create definitions for it with different versions and sequences.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier for the chaincode.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the chaincode. This is the chaincode name used in install/approve/commit operations.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"network_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the Fabric network (channel) this chaincode belongs to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"network_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the network (channel).",
			},
			"network_platform": schema.StringAttribute{
				Computed:    true,
				Description: "The network platform (e.g., 'FABRIC').",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the chaincode was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *FabricChaincodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FabricChaincodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricChaincodeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create chaincode request
	createReq := struct {
		Name      string `json:"name"`
		NetworkID int64  `json:"network_id"`
	}{
		Name:      data.Name.ValueString(),
		NetworkID: data.NetworkID.ValueInt64(),
	}

	body, err := r.client.DoRequest("POST", "/sc/fabric/chaincodes", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create chaincode: %s", err))
		return
	}

	var createResp struct {
		Chaincode struct {
			ID              int64  `json:"id"`
			Name            string `json:"name"`
			NetworkID       int64  `json:"network_id"`
			NetworkName     string `json:"network_name"`
			NetworkPlatform string `json:"network_platform"`
			CreatedAt       string `json:"created_at"`
		} `json:"chaincode"`
	}
	if err := json.Unmarshal(body, &createResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Set state
	data.ID = types.Int64Value(createResp.Chaincode.ID)
	data.Name = types.StringValue(createResp.Chaincode.Name)
	data.NetworkID = types.Int64Value(createResp.Chaincode.NetworkID)
	data.NetworkName = types.StringValue(createResp.Chaincode.NetworkName)
	data.NetworkPlatform = types.StringValue(createResp.Chaincode.NetworkPlatform)
	data.CreatedAt = types.StringValue(createResp.Chaincode.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricChaincodeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API doesn't have a GET /sc/fabric/chaincodes/{id} endpoint
	// We would need to list all chaincodes and find ours
	// For now, we'll keep the state as-is (optimistic approach)
	// In a production implementation, you might want to call LIST and verify existence

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FabricChaincodeResourceModel
	var state FabricChaincodeResourceModel

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

	// Preserve computed fields from state
	data.CreatedAt = state.CreatedAt

	// Note: The API doesn't have an UPDATE endpoint for chaincodes
	// Name and NetworkID are marked as RequiresReplace, so updates will trigger recreation
	// This Update method is here for consistency, but shouldn't be called

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FabricChaincodeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the chaincode
	endpoint := fmt.Sprintf("/sc/fabric/chaincodes/%d", data.ID.ValueInt64())
	_, err := r.client.DoRequest("DELETE", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete chaincode: %s", err))
		return
	}
}
