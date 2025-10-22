package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &FabricChaincodeInstallResource{}

func NewFabricChaincodeInstallResource() resource.Resource {
	return &FabricChaincodeInstallResource{}
}

type FabricChaincodeInstallResource struct {
	client *Client
}

type FabricChaincodeInstallResourceModel struct {
	ID           types.String `tfsdk:"id"`
	DefinitionID types.Int64  `tfsdk:"definition_id"`
	PeerIDs      types.List   `tfsdk:"peer_ids"`
	Status       types.String `tfsdk:"status"`
	Message      types.String `tfsdk:"message"`
}

func (r *FabricChaincodeInstallResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_chaincode_install"
}

func (r *FabricChaincodeInstallResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Installs a Fabric chaincode on peers based on a chaincode definition. This operation pulls the Docker image specified in the definition and installs it on the specified peers.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for this install operation (format: definition_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"definition_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the chaincode definition to install. The definition contains the Docker image to be used.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"peer_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.Int64Type,
				Description: "List of peer IDs to install the chaincode on.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the install operation.",
			},
			"message": schema.StringAttribute{
				Computed:    true,
				Description: "Message from the install operation.",
			},
		},
	}
}

func (r *FabricChaincodeInstallResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FabricChaincodeInstallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricChaincodeInstallResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert peer_ids list to slice
	var peerIDs []int64
	resp.Diagnostics.Append(data.PeerIDs.ElementsAs(ctx, &peerIDs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Install chaincode request
	installReq := struct {
		PeerIDs []int64 `json:"peer_ids"`
	}{
		PeerIDs: peerIDs,
	}

	endpoint := fmt.Sprintf("/sc/fabric/definitions/%d/install", data.DefinitionID.ValueInt64())
	body, err := r.client.DoRequest("POST", endpoint, installReq)
	if err != nil {
		// Check if the error is "already installed" which is not actually an error
		errStr := err.Error()
		if strings.Contains(errStr, "chaincode already successfully installed") {
			// Already installed - treat as success
			data.ID = types.StringValue(fmt.Sprintf("%d", data.DefinitionID.ValueInt64()))
			data.Status = types.StringValue("already_installed")
			data.Message = types.StringValue("Chaincode already successfully installed on peers")
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}

		// Real error
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to install chaincode on peers %v: %s\n\nPlease verify that:\n1. The peer IDs are correct and the peers exist\n2. The peers are running and healthy\n3. The definition ID %d exists", peerIDs, err, data.DefinitionID.ValueInt64()),
		)
		return
	}

	// Parse response (it returns a map[string]string)
	var installResp map[string]string
	if err := json.Unmarshal(body, &installResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Set state
	data.ID = types.StringValue(fmt.Sprintf("%d", data.DefinitionID.ValueInt64()))

	// Extract status and message from response
	if status, ok := installResp["status"]; ok {
		data.Status = types.StringValue(status)
	} else {
		data.Status = types.StringValue("success")
	}

	if message, ok := installResp["message"]; ok {
		data.Message = types.StringValue(message)
	} else {
		data.Message = types.StringValue("Chaincode installed successfully")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeInstallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricChaincodeInstallResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API doesn't provide a GET endpoint to verify chaincode installation
	// We keep the state as-is

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeInstallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FabricChaincodeInstallResourceModel

	// Get plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// All fields are marked as RequiresReplace, so this shouldn't be called
	// Including for consistency

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeInstallResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Chaincode packages cannot be uninstalled from peers in Fabric
	// Deletion just removes from Terraform state
	// No API call needed
}
