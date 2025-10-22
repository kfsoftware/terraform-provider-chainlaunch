package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KeyResource{}
var _ resource.ResourceWithImportState = &KeyResource{}

func NewKeyResource() resource.Resource {
	return &KeyResource{}
}

// KeyResource defines the resource implementation.
type KeyResource struct {
	client *Client
}

// KeyResourceModel describes the resource data model.
type KeyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Algorithm   types.String `tfsdk:"algorithm"`
	Curve       types.String `tfsdk:"curve"`
	KeySize     types.Int64  `tfsdk:"key_size"`
	ProviderID  types.Int64  `tfsdk:"provider_id"`
	IsCA        types.Bool   `tfsdk:"is_ca"`
	Description types.String `tfsdk:"description"`
	PublicKey   types.String `tfsdk:"public_key"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (r *KeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key"
}

func (r *KeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a cryptographic key in Chainlaunch. Supports RSA, EC, and ED25519 algorithms with various curves and key sizes.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Key identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the key",
				Required:            true,
			},
			"algorithm": schema.StringAttribute{
				MarkdownDescription: "Key algorithm. Valid values: RSA, EC, ED25519",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"curve": schema.StringAttribute{
				MarkdownDescription: "Elliptic curve for EC algorithm. Valid values: P-256, P-384, P-521, secp256k1. Note: Vault does not support secp256k1",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"key_size": schema.Int64Attribute{
				MarkdownDescription: "Key size in bits for RSA algorithm. Common values: 2048, 4096",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"provider_id": schema.Int64Attribute{
				MarkdownDescription: "ID of the key provider to use for storing this key",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"is_ca": schema.BoolAttribute{
				MarkdownDescription: "Whether this key is a Certificate Authority key (defaults to false)",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the key",
				Optional:            true,
			},
			"public_key": schema.StringAttribute{
				MarkdownDescription: "Public key in PEM format (computed)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp when the key was created (computed)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *KeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KeyResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build the request payload
	payload := map[string]interface{}{
		"name":       data.Name.ValueString(),
		"algorithm":  data.Algorithm.ValueString(),
		"providerId": data.ProviderID.ValueInt64(),
	}

	// Add optional fields
	if !data.Curve.IsNull() {
		payload["curve"] = data.Curve.ValueString()
	}

	if !data.KeySize.IsNull() {
		payload["keySize"] = data.KeySize.ValueInt64()
	}

	// Default is_ca to false if not specified
	if !data.IsCA.IsNull() {
		if data.IsCA.ValueBool() {
			payload["isCA"] = 1
		} else {
			payload["isCA"] = 0
		}
	} else {
		payload["isCA"] = 0 // Default to false
		data.IsCA = types.BoolValue(false)
	}

	if !data.Description.IsNull() {
		payload["description"] = data.Description.ValueString()
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to marshal key request: %s", err))
		return
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/keys", r.client.BaseURL), strings.NewReader(string(jsonData)))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if r.client.APIKey != "" {
		httpReq.Header.Set("X-API-Key", r.client.APIKey)
	} else {
		httpReq.SetBasicAuth(r.client.Username, r.client.Password)
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create key: %s", err))
		return
	}
	defer func() { _ = httpResp.Body.Close() }()

	body, _ := io.ReadAll(httpResp.Body)

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Unable to create key, status code: %d, body: %s", httpResp.StatusCode, string(body)),
		)
		return
	}

	var keyResp struct {
		ID        int64  `json:"id"`
		PublicKey string `json:"publicKey"`
		CreatedAt string `json:"createdAt"`
	}

	if err := json.Unmarshal(body, &keyResp); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse key response: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", keyResp.ID))
	data.PublicKey = types.StringValue(keyResp.PublicKey)
	data.CreatedAt = types.StringValue(keyResp.CreatedAt)

	// Ensure is_ca is set to a known value (false if not specified)
	if data.IsCA.IsNull() {
		data.IsCA = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/keys/%s", r.client.BaseURL, data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	if r.client.APIKey != "" {
		httpReq.Header.Set("X-API-Key", r.client.APIKey)
	} else {
		httpReq.SetBasicAuth(r.client.Username, r.client.Password)
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read key: %s", err))
		return
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	body, _ := io.ReadAll(httpResp.Body)

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Unable to read key, status code: %d, body: %s", httpResp.StatusCode, string(body)),
		)
		return
	}

	var key struct {
		ID          int64  `json:"id"`
		Algorithm   string `json:"algorithm"`
		Curve       string `json:"curve"`
		KeySize     int64  `json:"keySize"`
		IsCA        bool   `json:"isCA"`
		Description string `json:"description"`
		PublicKey   string `json:"publicKey"`
		CreatedAt   string `json:"createdAt"`
	}

	if err := json.Unmarshal(body, &key); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse key response: %s", err))
		return
	}

	// Preserve name from state (API may not return it)
	// data.Name is already set from req.State.Get(ctx, &data)

	data.Algorithm = types.StringValue(key.Algorithm)

	if key.Curve != "" {
		data.Curve = types.StringValue(key.Curve)
	}

	if key.KeySize > 0 {
		data.KeySize = types.Int64Value(key.KeySize)
	}

	// Set is_ca from API, defaulting to false if not present
	data.IsCA = types.BoolValue(key.IsCA)

	if key.Description != "" {
		data.Description = types.StringValue(key.Description)
	}

	data.PublicKey = types.StringValue(key.PublicKey)
	data.CreatedAt = types.StringValue(key.CreatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KeyResourceModel
	var state KeyResourceModel

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

	// Build the update payload (only mutable fields)
	payload := map[string]interface{}{}

	if !data.Description.IsNull() {
		payload["description"] = data.Description.ValueString()
	}

	// Default is_ca to false if not specified
	if !data.IsCA.IsNull() {
		if data.IsCA.ValueBool() {
			payload["isCA"] = 1
		} else {
			payload["isCA"] = 0
		}
	} else {
		payload["isCA"] = 0 // Default to false
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to marshal key update request: %s", err))
		return
	}

	httpReq, err := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/keys/%s", r.client.BaseURL, data.ID.ValueString()), strings.NewReader(string(jsonData)))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if r.client.APIKey != "" {
		httpReq.Header.Set("X-API-Key", r.client.APIKey)
	} else {
		httpReq.SetBasicAuth(r.client.Username, r.client.Password)
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update key: %s", err))
		return
	}
	defer func() { _ = httpResp.Body.Close() }()

	body, _ := io.ReadAll(httpResp.Body)

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Unable to update key, status code: %d, body: %s", httpResp.StatusCode, string(body)),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KeyResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/keys/%s", r.client.BaseURL, data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request: %s", err))
		return
	}

	if r.client.APIKey != "" {
		httpReq.Header.Set("X-API-Key", r.client.APIKey)
	} else {
		httpReq.SetBasicAuth(r.client.Username, r.client.Password)
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete key: %s", err))
		return
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Unable to delete key, status code: %d, body: %s", httpResp.StatusCode, string(body)),
		)
		return
	}
}

func (r *KeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	data := KeyResourceModel{
		ID: types.StringValue(req.ID),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
