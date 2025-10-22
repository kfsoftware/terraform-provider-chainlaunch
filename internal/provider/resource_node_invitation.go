package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &NodeInvitationResource{}

func NewNodeInvitationResource() resource.Resource {
	return &NodeInvitationResource{}
}

type NodeInvitationResource struct {
	client *Client
}

type NodeInvitationResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Bidirectional types.Bool   `tfsdk:"bidirectional"`
	InvitationJWT types.String `tfsdk:"invitation_jwt"`
}

func (r *NodeInvitationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_invitation"
}

func (r *NodeInvitationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates a node invitation JWT for sharing nodes between Chainlaunch instances.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for this invitation (timestamp-based)",
			},
			"bidirectional": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "If true, request bidirectional handshake (both parties can share nodes). Defaults to true.",
			},
			"invitation_jwt": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The generated JWT token for the invitation",
			},
		},
	}
}

func (r *NodeInvitationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NodeInvitationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NodeInvitationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := NodeInvitationRequest{
		Bidirectional: data.Bidirectional.ValueBool(),
	}

	body, err := r.client.DoRequest("POST", "/node/generate-invitation", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to generate invitation, got error: %s", err))
		return
	}

	var invResp NodeInvitationResponse
	if err := json.Unmarshal(body, &invResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse invitation response: %s", err))
		return
	}

	// Use the JWT as the ID (shortened for display)
	data.ID = types.StringValue(invResp.InvitationJWT[:20] + "...")
	data.InvitationJWT = types.StringValue(invResp.InvitationJWT)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NodeInvitationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NodeInvitationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Invitations are ephemeral - once generated, they don't have server-side state to query
	// Just keep the existing state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NodeInvitationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Invitations can't be updated - they're immutable
	resp.Diagnostics.AddError("Update Not Supported", "Node invitations cannot be updated. Delete and recreate instead.")
}

func (r *NodeInvitationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Invitations are ephemeral - nothing to delete on the server
	// Just remove from state
}
