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
var _ resource.Resource = &FabricOrganizationImportResource{}
var _ resource.ResourceWithImportState = &FabricOrganizationImportResource{}

func NewFabricOrganizationImportResource() resource.Resource {
	return &FabricOrganizationImportResource{}
}

// FabricOrganizationImportResource defines the resource implementation.
type FabricOrganizationImportResource struct {
	client *Client
}

// FabricOrganizationImportResourceModel describes the resource data model.
type FabricOrganizationImportResourceModel struct {
	ID              types.String        `tfsdk:"id"`
	MSPID           types.String        `tfsdk:"msp_id"`
	Name            types.String        `tfsdk:"name"`
	Description     types.String        `tfsdk:"description"`
	ProviderID      types.Int64         `tfsdk:"provider_id"`
	SourceType      types.String        `tfsdk:"source_type"`
	RawImport       []RawImportModel    `tfsdk:"raw_import"`
	VaultImport     []VaultImportModel  `tfsdk:"vault_import"`
	AWSKmsImport    []AWSKmsImportModel `tfsdk:"aws_kms_import"`
	AdminSignKeyID  types.Int64         `tfsdk:"admin_sign_key_id"`
	AdminTlsKeyID   types.Int64         `tfsdk:"admin_tls_key_id"`
	ClientSignKeyID types.Int64         `tfsdk:"client_sign_key_id"`
	SignCertificate types.String        `tfsdk:"sign_certificate"`
	SignPublicKey   types.String        `tfsdk:"sign_public_key"`
	TLSCertificate  types.String        `tfsdk:"tls_certificate"`
	TLSPublicKey    types.String        `tfsdk:"tls_public_key"`
	CreatedAt       types.String        `tfsdk:"created_at"`
	UpdatedAt       types.String        `tfsdk:"updated_at"`
}

type RawImportModel struct {
	SignCaCert       types.String `tfsdk:"sign_ca_cert"`
	SignCaPrivateKey types.String `tfsdk:"sign_ca_private_key"`
	TLSCaCert        types.String `tfsdk:"tls_ca_cert"`
	TLSCaPrivateKey  types.String `tfsdk:"tls_ca_private_key"`
}

type VaultImportModel struct {
	SignCaPath types.String `tfsdk:"sign_ca_path"`
	TlsCaPath  types.String `tfsdk:"tls_ca_path"`
}

type AWSKmsImportModel struct {
	SignCaCert  types.String `tfsdk:"sign_ca_cert"`
	SignCaKeyId types.String `tfsdk:"sign_ca_key_id"`
	TLSCaCert   types.String `tfsdk:"tls_ca_cert"`
	TLSCaKeyId  types.String `tfsdk:"tls_ca_key_id"`
}

func (r *FabricOrganizationImportResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_organization_import"
}

