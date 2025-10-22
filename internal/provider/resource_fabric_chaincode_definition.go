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

var _ resource.Resource = &FabricChaincodeDefinitionResource{}

func NewFabricChaincodeDefinitionResource() resource.Resource {
	return &FabricChaincodeDefinitionResource{}
}

type FabricChaincodeDefinitionResource struct {
	client *Client
}

type FabricChaincodeDefinitionResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	ChaincodeID       types.Int64  `tfsdk:"chaincode_id"`
	Version           types.String `tfsdk:"version"`
	Sequence          types.Int64  `tfsdk:"sequence"`
	DockerImage       types.String `tfsdk:"docker_image"`
	EndorsementPolicy types.String `tfsdk:"endorsement_policy"`
	ChaincodeAddress  types.String `tfsdk:"chaincode_address"`
	CreatedAt         types.String `tfsdk:"created_at"`
}

func (r *FabricChaincodeDefinitionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_chaincode_definition"
}

func (r *FabricChaincodeDefinitionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fabric chaincode definition. A chaincode can have multiple definitions with different versions and sequences. " +
			"Definitions are immutable - any changes will create a new definition and destroy the old one. " +
			"To upgrade chaincode, create a new definition resource with an incremented sequence number.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier for the chaincode definition.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"chaincode_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the chaincode this definition belongs to.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				Required:    true,
				Description: "The version of the chaincode (e.g., '1.0', '2.0'). Must be incremented for upgrades. Changes require replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sequence": schema.Int64Attribute{
				Required:    true,
				Description: "The sequence number for this definition. Must be incremented for each new definition on the same channel. Changes require replacement.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"docker_image": schema.StringAttribute{
				Required:    true,
				Description: "The Docker image for the chaincode (e.g., 'myregistry/mychaincode:1.0'). Changes require replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"endorsement_policy": schema.StringAttribute{
				Optional:    true,
				Description: "The endorsement policy using Fabric's policy expression language (e.g., \"OR('Org1MSP.member', 'Org2MSP.member')\"). Changes require replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"chaincode_address": schema.StringAttribute{
				Required:    true,
				Description: "The chaincode address for chaincode-as-a-service deployments (e.g., 'mycc.example.com:7052'). Changes require replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the definition was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *FabricChaincodeDefinitionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FabricChaincodeDefinitionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricChaincodeDefinitionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create definition request
	createReq := struct {
		ChaincodeID       int64  `json:"chaincode_id"`
		Version           string `json:"version"`
		Sequence          int64  `json:"sequence"`
		DockerImage       string `json:"docker_image"`
		EndorsementPolicy string `json:"endorsement_policy,omitempty"`
		ChaincodeAddress  string `json:"chaincode_address,omitempty"`
	}{
		ChaincodeID:       data.ChaincodeID.ValueInt64(),
		Version:           data.Version.ValueString(),
		Sequence:          data.Sequence.ValueInt64(),
		DockerImage:       data.DockerImage.ValueString(),
		EndorsementPolicy: data.EndorsementPolicy.ValueString(),
		ChaincodeAddress:  data.ChaincodeAddress.ValueString(),
	}

	endpoint := fmt.Sprintf("/sc/fabric/chaincodes/%d/definitions", data.ChaincodeID.ValueInt64())
	body, err := r.client.DoRequest("POST", endpoint, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create chaincode definition: %s", err))
		return
	}

	var createResp struct {
		Definition struct {
			ID                int64  `json:"id"`
			ChaincodeID       int64  `json:"chaincode_id"`
			Version           string `json:"version"`
			Sequence          int64  `json:"sequence"`
			DockerImage       string `json:"docker_image"`
			EndorsementPolicy string `json:"endorsement_policy"`
			ChaincodeAddress  string `json:"chaincode_address"`
			CreatedAt         string `json:"created_at"`
		} `json:"definition"`
	}
	if err := json.Unmarshal(body, &createResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Set state
	data.ID = types.Int64Value(createResp.Definition.ID)
	data.ChaincodeID = types.Int64Value(createResp.Definition.ChaincodeID)
	data.Version = types.StringValue(createResp.Definition.Version)
	data.Sequence = types.Int64Value(createResp.Definition.Sequence)
	data.DockerImage = types.StringValue(createResp.Definition.DockerImage)
	data.EndorsementPolicy = types.StringValue(createResp.Definition.EndorsementPolicy)
	data.ChaincodeAddress = types.StringValue(createResp.Definition.ChaincodeAddress)
	data.CreatedAt = types.StringValue(createResp.Definition.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeDefinitionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricChaincodeDefinitionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get definition details
	endpoint := fmt.Sprintf("/sc/fabric/chaincodes/%d/definitions/%d", data.ChaincodeID.ValueInt64(), data.ID.ValueInt64())
	body, err := r.client.DoRequest("GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read chaincode definition: %s", err))
		return
	}

	var readResp struct {
		Definition struct {
			ID                int64  `json:"id"`
			ChaincodeID       int64  `json:"chaincode_id"`
			Version           string `json:"version"`
			Sequence          int64  `json:"sequence"`
			DockerImage       string `json:"docker_image"`
			EndorsementPolicy string `json:"endorsement_policy"`
			ChaincodeAddress  string `json:"chaincode_address"`
			CreatedAt         string `json:"created_at"`
		} `json:"definition"`
	}
	if err := json.Unmarshal(body, &readResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Update state
	data.ID = types.Int64Value(readResp.Definition.ID)
	data.ChaincodeID = types.Int64Value(readResp.Definition.ChaincodeID)
	data.Version = types.StringValue(readResp.Definition.Version)
	data.Sequence = types.Int64Value(readResp.Definition.Sequence)
	data.DockerImage = types.StringValue(readResp.Definition.DockerImage)
	data.EndorsementPolicy = types.StringValue(readResp.Definition.EndorsementPolicy)
	data.ChaincodeAddress = types.StringValue(readResp.Definition.ChaincodeAddress)
	data.CreatedAt = types.StringValue(readResp.Definition.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeDefinitionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All fields are marked as RequiresReplace, so this method should never be called
	// Chaincode definitions are immutable in Fabric - any change requires creating a new definition
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Chaincode definitions are immutable and cannot be updated. Any changes require creating a new definition resource with an incremented sequence number. "+
			"This error should not occur as all fields are marked with RequiresReplace.",
	)
}

func (r *FabricChaincodeDefinitionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FabricChaincodeDefinitionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the definition
	endpoint := fmt.Sprintf("/sc/fabric/definitions/%d", data.ID.ValueInt64())
	_, err := r.client.DoRequest("DELETE", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete chaincode definition: %s", err))
		return
	}
}
