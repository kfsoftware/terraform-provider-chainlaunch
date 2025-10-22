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

var _ resource.Resource = &FabricChaincodeApproveResource{}

func NewFabricChaincodeApproveResource() resource.Resource {
	return &FabricChaincodeApproveResource{}
}

type FabricChaincodeApproveResource struct {
	client *Client
}

type FabricChaincodeApproveResourceModel struct {
	ID           types.String `tfsdk:"id"`
	DefinitionID types.Int64  `tfsdk:"definition_id"`
	PeerID       types.Int64  `tfsdk:"peer_id"`
	Status       types.String `tfsdk:"status"`
	Message      types.String `tfsdk:"message"`
}

func (r *FabricChaincodeApproveResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_chaincode_approve"
}

func (r *FabricChaincodeApproveResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Approves a Fabric chaincode definition for an organization. This performs the 'peer lifecycle chaincode approveformyorg' operation. Each organization must approve the chaincode definition before it can be committed to the channel.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for this approval (format: definition_id:peer_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"definition_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the chaincode definition to approve.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"peer_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the peer to use for the approval operation.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the approval operation.",
			},
			"message": schema.StringAttribute{
				Computed:    true,
				Description: "Message from the approval operation.",
			},
		},
	}
}

func (r *FabricChaincodeApproveResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FabricChaincodeApproveResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricChaincodeApproveResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Approve chaincode request
	approveReq := struct {
		PeerID int64 `json:"peer_id"`
	}{
		PeerID: data.PeerID.ValueInt64(),
	}

	endpoint := fmt.Sprintf("/sc/fabric/definitions/%d/approve", data.DefinitionID.ValueInt64())
	body, err := r.client.DoRequest("POST", endpoint, approveReq)
	if err != nil {
		// Check if the error is "attempted to redefine the current committed sequence"
		// This means the chaincode is already approved at this sequence
		errStr := err.Error()
		if strings.Contains(errStr, "attempted to redefine the current committed sequence") {
			// Already approved - treat as success
			data.ID = types.StringValue(fmt.Sprintf("%d:%d", data.DefinitionID.ValueInt64(), data.PeerID.ValueInt64()))
			data.Status = types.StringValue("already_approved")
			data.Message = types.StringValue("Chaincode already approved at this sequence")
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}

		// Real error
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to approve chaincode: %s", err))
		return
	}

	// Parse response (it returns a map[string]string)
	var approveResp map[string]string
	if err := json.Unmarshal(body, &approveResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Set state
	data.ID = types.StringValue(fmt.Sprintf("%d:%d", data.DefinitionID.ValueInt64(), data.PeerID.ValueInt64()))

	// Extract status and message from response
	if status, ok := approveResp["status"]; ok {
		data.Status = types.StringValue(status)
	} else {
		data.Status = types.StringValue("success")
	}

	if message, ok := approveResp["message"]; ok {
		data.Message = types.StringValue(message)
	} else {
		data.Message = types.StringValue("Chaincode approved successfully")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeApproveResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricChaincodeApproveResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API doesn't provide a GET endpoint to verify chaincode approval
	// We keep the state as-is

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeApproveResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FabricChaincodeApproveResourceModel

	// Get plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Re-approve for updates
	approveReq := struct {
		PeerID int64 `json:"peer_id"`
	}{
		PeerID: data.PeerID.ValueInt64(),
	}

	endpoint := fmt.Sprintf("/sc/fabric/definitions/%d/approve", data.DefinitionID.ValueInt64())
	body, err := r.client.DoRequest("POST", endpoint, approveReq)
	if err != nil {
		// Check if the error is "attempted to redefine the current committed sequence"
		errStr := err.Error()
		if strings.Contains(errStr, "attempted to redefine the current committed sequence") {
			// Already approved - treat as success
			data.Status = types.StringValue("already_approved")
			data.Message = types.StringValue("Chaincode already approved at this sequence")
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}

		// Real error
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to re-approve chaincode: %s", err))
		return
	}

	// Parse response
	var approveResp map[string]string
	if err := json.Unmarshal(body, &approveResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Update state
	if status, ok := approveResp["status"]; ok {
		data.Status = types.StringValue(status)
	}

	if message, ok := approveResp["message"]; ok {
		data.Message = types.StringValue(message)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeApproveResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Chaincode approvals cannot be revoked in Fabric
	// Deletion just removes from Terraform state
	// No API call needed
}
