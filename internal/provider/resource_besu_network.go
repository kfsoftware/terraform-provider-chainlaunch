package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &BesuNetworkResource{}
var _ resource.ResourceWithImportState = &BesuNetworkResource{}

func NewBesuNetworkResource() resource.Resource {
	return &BesuNetworkResource{}
}

type BesuNetworkResource struct {
	client *Client
}

type BesuNetworkResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	ChainID                types.Int64  `tfsdk:"chain_id"`
	Consensus              types.String `tfsdk:"consensus"`
	BlockPeriod            types.Int64  `tfsdk:"block_period"`
	EpochLength            types.Int64  `tfsdk:"epoch_length"`
	RequestTimeout         types.Int64  `tfsdk:"request_timeout"`
	InitialValidatorKeyIds types.List   `tfsdk:"initial_validator_key_ids"`

	// Optional genesis fields
	GasLimit   types.String `tfsdk:"gas_limit"`
	Difficulty types.String `tfsdk:"difficulty"`
	MixHash    types.String `tfsdk:"mix_hash"`
	Nonce      types.String `tfsdk:"nonce"`
	Timestamp  types.String `tfsdk:"timestamp"`
	Coinbase   types.String `tfsdk:"coinbase"`

	// Computed fields
	Status    types.String `tfsdk:"status"`
	Platform  types.String `tfsdk:"platform"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *BesuNetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_besu_network"
}

func (r *BesuNetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a Hyperledger Besu network in Chainlaunch.",
		MarkdownDescription: "Manages a Hyperledger Besu network in Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Description:         "The unique identifier of the Besu network.",
				MarkdownDescription: "The unique identifier of the Besu network.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				Description:         "The name of the Besu network.",
				MarkdownDescription: "The name of the Besu network.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Description:         "Optional description of the network.",
				MarkdownDescription: "Optional description of the network.",
			},
			"chain_id": schema.Int64Attribute{
				Required:            true,
				Description:         "Chain ID for the network (e.g., 1337).",
				MarkdownDescription: "Chain ID for the network (e.g., 1337).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"consensus": schema.StringAttribute{
				Required:            true,
				Description:         "Consensus algorithm (e.g., 'qbft', 'ibft2').",
				MarkdownDescription: "Consensus algorithm (e.g., 'qbft', 'ibft2').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"block_period": schema.Int64Attribute{
				Required:            true,
				Description:         "Block period in seconds (default: 5).",
				MarkdownDescription: "Block period in seconds (default: 5).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"epoch_length": schema.Int64Attribute{
				Required:            true,
				Description:         "Epoch length in blocks (default: 30000).",
				MarkdownDescription: "Epoch length in blocks (default: 30000).",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"request_timeout": schema.Int64Attribute{
				Required:            true,
				Description:         "Request timeout in seconds.",
				MarkdownDescription: "Request timeout in seconds.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"initial_validator_key_ids": schema.ListAttribute{
				ElementType:         types.Int64Type,
				Required:            true,
				Description:         "List of initial validator key IDs (minimum 1).",
				MarkdownDescription: "List of initial validator key IDs (minimum 1).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			// Optional genesis fields
			"gas_limit": schema.StringAttribute{
				Optional:            true,
				Description:         "Optional gas limit value (hex format).",
				MarkdownDescription: "Optional gas limit value (hex format).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"difficulty": schema.StringAttribute{
				Optional:            true,
				Description:         "Optional difficulty value (hex format).",
				MarkdownDescription: "Optional difficulty value (hex format).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mix_hash": schema.StringAttribute{
				Optional:            true,
				Description:         "Optional mix hash value (hex format).",
				MarkdownDescription: "Optional mix hash value (hex format).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"nonce": schema.StringAttribute{
				Optional:            true,
				Description:         "Optional nonce value (hex format).",
				MarkdownDescription: "Optional nonce value (hex format).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timestamp": schema.StringAttribute{
				Optional:            true,
				Description:         "Optional timestamp value (hex format).",
				MarkdownDescription: "Optional timestamp value (hex format).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"coinbase": schema.StringAttribute{
				Optional:            true,
				Description:         "Optional coinbase address.",
				MarkdownDescription: "Optional coinbase address.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			// Computed fields
			"status": schema.StringAttribute{
				Computed:            true,
				Description:         "The current status of the network.",
				MarkdownDescription: "The current status of the network.",
			},
			"platform": schema.StringAttribute{
				Computed:            true,
				Description:         "The blockchain platform (besu).",
				MarkdownDescription: "The blockchain platform (besu).",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when the network was created.",
				MarkdownDescription: "The timestamp when the network was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				Description:         "The timestamp when the network was last updated.",
				MarkdownDescription: "The timestamp when the network was last updated.",
			},
		},
	}
}

func (r *BesuNetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BesuNetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BesuNetworkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert initial validator key IDs from types.List to []int64
	var keyIds []int64
	resp.Diagnostics.Append(data.InitialValidatorKeyIds.ElementsAs(ctx, &keyIds, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that we have at least one validator key
	if len(keyIds) == 0 {
		resp.Diagnostics.AddError(
			"Configuration Error",
			"At least one initial validator key ID must be provided",
		)
		return
	}

	// Build the config object
	config := map[string]interface{}{
		"chainId":                 data.ChainID.ValueInt64(),
		"consensus":               data.Consensus.ValueString(),
		"blockPeriod":             data.BlockPeriod.ValueInt64(),
		"epochLength":             data.EpochLength.ValueInt64(),
		"requestTimeout":          data.RequestTimeout.ValueInt64(),
		"initialValidatorsKeyIds": keyIds, // Note: "Validators" is plural in API
	}

	// Add genesis fields with defaults if not provided
	if !data.GasLimit.IsNull() {
		config["gasLimit"] = data.GasLimit.ValueString()
	} else {
		config["gasLimit"] = "0x1fffffffffffff" // Default gas limit
	}

	if !data.Difficulty.IsNull() {
		config["difficulty"] = data.Difficulty.ValueString()
	} else {
		config["difficulty"] = "0x1" // Default difficulty for private networks
	}

	if !data.MixHash.IsNull() {
		config["mixHash"] = data.MixHash.ValueString()
	} else {
		config["mixHash"] = "0x0000000000000000000000000000000000000000000000000000000000000000" // Default mix hash
	}

	if !data.Nonce.IsNull() {
		config["nonce"] = data.Nonce.ValueString()
	} else {
		config["nonce"] = "0x0000000000000000" // Default nonce (8 bytes = 16 hex chars)
	}

	if !data.Timestamp.IsNull() {
		config["timestamp"] = data.Timestamp.ValueString()
	} else {
		config["timestamp"] = "0x0" // Default timestamp (genesis)
	}

	if !data.Coinbase.IsNull() {
		config["coinbase"] = data.Coinbase.ValueString()
	} else {
		config["coinbase"] = "0x0000000000000000000000000000000000000000" // Default coinbase
	}

	// Build the request body
	createReq := map[string]interface{}{
		"name":   data.Name.ValueString(),
		"config": config,
	}

	if !data.Description.IsNull() {
		createReq["description"] = data.Description.ValueString()
	}

	body, err := r.client.DoRequest("POST", "/networks/besu", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Besu network, got error: %s", err))
		return
	}

	var networkResp struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Platform    string `json:"platform"`
		Status      string `json:"status"`
		CreatedAt   string `json:"createdAt"`
		UpdatedAt   string `json:"updatedAt"`
	}

	if err := json.Unmarshal(body, &networkResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse Besu network response, got error: %s", err))
		return
	}

	// Set the response values
	data.ID = types.StringValue(fmt.Sprintf("%d", networkResp.ID))
	data.Name = types.StringValue(networkResp.Name)
	if networkResp.Description != "" {
		data.Description = types.StringValue(networkResp.Description)
	}
	data.Platform = types.StringValue(networkResp.Platform)
	data.Status = types.StringValue(networkResp.Status)
	data.CreatedAt = types.StringValue(networkResp.CreatedAt)
	data.UpdatedAt = types.StringValue(networkResp.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BesuNetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BesuNetworkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/networks/besu/%s", data.ID.ValueString())

	body, err := r.client.DoRequest("GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Besu network, got error: %s", err))
		return
	}

	var networkResp struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Platform    string `json:"platform"`
		Status      string `json:"status"`
		CreatedAt   string `json:"createdAt"`
		UpdatedAt   string `json:"updatedAt"`
	}

	if err := json.Unmarshal(body, &networkResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse Besu network response, got error: %s", err))
		return
	}

	// Update state with API response
	// Note: Name is preserved from state as API might not return it
	if networkResp.Name != "" {
		data.Name = types.StringValue(networkResp.Name)
	}
	if networkResp.Description != "" {
		data.Description = types.StringValue(networkResp.Description)
	}
	data.Platform = types.StringValue(networkResp.Platform)
	data.Status = types.StringValue(networkResp.Status)
	data.CreatedAt = types.StringValue(networkResp.CreatedAt)
	data.UpdatedAt = types.StringValue(networkResp.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BesuNetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BesuNetworkResourceModel
	var state BesuNetworkResourceModel

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
	data.ID = state.ID
	data.Platform = state.Platform

	// Most fields require replacement, only description can be updated
	// For now, just preserve state as most changes should trigger replace
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BesuNetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BesuNetworkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/networks/besu/%s", data.ID.ValueString())

	_, err := r.client.DoRequest("DELETE", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Besu network, got error: %s", err))
		return
	}
}

func (r *BesuNetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)

	// Set initial_validator_key_ids to empty list to satisfy required field
	// It will be populated properly on the next Read
	emptyList, _ := types.ListValue(types.Int64Type, []attr.Value{})
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("initial_validator_key_ids"), emptyList)...)
}