func (r *FabricOrganizationImportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Imports an existing Hyperledger Fabric organization from various certificate sources.

This resource supports importing organizations from:
- **Raw Certificates**: Direct PEM-encoded certificates and keys
- **HashiCorp Vault**: Certificates and keys stored in Vault paths
- **AWS KMS**: Certificates stored in ChainLaunch with keys in AWS KMS

## Strict Validation Rules

- **provider_id** must match the specified key provider (Database, Vault, or AWS KMS)
- **source_type** determines which import configuration block is required and must be explicitly specified
- For **raw** imports: You must provide PEM-encoded certificates directly
- For **vault** imports: Specify paths to existing Vault secrets containing certificates
- For **aws_kms** imports: Provide certificates and AWS KMS key IDs/ARNs

## Source Type Matching

Each source type requires the provider to support that backend:
- **raw**: Works with any provider (Database, Vault, AWS KMS)
- **vault**: Requires provider_id pointing to a Vault key provider
- **aws_kms**: Requires provider_id pointing to an AWS KMS key provider
		`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier of the imported organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"msp_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The MSP ID of the organization. Must be unique and match your Fabric network configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the organization.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "A description of the organization.",
			},
			"provider_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The ID of the key provider (Database, Vault, or AWS KMS) where certificates will be stored. This must match the source_type. For raw imports with Database provider, use the default Database provider ID.",
			},
			"source_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The source type for importing certificates: `raw` (PEM-encoded), `vault` (HashiCorp Vault), or `aws_kms` (AWS KMS). This must match your provider_id configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"raw_import": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Raw certificate and key data (required when source_type = 'raw'). Provide PEM-encoded certificates and keys directly.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"sign_ca_cert": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "PEM-encoded signing CA certificate.",
						},
						"sign_ca_private_key": schema.StringAttribute{
							Optional:            true,
							Sensitive:           true,
							MarkdownDescription: "PEM-encoded signing CA private key (optional, allows generating new identities).",
						},
						"tls_ca_cert": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "PEM-encoded TLS CA certificate.",
						},
						"tls_ca_private_key": schema.StringAttribute{
							Optional:            true,
							Sensitive:           true,
							MarkdownDescription: "PEM-encoded TLS CA private key (optional).",
						},
					},
				},
			},
			"vault_import": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Vault paths for retrieving certificates (required when source_type = 'vault'). Specify paths to existing Vault secrets containing certificates and keys.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"sign_ca_path": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Vault path containing the signing CA certificate and key (e.g., 'secret/data/fabric/sign-ca').",
						},
						"tls_ca_path": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "Vault path containing the TLS CA certificate and key (e.g., 'secret/data/fabric/tls-ca').",
						},
					},
				},
			},
			"aws_kms_import": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "AWS KMS configuration (required when source_type = 'aws_kms'). Certificates are stored in ChainLaunch, but private keys are tracked in AWS KMS.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"sign_ca_cert": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "PEM-encoded signing CA certificate (stored in ChainLaunch).",
						},
						"sign_ca_key_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "AWS KMS key ID or ARN for the signing CA key.",
						},
						"tls_ca_cert": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "PEM-encoded TLS CA certificate (stored in ChainLaunch).",
						},
						"tls_ca_key_id": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "AWS KMS key ID or ARN for the TLS CA key.",
						},
					},
				},
			},
			"admin_sign_key_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the admin signing key created during import.",
			},
			"admin_tls_key_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the admin TLS key created during import.",
			},
			"client_sign_key_id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The ID of the client signing key created during import.",
			},
			"sign_certificate": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The signing CA certificate in PEM format.",
			},
			"sign_public_key": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The signing CA public key in PEM format.",
			},
			"tls_certificate": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The TLS CA certificate in PEM format.",
			},
			"tls_public_key": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The TLS CA public key in PEM format.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the organization was imported.",
			},
			"updated_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The timestamp when the organization was last updated.",
			},
		},
	}
}

func (r *FabricOrganizationImportResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FabricOrganizationImportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricOrganizationImportResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate strict requirements
	sourceType := data.SourceType.ValueString()
	providerID := data.ProviderID.ValueInt64()

	// Validate that source_type is specified
	if sourceType == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("source_type"),
			"Missing Source Type",
			"source_type is required and must be one of: raw, vault, aws_kms",
		)
		return
	}

	// Validate that only the appropriate import config is provided for the source type
	hasRawImport := len(data.RawImport) > 0
	hasVaultImport := len(data.VaultImport) > 0
	hasAWSKmsImport := len(data.AWSKmsImport) > 0

	importConfigCount := 0
	if hasRawImport {
		importConfigCount++
	}
	if hasVaultImport {
		importConfigCount++
	}
	if hasAWSKmsImport {
		importConfigCount++
	}

	if importConfigCount == 0 {
		resp.Diagnostics.AddError(
			"Missing Import Configuration",
			fmt.Sprintf("Must provide exactly one import configuration block matching source_type='%s'. Available: raw_import, vault_import, or aws_kms_import", sourceType),
		)
		return
	}

	if importConfigCount > 1 {
		resp.Diagnostics.AddError(
			"Multiple Import Configurations",
			fmt.Sprintf("Only one import configuration block is allowed for source_type='%s'. Remove extra blocks.", sourceType),
		)
		return
	}

	// Validate source_type matches provided import configuration
	switch sourceType {
	case "raw":
		if !hasRawImport {
			resp.Diagnostics.AddAttributeError(
				path.Root("raw_import"),
				"Missing Raw Import Configuration",
				"raw_import block is required when source_type='raw'",
			)
			return
		}

	case "vault":
		if !hasVaultImport {
			resp.Diagnostics.AddAttributeError(
				path.Root("vault_import"),
				"Missing Vault Import Configuration",
				"vault_import block is required when source_type='vault'. Ensure provider_id points to a Vault key provider.",
			)
			return
		}

	case "aws_kms":
		if !hasAWSKmsImport {
			resp.Diagnostics.AddAttributeError(
				path.Root("aws_kms_import"),
				"Missing AWS KMS Import Configuration",
				"aws_kms_import block is required when source_type='aws_kms'. Ensure provider_id points to an AWS KMS key provider.",
			)
			return
		}

	default:
		resp.Diagnostics.AddAttributeError(
			path.Root("source_type"),
			"Invalid Source Type",
			fmt.Sprintf("source_type must be 'raw', 'vault', or 'aws_kms', got '%s'", sourceType),
		)
		return
	}

	// Validate provider_id is provided
	if data.ProviderID.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("provider_id"),
			"Missing Provider ID",
			"provider_id is required and must match the certificate source (Database for raw, Vault for vault imports, AWS KMS for aws_kms)",
		)
		return
	}

	// Build import request
	importReq := ImportOrganizationRequest{
		MSPID:       data.MSPID.ValueString(),
		Name:        data.Name.ValueString(),
		ProviderID:  providerID,
		SourceType:  sourceType,
		Description: data.Description.ValueString(),
	}

	// Populate the appropriate import config based on source type
	switch sourceType {
	case "raw":
		if len(data.RawImport) > 0 {
			raw := data.RawImport[0]
			importReq.RawImport = &RawImportData{
				SignCaCert:       raw.SignCaCert.ValueString(),
				SignCaPrivateKey: raw.SignCaPrivateKey.ValueString(),
				TLSCaCert:        raw.TLSCaCert.ValueString(),
				TLSCaPrivateKey:  raw.TLSCaPrivateKey.ValueString(),
			}

			// Validate required fields
			if importReq.RawImport.SignCaCert == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("raw_import").AtListIndex(0).AtName("sign_ca_cert"),
					"Missing Certificate",
					"sign_ca_cert is required for raw imports",
				)
				return
			}
			if importReq.RawImport.TLSCaCert == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("raw_import").AtListIndex(0).AtName("tls_ca_cert"),
					"Missing Certificate",
					"tls_ca_cert is required for raw imports",
				)
				return
			}
		}

	case "vault":
		if len(data.VaultImport) > 0 {
			vault := data.VaultImport[0]
			importReq.VaultImport = &VaultImportData{
				SignCaPath: vault.SignCaPath.ValueString(),
				TlsCaPath:  vault.TlsCaPath.ValueString(),
			}

			// Validate required paths
			if importReq.VaultImport.SignCaPath == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("vault_import").AtListIndex(0).AtName("sign_ca_path"),
					"Missing Vault Path",
					"sign_ca_path is required for vault imports (e.g., 'secret/data/fabric/sign-ca')",
				)
				return
			}
			if importReq.VaultImport.TlsCaPath == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("vault_import").AtListIndex(0).AtName("tls_ca_path"),
					"Missing Vault Path",
					"tls_ca_path is required for vault imports (e.g., 'secret/data/fabric/tls-ca')",
				)
				return
			}
		}

	case "aws_kms":
		if len(data.AWSKmsImport) > 0 {
			awskms := data.AWSKmsImport[0]
			importReq.AWSKmsImport = &AWSKMSImportData{
				SignCaCert:  awskms.SignCaCert.ValueString(),
				SignCaKeyId: awskms.SignCaKeyId.ValueString(),
				TLSCaCert:   awskms.TLSCaCert.ValueString(),
				TLSCaKeyId:  awskms.TLSCaKeyId.ValueString(),
			}

			// Validate required fields
			if importReq.AWSKmsImport.SignCaCert == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("aws_kms_import").AtListIndex(0).AtName("sign_ca_cert"),
					"Missing Certificate",
					"sign_ca_cert is required for AWS KMS imports",
				)
				return
			}
			if importReq.AWSKmsImport.SignCaKeyId == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("aws_kms_import").AtListIndex(0).AtName("sign_ca_key_id"),
					"Missing AWS KMS Key ID",
					"sign_ca_key_id is required for AWS KMS imports (e.g., 'arn:aws:kms:us-east-1:123456789012:key/...' or key ID)",
				)
				return
			}
			if importReq.AWSKmsImport.TLSCaCert == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("aws_kms_import").AtListIndex(0).AtName("tls_ca_cert"),
					"Missing Certificate",
					"tls_ca_cert is required for AWS KMS imports",
				)
				return
			}
			if importReq.AWSKmsImport.TLSCaKeyId == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("aws_kms_import").AtListIndex(0).AtName("tls_ca_key_id"),
					"Missing AWS KMS Key ID",
					"tls_ca_key_id is required for AWS KMS imports",
				)
				return
			}
		}
	}

	body, err := r.client.DoRequest("POST", "/organizations/import", importReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to import organization, got error: %s", err))
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
			fmt.Sprintf("Organization import failed: API returned invalid or empty organization ID. Response body: %s", string(body)),
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
	if org.Description != "" {
		data.Description = types.StringValue(org.Description)
	} else {
		data.Description = types.StringNull()
	}
	if org.AdminSignKeyId != 0 {
		data.AdminSignKeyID = types.Int64Value(org.AdminSignKeyId)
	} else {
		data.AdminSignKeyID = types.Int64Null()
	}
	if org.AdminTlsKeyId != 0 {
		data.AdminTlsKeyID = types.Int64Value(org.AdminTlsKeyId)
	} else {
		data.AdminTlsKeyID = types.Int64Null()
	}
	if org.ClientSignKeyId != 0 {
		data.ClientSignKeyID = types.Int64Value(org.ClientSignKeyId)
	} else {
		data.ClientSignKeyID = types.Int64Null()
	}
	if org.SignCertificate != "" {
		data.SignCertificate = types.StringValue(org.SignCertificate)
	} else {
		data.SignCertificate = types.StringNull()
	}
	if org.SignPublicKey != "" {
		data.SignPublicKey = types.StringValue(org.SignPublicKey)
	} else {
		data.SignPublicKey = types.StringNull()
	}
	if org.TlsCertificate != "" {
		data.TLSCertificate = types.StringValue(org.TlsCertificate)
	} else {
		data.TLSCertificate = types.StringNull()
	}
	if org.TlsPublicKey != "" {
		data.TLSPublicKey = types.StringValue(org.TlsPublicKey)
	} else {
		data.TLSPublicKey = types.StringNull()
	}
	if org.CreatedAt != "" {
		data.CreatedAt = types.StringValue(org.CreatedAt)
	} else {
		data.CreatedAt = types.StringNull()
	}
	if org.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(org.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	// Only save to state if no errors occurred
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricOrganizationImportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricOrganizationImportResourceModel

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
	} else {
		data.Description = types.StringNull()
	}
	if org.AdminSignKeyId != 0 {
		data.AdminSignKeyID = types.Int64Value(org.AdminSignKeyId)
	} else {
		data.AdminSignKeyID = types.Int64Null()
	}
	if org.AdminTlsKeyId != 0 {
		data.AdminTlsKeyID = types.Int64Value(org.AdminTlsKeyId)
	} else {
		data.AdminTlsKeyID = types.Int64Null()
	}
	if org.ClientSignKeyId != 0 {
		data.ClientSignKeyID = types.Int64Value(org.ClientSignKeyId)
	} else {
		data.ClientSignKeyID = types.Int64Null()
	}
	if org.SignCertificate != "" {
		data.SignCertificate = types.StringValue(org.SignCertificate)
	} else {
		data.SignCertificate = types.StringNull()
	}
	if org.SignPublicKey != "" {
		data.SignPublicKey = types.StringValue(org.SignPublicKey)
	} else {
		data.SignPublicKey = types.StringNull()
	}
	if org.TlsCertificate != "" {
		data.TLSCertificate = types.StringValue(org.TlsCertificate)
	} else {
		data.TLSCertificate = types.StringNull()
	}
	if org.TlsPublicKey != "" {
		data.TLSPublicKey = types.StringValue(org.TlsPublicKey)
	} else {
		data.TLSPublicKey = types.StringNull()
	}
	if org.CreatedAt != "" {
		data.CreatedAt = types.StringValue(org.CreatedAt)
	} else {
		data.CreatedAt = types.StringNull()
	}
	if org.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(org.UpdatedAt)
	} else {
		data.UpdatedAt = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricOrganizationImportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FabricOrganizationImportResourceModel
	var state FabricOrganizationImportResourceModel

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

	// Preserve computed fields from state (fields that don't change on update)
	data.CreatedAt = state.CreatedAt
	data.AdminSignKeyID = state.AdminSignKeyID
	data.AdminTlsKeyID = state.AdminTlsKeyID
	data.ClientSignKeyID = state.ClientSignKeyID
	data.SignCertificate = state.SignCertificate
	data.SignPublicKey = state.SignPublicKey
	data.TLSCertificate = state.TLSCertificate
	data.TLSPublicKey = state.TLSPublicKey
	data.SourceType = state.SourceType
	data.RawImport = state.RawImport
	data.VaultImport = state.VaultImport
	data.AWSKmsImport = state.AWSKmsImport

	// For imported organizations, only description can be updated
	updateReq := map[string]interface{}{
		"description": data.Description.ValueString(),
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

	if org.Description != "" {
		data.Description = types.StringValue(org.Description)
	}
	if org.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(org.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricOrganizationImportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FabricOrganizationImportResourceModel

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

func (r *FabricOrganizationImportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
