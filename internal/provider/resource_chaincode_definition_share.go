package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ChaincodeDefinitionShareResource{}

func NewChaincodeDefinitionShareResource() resource.Resource {
	return &ChaincodeDefinitionShareResource{}
}

type ChaincodeDefinitionShareResource struct {
	client *Client
}

type ChaincodeDefinitionShareResourceModel struct {
	ID          types.String `tfsdk:"id"`
	ChaincodeID types.String `tfsdk:"chaincode_id"`
	Version     types.String `tfsdk:"version"`
	Sequence    types.Int64  `tfsdk:"sequence"`
	DockerImage types.String `tfsdk:"docker_image"`
	Recipients  types.List   `tfsdk:"recipients"`
	Metadata    types.Map    `tfsdk:"metadata"`
	ExpiresAt   types.String `tfsdk:"expires_at"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (r *ChaincodeDefinitionShareResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_chaincode_definition_share"
}

func (r *ChaincodeDefinitionShareResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Share a chaincode definition with other Chainlaunch peer nodes.

This resource allows you to share chaincode definitions (version, sequence, docker image, endorsement policy) with specific peer nodes that have been connected via node invitations.

**Important**: This is a Pro-only feature. Both the sharing and receiving nodes must have Pro features enabled.

**Workflow**:
1. Connect to peer nodes using ` + "`chainlaunch_node_invitation`" + ` and ` + "`chainlaunch_node_accept_invitation`" + `
2. Create the chaincode definition locally using ` + "`chainlaunch_fabric_chaincode_definition`" + `
3. Use this resource to share the definition with connected peers
4. Peers can accept the shared definition on their end and install/approve/commit it

**Use Cases**:
- Share chaincode definitions across consortium members
- Coordinate chaincode upgrades across multiple organizations
- Distribute chaincode package information without manual configuration`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for this share",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"chaincode_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "ID of the chaincode to share",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Version of the chaincode definition (e.g., '1.0', '2.0')",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sequence": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Sequence number of the chaincode definition",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"docker_image": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Docker image for the chaincode (e.g., 'myregistry/mychaincode:1.0')",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"recipients": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "List of peer node IDs (connection IDs) to share the chaincode definition with. These must be connected peers from accepted invitations.",
			},
			"metadata": schema.MapAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				MarkdownDescription: "Optional metadata to include with the share (key-value pairs)",
			},
			"expires_at": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Optional expiration timestamp for the share (ISO 8601 format)",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Status of the share (e.g., 'pending', 'accepted', 'rejected')",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the share was created",
			},
		},
	}
}

func (r *ChaincodeDefinitionShareResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ChaincodeDefinitionShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ChaincodeDefinitionShareResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse recipients
	var recipients []string
	resp.Diagnostics.Append(data.Recipients.ElementsAs(ctx, &recipients, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert recipient IDs to integers
	var recipientIDs []int64
	for _, r := range recipients {
		recipientID, err := strconv.ParseInt(r, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Recipient ID",
				fmt.Sprintf("Unable to parse recipient ID '%s': %s", r, err),
			)
			return
		}
		recipientIDs = append(recipientIDs, recipientID)
	}

	if len(recipientIDs) == 0 {
		resp.Diagnostics.AddError(
			"No Recipients",
			"At least one recipient peer node ID must be specified",
		)
		return
	}

	// Build request payload
	payload := map[string]interface{}{
		"chaincodeId": data.ChaincodeID.ValueString(),
		"version":     data.Version.ValueString(),
		"sequence":    data.Sequence.ValueInt64(),
		"dockerImage": data.DockerImage.ValueString(),
		"recipients":  recipientIDs,
	}

	// Add optional fields
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		var metadata map[string]string
		resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadata, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload["metadata"] = metadata
	}

	if !data.ExpiresAt.IsNull() && !data.ExpiresAt.IsUnknown() {
		payload["expiresAt"] = data.ExpiresAt.ValueString()
	}

	// Make API call
	endpoint := "/pro/sharing/chaincode"
	body, err := r.client.DoRequest("POST", endpoint, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to share chaincode definition, got error: %s", err),
		)
		return
	}

	// Parse response
	var shareResp ChaincodeShareResponse
	if err := json.Unmarshal(body, &shareResp); err != nil {
		resp.Diagnostics.AddError(
			"Parse Error",
			fmt.Sprintf("Unable to parse share response: %s", err),
		)
		return
	}

	// Set state
	data.ID = types.StringValue(shareResp.ChaincodeID)
	data.Status = types.StringValue("pending")
	if shareResp.CreatedAt != "" {
		data.CreatedAt = types.StringValue(shareResp.CreatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ChaincodeDefinitionShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ChaincodeDefinitionShareResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Chaincode shares are ephemeral - once created, they're sent to recipients
	// We can't query individual share status, so we just keep the existing state
	// The receiving end can accept/reject via their own resources

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ChaincodeDefinitionShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ChaincodeDefinitionShareResourceModel
	var state ChaincodeDefinitionShareResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve computed fields from state
	data.ID = state.ID
	data.Status = state.Status
	data.CreatedAt = state.CreatedAt

	// Update recipients by re-sharing
	// Parse recipients
	var recipients []string
	resp.Diagnostics.Append(data.Recipients.ElementsAs(ctx, &recipients, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert recipient IDs to integers
	var recipientIDs []int64
	for _, r := range recipients {
		recipientID, err := strconv.ParseInt(r, 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid Recipient ID",
				fmt.Sprintf("Unable to parse recipient ID '%s': %s", r, err),
			)
			return
		}
		recipientIDs = append(recipientIDs, recipientID)
	}

	// Build request payload
	payload := map[string]interface{}{
		"chaincodeId": data.ChaincodeID.ValueString(),
		"version":     data.Version.ValueString(),
		"sequence":    data.Sequence.ValueInt64(),
		"dockerImage": data.DockerImage.ValueString(),
		"recipients":  recipientIDs,
	}

	// Add optional fields
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		var metadata map[string]string
		resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadata, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload["metadata"] = metadata
	}

	if !data.ExpiresAt.IsNull() && !data.ExpiresAt.IsUnknown() {
		payload["expiresAt"] = data.ExpiresAt.ValueString()
	}

	// Make API call
	endpoint := "/pro/sharing/chaincode"
	_, err := r.client.DoRequest("POST", endpoint, payload)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to update chaincode definition share, got error: %s", err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ChaincodeDefinitionShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Chaincode definition shares can't be "deleted" - once sent, they're on the recipient's end
	// The recipient can reject them if needed
	// Just remove from Terraform state
}
