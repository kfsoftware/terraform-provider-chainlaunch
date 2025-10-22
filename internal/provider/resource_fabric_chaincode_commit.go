package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &FabricChaincodeCommitResource{}

func NewFabricChaincodeCommitResource() resource.Resource {
	return &FabricChaincodeCommitResource{}
}

type FabricChaincodeCommitResource struct {
	client *Client
}

type FabricChaincodeCommitResourceModel struct {
	ID           types.String `tfsdk:"id"`
	DefinitionID types.Int64  `tfsdk:"definition_id"`
	PeerID       types.Int64  `tfsdk:"peer_id"`
	Status       types.String `tfsdk:"status"`
	Message      types.String `tfsdk:"message"`
}

func (r *FabricChaincodeCommitResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_chaincode_commit"
}

func (r *FabricChaincodeCommitResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Commits a Fabric chaincode definition to a channel. This performs the 'peer lifecycle chaincode commit' operation. The chaincode definition must be approved by sufficient organizations (based on the lifecycle endorsement policy) before it can be committed.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for this commit (format: definition_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"definition_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the chaincode definition to commit.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"peer_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the peer to use for the commit operation.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the commit operation.",
			},
			"message": schema.StringAttribute{
				Computed:    true,
				Description: "Message from the commit operation.",
			},
		},
	}
}

func (r *FabricChaincodeCommitResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FabricChaincodeCommitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricChaincodeCommitResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Commit chaincode request
	commitReq := struct {
		PeerID int64 `json:"peer_id"`
	}{
		PeerID: data.PeerID.ValueInt64(),
	}

	endpoint := fmt.Sprintf("/sc/fabric/definitions/%d/commit", data.DefinitionID.ValueInt64())
	body, err := r.client.DoRequest("POST", endpoint, commitReq)
	if err != nil {
		// Check if the error is "attempted to redefine the current committed sequence"
		// This means the chaincode is already committed at this sequence
		errStr := err.Error()
		if strings.Contains(errStr, "attempted to redefine the current committed sequence") {
			// Already committed - treat as success
			data.ID = types.StringValue(fmt.Sprintf("%d", data.DefinitionID.ValueInt64()))
			data.Status = types.StringValue("already_committed")
			data.Message = types.StringValue("Chaincode already committed at this sequence")
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}

		// Real error
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to commit chaincode: %s", err))
		return
	}

	// Parse response (it returns a map[string]string)
	var commitResp map[string]string
	if err := json.Unmarshal(body, &commitResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Set state
	data.ID = types.StringValue(fmt.Sprintf("%d", data.DefinitionID.ValueInt64()))

	// Extract status and message from response
	if status, ok := commitResp["status"]; ok {
		data.Status = types.StringValue(status)
	} else {
		data.Status = types.StringValue("success")
	}

	if message, ok := commitResp["message"]; ok {
		data.Message = types.StringValue(message)
	} else {
		data.Message = types.StringValue("Chaincode committed successfully")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeCommitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricChaincodeCommitResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API doesn't provide a GET endpoint to verify chaincode commit
	// We keep the state as-is

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeCommitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FabricChaincodeCommitResourceModel

	// Get plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Re-commit for updates
	commitReq := struct {
		PeerID int64 `json:"peer_id"`
	}{
		PeerID: data.PeerID.ValueInt64(),
	}

	endpoint := fmt.Sprintf("/sc/fabric/definitions/%d/commit", data.DefinitionID.ValueInt64())
	body, err := r.client.DoRequest("POST", endpoint, commitReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to re-commit chaincode: %s", err))
		return
	}

	// Parse response
	var commitResp map[string]string
	if err := json.Unmarshal(body, &commitResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Update state
	if status, ok := commitResp["status"]; ok {
		data.Status = types.StringValue(status)
	}

	if message, ok := commitResp["message"]; ok {
		data.Message = types.StringValue(message)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeCommitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Chaincode commits cannot be reverted in Fabric
	// Deletion just removes from Terraform state
	// No API call needed
}
