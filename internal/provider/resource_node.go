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

var _ resource.Resource = &NodeResource{}
var _ resource.ResourceWithImportState = &NodeResource{}

func NewNodeResource() resource.Resource {
	return &NodeResource{}
}

type NodeResource struct {
	client *Client
}

type NodeResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Platform  types.String `tfsdk:"platform"`
	Type      types.String `tfsdk:"type"`
	Status    types.String `tfsdk:"status"`
	Config    types.String `tfsdk:"config"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (r *NodeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

func (r *NodeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a blockchain node in Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the node.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the node.",
			},
			"platform": schema.StringAttribute{
				Required:    true,
				Description: "The blockchain platform (e.g., fabric, besu).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of node (e.g., peer, orderer).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the node.",
			},
			"config": schema.StringAttribute{
				Optional:    true,
				Description: "JSON configuration for the node.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the node was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the node was last updated.",
			},
		},
	}
}

func (r *NodeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NodeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NodeResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := CreateNodeRequest{
		Name:     data.Name.ValueString(),
		Platform: data.Platform.ValueString(),
		Type:     data.Type.ValueString(),
	}

	if !data.Config.IsNull() && data.Config.ValueString() != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(data.Config.ValueString()), &config); err != nil {
			resp.Diagnostics.AddError("Config Parse Error", fmt.Sprintf("Unable to parse config JSON: %s", err))
			return
		}
		createReq.Config = config
	}

	body, err := r.client.DoRequest("POST", "/nodes", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create node, got error: %s", err))
		return
	}

	var node Node
	if err := json.Unmarshal(body, &node); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse node response, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", node.ID))
	data.Name = types.StringValue(node.Name)
	data.Platform = types.StringValue(node.Platform)
	data.Type = types.StringValue(node.Type)
	if node.Status != "" {
		data.Status = types.StringValue(node.Status)
	}
	if node.CreatedAt != "" {
		data.CreatedAt = types.StringValue(node.CreatedAt)
	}
	if node.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(node.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NodeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NodeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.DoRequest("GET", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read node, got error: %s", err))
		return
	}

	var node Node
	if err := json.Unmarshal(body, &node); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse node response, got error: %s", err))
		return
	}

	data.Name = types.StringValue(node.Name)
	data.Platform = types.StringValue(node.Platform)
	data.Type = types.StringValue(node.Type)
	if node.Status != "" {
		data.Status = types.StringValue(node.Status)
	}
	if node.CreatedAt != "" {
		data.CreatedAt = types.StringValue(node.CreatedAt)
	}
	if node.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(node.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NodeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NodeResourceModel
	var state NodeResourceModel

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

	updateReq := CreateNodeRequest{
		Name:     data.Name.ValueString(),
		Platform: data.Platform.ValueString(),
		Type:     data.Type.ValueString(),
	}

	if !data.Config.IsNull() && data.Config.ValueString() != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(data.Config.ValueString()), &config); err != nil {
			resp.Diagnostics.AddError("Config Parse Error", fmt.Sprintf("Unable to parse config JSON: %s", err))
			return
		}
		updateReq.Config = config
	}

	body, err := r.client.DoRequest("PUT", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update node, got error: %s", err))
		return
	}

	var node Node
	if err := json.Unmarshal(body, &node); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse node response, got error: %s", err))
		return
	}

	data.Name = types.StringValue(node.Name)
	if node.Status != "" {
		data.Status = types.StringValue(node.Status)
	}
	if node.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(node.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NodeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NodeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DoRequest("DELETE", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete node, got error: %s", err))
		return
	}
}

func (r *NodeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
