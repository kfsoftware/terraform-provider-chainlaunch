package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &NetworkResource{}
var _ resource.ResourceWithImportState = &NetworkResource{}

func NewNetworkResource() resource.Resource {
	return &NetworkResource{}
}

type NetworkResource struct {
	client *Client
}

type NetworkResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Status    types.String `tfsdk:"status"`
	Config    types.String `tfsdk:"config"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *NetworkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (r *NetworkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a blockchain network in Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the network.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the network.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of network (e.g., fabric, besu).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the network.",
			},
			"config": schema.StringAttribute{
				Optional:    true,
				Description: "JSON configuration for the network.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the network was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the network was last updated.",
			},
		},
	}
}

func (r *NetworkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NetworkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NetworkResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := CreateNetworkRequest{
		Name: data.Name.ValueString(),
		Type: data.Type.ValueString(),
	}

	if !data.Config.IsNull() && data.Config.ValueString() != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(data.Config.ValueString()), &config); err != nil {
			resp.Diagnostics.AddError("Config Parse Error", fmt.Sprintf("Unable to parse config JSON: %s", err))
			return
		}
		createReq.Config = config
	}

	// Determine the API endpoint based on network type
	endpoint := fmt.Sprintf("/networks/%s", data.Type.ValueString())

	body, err := r.client.DoRequest("POST", endpoint, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create network, got error: %s", err))
		return
	}

	var network Network
	if err := json.Unmarshal(body, &network); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse network response, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", network.ID))
	data.Name = types.StringValue(network.Name)
	data.Type = types.StringValue(network.Type)
	if network.Status != "" {
		data.Status = types.StringValue(network.Status)
	}
	if network.CreatedAt != "" {
		data.CreatedAt = types.StringValue(network.CreatedAt)
	}
	if network.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(network.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NetworkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/networks/%s/%s", data.Type.ValueString(), data.ID.ValueString())

	body, err := r.client.DoRequest("GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read network, got error: %s", err))
		return
	}

	var network Network
	if err := json.Unmarshal(body, &network); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse network response, got error: %s", err))
		return
	}

	data.Name = types.StringValue(network.Name)
	data.Type = types.StringValue(network.Type)
	if network.Status != "" {
		data.Status = types.StringValue(network.Status)
	}
	if network.CreatedAt != "" {
		data.CreatedAt = types.StringValue(network.CreatedAt)
	}
	if network.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(network.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NetworkResourceModel
	var state NetworkResourceModel

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

	updateReq := CreateNetworkRequest{
		Name: data.Name.ValueString(),
		Type: data.Type.ValueString(),
	}

	if !data.Config.IsNull() && data.Config.ValueString() != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(data.Config.ValueString()), &config); err != nil {
			resp.Diagnostics.AddError("Config Parse Error", fmt.Sprintf("Unable to parse config JSON: %s", err))
			return
		}
		updateReq.Config = config
	}

	endpoint := fmt.Sprintf("/networks/%s/%s", data.Type.ValueString(), data.ID.ValueString())

	body, err := r.client.DoRequest("PUT", endpoint, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update network, got error: %s", err))
		return
	}

	var network Network
	if err := json.Unmarshal(body, &network); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse network response, got error: %s", err))
		return
	}

	data.Name = types.StringValue(network.Name)
	if network.Status != "" {
		data.Status = types.StringValue(network.Status)
	}
	if network.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(network.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NetworkResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/networks/%s/%s", data.Type.ValueString(), data.ID.ValueString())

	_, err := r.client.DoRequest("DELETE", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete network, got error: %s", err))
		return
	}
}

func (r *NetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
