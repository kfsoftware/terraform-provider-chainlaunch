package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &NodeAcceptInvitationResource{}

func NewNodeAcceptInvitationResource() resource.Resource {
	return &NodeAcceptInvitationResource{}
}

type NodeAcceptInvitationResource struct {
	client *Client
}

type NodeAcceptInvitationResourceModel struct {
	ID            types.String `tfsdk:"id"`
	InvitationJWT types.String `tfsdk:"invitation_jwt"`
	Success       types.Bool   `tfsdk:"success"`
	Error         types.String `tfsdk:"error"`
}

func (r *NodeAcceptInvitationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_accept_invitation"
}

func (r *NodeAcceptInvitationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Accepts a node invitation JWT to establish node sharing between Chainlaunch instances.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for this acceptance (timestamp-based)",
			},
			"invitation_jwt": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "The invitation JWT token to accept",
			},
			"success": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the invitation was successfully accepted",
			},
			"error": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Error message if acceptance failed",
			},
		},
	}
}

func (r *NodeAcceptInvitationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NodeAcceptInvitationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NodeAcceptInvitationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	acceptReq := NodeAcceptInvitationRequest{
		InvitationJWT: data.InvitationJWT.ValueString(),
	}

	body, err := r.client.DoRequest("POST", "/node/accept-invitation", acceptReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to accept invitation, got error: %s", err))
		return
	}

	var acceptResp NodeAcceptInvitationResponse
	if err := json.Unmarshal(body, &acceptResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse acceptance response: %s", err))
		return
	}

	// Use the first 20 chars of JWT as ID
	jwt := data.InvitationJWT.ValueString()
	if len(jwt) > 20 {
		data.ID = types.StringValue(jwt[:20] + "...")
	} else {
		data.ID = types.StringValue(jwt)
	}

	data.Success = types.BoolValue(acceptResp.Success)

	// Always set error field to a known value (empty string if no error)
	if acceptResp.Error != "" {
		data.Error = types.StringValue(acceptResp.Error)
	} else {
		data.Error = types.StringValue("")
	}

	if !acceptResp.Success {
		resp.Diagnostics.AddWarning("Invitation Acceptance Failed", acceptResp.Error)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NodeAcceptInvitationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NodeAcceptInvitationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Acceptance is a one-time action - just keep the existing state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NodeAcceptInvitationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Acceptance can't be updated
	resp.Diagnostics.AddError("Update Not Supported", "Node invitation acceptance cannot be updated. Delete and recreate instead.")
}

func (r *NodeAcceptInvitationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Acceptance is a one-time action - nothing to delete on the server
	// Just remove from state
}
