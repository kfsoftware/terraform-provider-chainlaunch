package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &KeyDataSource{}

func NewKeyDataSource() datasource.DataSource {
	return &KeyDataSource{}
}

// KeyDataSource defines the data source implementation.
type KeyDataSource struct {
	client *Client
}

// KeyDataSourceModel describes the data source data model.
type KeyDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Algorithm           types.String `tfsdk:"algorithm"`
	Curve               types.String `tfsdk:"curve"`
	KeySize             types.Int64  `tfsdk:"key_size"`
	ProviderID          types.Int64  `tfsdk:"provider_id"`
	IsCA                types.Bool   `tfsdk:"is_ca"`
	Description         types.String `tfsdk:"description"`
	PublicKey           types.String `tfsdk:"public_key"`
	Certificate         types.String `tfsdk:"certificate"`
	CreatedAt           types.String `tfsdk:"created_at"`
	CanUseInNodes       types.Bool   `tfsdk:"can_use_in_nodes"`
	EthereumAddress     types.String `tfsdk:"ethereum_address"`
	ExpiresAt           types.String `tfsdk:"expires_at"`
	Format              types.String `tfsdk:"format"`
	GeneratedLocally    types.Bool   `tfsdk:"generated_locally"`
	ImportedToProvider  types.Bool   `tfsdk:"imported_to_provider"`
	IsExportable        types.Bool   `tfsdk:"is_exportable"`
	KmsKeyId            types.String `tfsdk:"kms_key_id"`
	LastRotatedAt       types.String `tfsdk:"last_rotated_at"`
	Sha1Fingerprint     types.String `tfsdk:"sha1_fingerprint"`
	Sha256Fingerprint   types.String `tfsdk:"sha256_fingerprint"`
	SigningKeyID        types.Int64  `tfsdk:"signing_key_id"`
	Status              types.String `tfsdk:"status"`
	VaultPath           types.String `tfsdk:"vault_path"`
	WasExported         types.Bool   `tfsdk:"was_exported"`
}

// KeyResponse represents the API response for a key
type KeyResponse struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	Algorithm          string `json:"algorithm"`
	Curve              string `json:"curve,omitempty"`
	KeySize            int64  `json:"keySize,omitempty"`
	ProviderID         int64  `json:"providerId"`
	IsCA               bool   `json:"isCA"`
	Description        string `json:"description,omitempty"`
	PublicKey          string `json:"publicKey,omitempty"`
	Certificate        string `json:"certificate,omitempty"`
	CreatedAt          string `json:"createdAt,omitempty"`
	CanUseInNodes      bool   `json:"canUseInNodes"`
	EthereumAddress    string `json:"ethereumAddress,omitempty"`
	ExpiresAt          string `json:"expiresAt,omitempty"`
	Format             string `json:"format,omitempty"`
	GeneratedLocally   bool   `json:"generatedLocally"`
	ImportedToProvider bool   `json:"importedToProvider"`
	IsExportable       bool   `json:"isExportable"`
	KmsKeyId           string `json:"kmsKeyId,omitempty"`
	LastRotatedAt      string `json:"lastRotatedAt,omitempty"`
	Sha1Fingerprint    string `json:"sha1Fingerprint,omitempty"`
	Sha256Fingerprint  string `json:"sha256Fingerprint,omitempty"`
	SigningKeyID       int64  `json:"signingKeyID,omitempty"`
	Status             string `json:"status,omitempty"`
	VaultPath          string `json:"vaultPath,omitempty"`
	WasExported        bool   `json:"wasExported"`
}

func (d *KeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key"
}

func (d *KeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details about a specific cryptographic key in Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the key to look up.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the key.",
			},
			"algorithm": schema.StringAttribute{
				Computed:    true,
				Description: "The key algorithm (RSA, EC, ED25519).",
			},
			"curve": schema.StringAttribute{
				Computed:    true,
				Description: "The elliptic curve for EC keys (P-256, P-384, P-521, secp256k1).",
			},
			"key_size": schema.Int64Attribute{
				Computed:    true,
				Description: "The key size in bits for RSA keys.",
			},
			"provider_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the key provider.",
			},
			"is_ca": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this key is a Certificate Authority key.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "A description of the key.",
			},
			"public_key": schema.StringAttribute{
				Computed:    true,
				Description: "The public key in PEM format.",
			},
			"certificate": schema.StringAttribute{
				Computed:    true,
				Description: "The certificate in PEM format.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the key was created.",
			},
			"can_use_in_nodes": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the key can be used in Fabric nodes (exportable + was_exported).",
			},
			"ethereum_address": schema.StringAttribute{
				Computed:    true,
				Description: "The Ethereum address derived from the key (for secp256k1 keys).",
			},
			"expires_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the key certificate expires.",
			},
			"format": schema.StringAttribute{
				Computed:    true,
				Description: "The key format.",
			},
			"generated_locally": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the key was generated locally.",
			},
			"imported_to_provider": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the key was imported to the provider.",
			},
			"is_exportable": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the private key can be exported/accessed locally.",
			},
			"kms_key_id": schema.StringAttribute{
				Computed:    true,
				Description: "The AWS KMS key ID/ARN (for AWS KMS provider keys).",
			},
			"last_rotated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the key was last rotated.",
			},
			"sha1_fingerprint": schema.StringAttribute{
				Computed:    true,
				Description: "The SHA1 fingerprint of the public key.",
			},
			"sha256_fingerprint": schema.StringAttribute{
				Computed:    true,
				Description: "The SHA256 fingerprint of the public key.",
			},
			"signing_key_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the signing key if applicable.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the key.",
			},
			"vault_path": schema.StringAttribute{
				Computed:    true,
				Description: "The Vault secret path (for Vault provider keys).",
			},
			"was_exported": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the private key was exported (or generated locally).",
			},
		},
	}
}

