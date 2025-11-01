package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &OrganizationResource{}
var _ resource.ResourceWithImportState = &OrganizationResource{}

func NewOrganizationResource() resource.Resource {
	return &OrganizationResource{}
}

// OrganizationResource defines the resource implementation.
type OrganizationResource struct {
	client *Client
}

// OrganizationResourceModel describes the resource data model.
type OrganizationResourceModel struct {
	ID          types.String `tfsdk:"id"`
	MSPID       types.String `tfsdk:"msp_id"`
	Description types.String `tfsdk:"description"`
	ProviderID  types.Int64  `tfsdk:"provider_id"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *OrganizationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_organization"
}

func (r *OrganizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Hyperledger Fabric organization in Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"msp_id": schema.StringAttribute{
				Required:    true,
				Description: "The MSP ID of the organization. This will be used as the organization name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the organization.",
			},
			"provider_id": schema.Int64Attribute{
				Optional:    true,
				Description: "The ID of the key management provider to use for this organization.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the organization was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the organization was last updated.",
			},
		},
	}
}

func (r *OrganizationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrganizationResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Use msp_id as the name
	createReq := CreateOrganizationRequest{
		Name:        data.MSPID.ValueString(),
		MSPID:       data.MSPID.ValueString(),
		Description: data.Description.ValueString(),
	}

	if !data.ProviderID.IsNull() {
		createReq.ProviderID = int(data.ProviderID.ValueInt64())
	}

	body, err := r.client.DoRequest("POST", "/organizations", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create organization, got error: %s", err))
		return
	}

	// Try to parse as wrapped response first
	var wrappedResp struct {
		Data Organization `json:"data"`
	}
	var org Organization

	if err := json.Unmarshal(body, &wrappedResp); err == nil && wrappedResp.Data.ID != 0 {
		org = wrappedResp.Data
	} else if err := json.Unmarshal(body, &org); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse organization response, got error: %s\nResponse body: %s", err, string(body)))
		return
	}

	// Validate that we have a valid organization ID before saving to state
	if org.ID == 0 {
		resp.Diagnostics.AddError(
			"Invalid Organization Response",
			fmt.Sprintf("Organization creation failed: API returned invalid or empty organization ID. Response body: %s", string(body)),
		)
		return
	}

	// Validate that MSPID matches what was requested
	if org.MSPID != data.MSPID.ValueString() {
		resp.Diagnostics.AddError(
			"MSPID Mismatch",
			fmt.Sprintf("Expected MSPID %s but got %s from API", data.MSPID.ValueString(), org.MSPID),
		)
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", org.ID))
	data.MSPID = types.StringValue(org.MSPID)
	if org.Description != "" {
		data.Description = types.StringValue(org.Description)
	}
	if org.CreatedAt != "" {
		data.CreatedAt = types.StringValue(org.CreatedAt)
	}
	if org.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(org.UpdatedAt)
	}

	// Only save to state if no errors occurred
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrganizationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.DoRequest("GET", fmt.Sprintf("/organizations/%s", data.ID.ValueString()), nil)
	if err != nil {
		// Check if the error is a NOT_FOUND error
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "not_found") || strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "404") {
			// Organization was deleted outside of Terraform - remove from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization, got error: %s", err))
		return
	}

	var org Organization
	if err := json.Unmarshal(body, &org); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse organization response, got error: %s", err))
		return
	}

	// Validate that we got a valid organization
	if org.ID == 0 {
		// Invalid response - organization might have been deleted
		resp.State.RemoveResource(ctx)
		return
	}

	data.MSPID = types.StringValue(org.MSPID)
	if org.Description != "" {
		data.Description = types.StringValue(org.Description)
	}
	if org.CreatedAt != "" {
		data.CreatedAt = types.StringValue(org.CreatedAt)
	}
	if org.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(org.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrganizationResourceModel
	var state OrganizationResourceModel

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

	// Use msp_id as the name
	updateReq := CreateOrganizationRequest{
		Name:        data.MSPID.ValueString(),
		MSPID:       data.MSPID.ValueString(),
		Description: data.Description.ValueString(),
	}

	body, err := r.client.DoRequest("PUT", fmt.Sprintf("/organizations/%s", data.ID.ValueString()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update organization, got error: %s", err))
		return
	}

	var org Organization
	if err := json.Unmarshal(body, &org); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse organization response, got error: %s", err))
		return
	}

	data.MSPID = types.StringValue(org.MSPID)
	if org.Description != "" {
		data.Description = types.StringValue(org.Description)
	}
	if org.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(org.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data OrganizationResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DoRequest("DELETE", fmt.Sprintf("/organizations/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete organization, got error: %s", err))
		return
	}
}

func (r *OrganizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
