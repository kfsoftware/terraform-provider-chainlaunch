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

var (
	_ resource.Resource              = &FabricIdentityResource{}
	_ resource.ResourceWithConfigure = &FabricIdentityResource{}
)

func NewFabricIdentityResource() resource.Resource {
	return &FabricIdentityResource{}
}

type FabricIdentityResource struct {
	client *Client
}

type FabricIdentityResourceModel struct {
	ID              types.String `tfsdk:"id"`
	OrganizationID  types.String `tfsdk:"organization_id"`
	Name            types.String `tfsdk:"name"`
	Role            types.String `tfsdk:"role"`
	Description     types.String `tfsdk:"description"`
	DNSNames        types.List   `tfsdk:"dns_names"`
	IPAddresses     types.List   `tfsdk:"ip_addresses"`
	Algorithm       types.String `tfsdk:"algorithm"`
	Certificate     types.String `tfsdk:"certificate"`
	PublicKey       types.String `tfsdk:"public_key"`
	SHA1Fingerprint types.String `tfsdk:"sha1_fingerprint"`
	EthereumAddress types.String `tfsdk:"ethereum_address"`
	ExpiresAt       types.String `tfsdk:"expires_at"`
	LastRotatedAt   types.String `tfsdk:"last_rotated_at"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

func (r *FabricIdentityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_identity"
}

func (r *FabricIdentityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Fabric identity (admin or client) for an organization. Creates a key pair and certificate that can be used for Fabric operations.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the identity (key ID)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "ID of the Fabric organization this identity belongs to",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the identity (e.g., 'admin', 'client1')",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Description: "Role of the identity ('admin' or 'client')",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the identity",
				Optional:    true,
			},
			"dns_names": schema.ListAttribute{
				Description: "DNS names to include in the certificate (optional)",
				Optional:    true,
				ElementType: types.StringType,
			},
			"ip_addresses": schema.ListAttribute{
				Description: "IP addresses to include in the certificate (optional)",
				Optional:    true,
				ElementType: types.StringType,
			},
			"algorithm": schema.StringAttribute{
				Description: "Algorithm used for the key (e.g., 'ECDSA')",
				Computed:    true,
			},
			"certificate": schema.StringAttribute{
				Description: "PEM-encoded certificate",
				Computed:    true,
				Sensitive:   true,
			},
			"public_key": schema.StringAttribute{
				Description: "PEM-encoded public key",
				Computed:    true,
			},
			"sha1_fingerprint": schema.StringAttribute{
				Description: "SHA1 fingerprint of the certificate",
				Computed:    true,
			},
			"ethereum_address": schema.StringAttribute{
				Description: "Ethereum address derived from the key (if applicable)",
				Computed:    true,
			},
			"expires_at": schema.StringAttribute{
				Description: "Certificate expiration timestamp",
				Computed:    true,
			},
			"last_rotated_at": schema.StringAttribute{
				Description: "Last rotation timestamp",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "Creation timestamp",
				Computed:    true,
			},
		},
	}
}

func (r *FabricIdentityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FabricIdentityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricIdentityResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare request body
	createReq := map[string]interface{}{
		"name": data.Name.ValueString(),
		"role": data.Role.ValueString(),
	}

	if !data.Description.IsNull() {
		createReq["description"] = data.Description.ValueString()
	}

	// Add DNS names if provided
	if !data.DNSNames.IsNull() {
		var dnsNames []string
		data.DNSNames.ElementsAs(ctx, &dnsNames, false)
		if len(dnsNames) > 0 {
			createReq["dnsNames"] = dnsNames
		}
	}

	// Add IP addresses if provided
	if !data.IPAddresses.IsNull() {
		var ipAddresses []string
		data.IPAddresses.ElementsAs(ctx, &ipAddresses, false)
		if len(ipAddresses) > 0 {
			createReq["ipAddresses"] = ipAddresses
		}
	}

	// Create identity via API
	identityResp, err := r.client.DoRequest("POST", fmt.Sprintf("/organizations/%s/keys", data.OrganizationID.ValueString()), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Fabric identity", err.Error())
		return
	}

	// Parse response
	var identityResult map[string]interface{}
	if err := json.Unmarshal(identityResp, &identityResult); err != nil {
		resp.Diagnostics.AddError("Failed to parse identity response", err.Error())
		return
	}

	// Set computed fields
	if id, ok := identityResult["id"].(float64); ok {
		data.ID = types.StringValue(fmt.Sprintf("%d", int64(id)))
	}
	if algorithm, ok := identityResult["algorithm"].(string); ok {
		data.Algorithm = types.StringValue(algorithm)
	}
	if certificate, ok := identityResult["certificate"].(string); ok {
		data.Certificate = types.StringValue(certificate)
	}
	if publicKey, ok := identityResult["publicKey"].(string); ok {
		data.PublicKey = types.StringValue(publicKey)
	}
	if sha1Fingerprint, ok := identityResult["sha1Fingerprint"].(string); ok {
		data.SHA1Fingerprint = types.StringValue(sha1Fingerprint)
	}

	// Handle optional fields - set to null if not provided
	if ethereumAddress, ok := identityResult["ethereumAddress"].(string); ok && ethereumAddress != "" {
		data.EthereumAddress = types.StringValue(ethereumAddress)
	} else {
		data.EthereumAddress = types.StringNull()
	}

	if expiresAt, ok := identityResult["expiresAt"].(string); ok && expiresAt != "" {
		data.ExpiresAt = types.StringValue(expiresAt)
	} else {
		data.ExpiresAt = types.StringNull()
	}

	if lastRotatedAt, ok := identityResult["lastRotatedAt"].(string); ok && lastRotatedAt != "" {
		data.LastRotatedAt = types.StringValue(lastRotatedAt)
	} else {
		data.LastRotatedAt = types.StringNull()
	}

	if createdAt, ok := identityResult["createdAt"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricIdentityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricIdentityResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get identity from API
	identityResp, err := r.client.DoRequest("GET", fmt.Sprintf("/organizations/%s/keys", data.OrganizationID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read identity", err.Error())
		return
	}

	// Parse response (map of keys)
	var keysResult map[string]interface{}
	if err := json.Unmarshal(identityResp, &keysResult); err != nil {
		resp.Diagnostics.AddError("Failed to parse keys response", err.Error())
		return
	}

	// Find our key in the map
	// The API returns: {"keys": {"1": {...}, "2": {...}}}
	keysMap, ok := keysResult["keys"].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("Invalid keys response", "Expected 'keys' object")
		return
	}

	var identityResult map[string]interface{}
	keyID := data.ID.ValueString()
	found := false

	// Look up our key by ID
	if keyData, exists := keysMap[keyID]; exists {
		if keyMap, ok := keyData.(map[string]interface{}); ok {
			identityResult = keyMap
			found = true
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update computed fields
	if algorithm, ok := identityResult["algorithm"].(string); ok {
		data.Algorithm = types.StringValue(algorithm)
	}
	if certificate, ok := identityResult["certificate"].(string); ok {
		data.Certificate = types.StringValue(certificate)
	}
	if publicKey, ok := identityResult["publicKey"].(string); ok {
		data.PublicKey = types.StringValue(publicKey)
	}
	if sha1Fingerprint, ok := identityResult["sha1Fingerprint"].(string); ok {
		data.SHA1Fingerprint = types.StringValue(sha1Fingerprint)
	}

	// Handle optional fields - set to null if not provided
	if ethereumAddress, ok := identityResult["ethereumAddress"].(string); ok && ethereumAddress != "" {
		data.EthereumAddress = types.StringValue(ethereumAddress)
	} else {
		data.EthereumAddress = types.StringNull()
	}

	if expiresAt, ok := identityResult["expiresAt"].(string); ok && expiresAt != "" {
		data.ExpiresAt = types.StringValue(expiresAt)
	} else {
		data.ExpiresAt = types.StringNull()
	}

	if lastRotatedAt, ok := identityResult["lastRotatedAt"].(string); ok && lastRotatedAt != "" {
		data.LastRotatedAt = types.StringValue(lastRotatedAt)
	} else {
		data.LastRotatedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricIdentityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FabricIdentityResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Most fields require replacement, only description can be updated
	// Since description updates aren't supported by the API, we just preserve state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricIdentityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FabricIdentityResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete identity via API
	_, err := r.client.DoRequest("DELETE", fmt.Sprintf("/organizations/%s/keys/%s", data.OrganizationID.ValueString(), data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete Fabric identity", err.Error())
		return
	}
}

func (r *FabricIdentityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: "organization_id/key_id"
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