func (d *KeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *KeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KeyDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that ID is provided
	keyID := data.ID.ValueString()
	if keyID == "" {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"The 'id' attribute must be specified to look up a key.",
		)
		return
	}

	// Fetch the key from the API
	body, err := d.client.DoRequest("GET", fmt.Sprintf("/keys/%s", keyID), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read key, got error: %s", err))
		return
	}

	var keyResp KeyResponse
	if err := json.Unmarshal(body, &keyResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse key response, got error: %s", err))
		return
	}

	// Validate that we got a valid key
	if keyResp.ID == 0 {
		resp.Diagnostics.AddError(
			"Key Not Found",
			fmt.Sprintf("No key found with ID: %s", keyID),
		)
		return
	}

	// Set all fields from the response
	data.ID = types.StringValue(fmt.Sprintf("%d", keyResp.ID))
	data.Name = types.StringValue(keyResp.Name)
	data.Algorithm = types.StringValue(keyResp.Algorithm)
	data.ProviderID = types.Int64Value(keyResp.ProviderID)
	data.IsCA = types.BoolValue(keyResp.IsCA)

	// Set boolean flags
	data.CanUseInNodes = types.BoolValue(keyResp.CanUseInNodes)
	data.GeneratedLocally = types.BoolValue(keyResp.GeneratedLocally)
	data.ImportedToProvider = types.BoolValue(keyResp.ImportedToProvider)
	data.IsExportable = types.BoolValue(keyResp.IsExportable)
	data.WasExported = types.BoolValue(keyResp.WasExported)

	// Set optional string fields
	if keyResp.Curve != "" {
		data.Curve = types.StringValue(keyResp.Curve)
	}
	if keyResp.KeySize > 0 {
		data.KeySize = types.Int64Value(keyResp.KeySize)
	}
	if keyResp.Description != "" {
		data.Description = types.StringValue(keyResp.Description)
	}
	if keyResp.PublicKey != "" {
		data.PublicKey = types.StringValue(keyResp.PublicKey)
	}
	if keyResp.Certificate != "" {
		data.Certificate = types.StringValue(keyResp.Certificate)
	}
	if keyResp.CreatedAt != "" {
		data.CreatedAt = types.StringValue(keyResp.CreatedAt)
	}
	if keyResp.EthereumAddress != "" {
		data.EthereumAddress = types.StringValue(keyResp.EthereumAddress)
	}
	if keyResp.ExpiresAt != "" {
		data.ExpiresAt = types.StringValue(keyResp.ExpiresAt)
	}
	if keyResp.Format != "" {
		data.Format = types.StringValue(keyResp.Format)
	}
	if keyResp.KmsKeyId != "" {
		data.KmsKeyId = types.StringValue(keyResp.KmsKeyId)
	}
	if keyResp.LastRotatedAt != "" {
		data.LastRotatedAt = types.StringValue(keyResp.LastRotatedAt)
	}
	if keyResp.Sha1Fingerprint != "" {
		data.Sha1Fingerprint = types.StringValue(keyResp.Sha1Fingerprint)
	}
	if keyResp.Sha256Fingerprint != "" {
		data.Sha256Fingerprint = types.StringValue(keyResp.Sha256Fingerprint)
	}
	if keyResp.SigningKeyID > 0 {
		data.SigningKeyID = types.Int64Value(keyResp.SigningKeyID)
	}
	if keyResp.Status != "" {
		data.Status = types.StringValue(keyResp.Status)
	}
	if keyResp.VaultPath != "" {
		data.VaultPath = types.StringValue(keyResp.VaultPath)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
