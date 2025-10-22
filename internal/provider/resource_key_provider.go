package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ resource.Resource = &KeyProviderResource{}
var _ resource.ResourceWithImportState = &KeyProviderResource{}

func NewKeyProviderResource() resource.Resource {
	return &KeyProviderResource{}
}

type KeyProviderResource struct {
	client *Client
}

type KeyProviderResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	IsDefault    types.Bool   `tfsdk:"is_default"`
	AWSKMSConfig types.Object `tfsdk:"aws_kms_config"`
	VaultConfig  types.Object `tfsdk:"vault_config"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

// AWSKMSConfigModel describes AWS KMS configuration.
type AWSKMSConfigModel struct {
	Operation          types.String `tfsdk:"operation"`
	AWSRegion          types.String `tfsdk:"aws_region"`
	AWSAccessKeyID     types.String `tfsdk:"aws_access_key_id"`
	AWSSecretAccessKey types.String `tfsdk:"aws_secret_access_key"`
	AWSSessionToken    types.String `tfsdk:"aws_session_token"`
	AssumeRoleARN      types.String `tfsdk:"assume_role_arn"`
	ExternalID         types.String `tfsdk:"external_id"`
	EndpointURL        types.String `tfsdk:"endpoint_url"`
	KMSKeyAliasPrefix  types.String `tfsdk:"kms_key_alias_prefix"`
}

// VaultConfigModel describes Vault configuration.
type VaultConfigModel struct {
	Operation types.String `tfsdk:"operation"`
	// IMPORT mode fields
	Address   types.String `tfsdk:"address"`
	Token     types.String `tfsdk:"token"`
	Mount     types.String `tfsdk:"mount"`
	CACert    types.String `tfsdk:"ca_cert"`
	Namespace types.String `tfsdk:"namespace"`
	// CREATE mode fields
	Mode    types.String `tfsdk:"mode"`
	Network types.String `tfsdk:"network"`
	Port    types.Int64  `tfsdk:"port"`
	Version types.String `tfsdk:"version"` // Vault version for CREATE
	// PKI configuration
	PKIMount         types.String `tfsdk:"pki_mount"`
	KVMount          types.String `tfsdk:"kv_mount"`
	DefaultCertTTL   types.String `tfsdk:"default_cert_ttl"`
	MaxCertTTL       types.String `tfsdk:"max_cert_ttl"`
	DefaultCACertTTL types.String `tfsdk:"default_ca_cert_ttl"`
	MaxCACertTTL     types.String `tfsdk:"max_ca_cert_ttl"`
}

func (r *KeyProviderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key_provider"
}

func (r *KeyProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a key provider in Chainlaunch for managing cryptographic keys.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the key provider.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the key provider.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of key provider (AWS_KMS, VAULT, DATABASE, HSM).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_default": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether this is the default key provider.",
			},
			"aws_kms_config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "AWS KMS configuration. Required when type is AWS_KMS.",
				Attributes: map[string]schema.Attribute{
					"operation": schema.StringAttribute{
						Required:    true,
						Description: "Operation mode: IMPORT (use existing KMS keys) or CREATE (create new KMS keys).",
					},
					"aws_region": schema.StringAttribute{
						Required:    true,
						Description: "AWS region where KMS keys are located.",
					},
					"aws_access_key_id": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "AWS access key ID for authentication (optional if using IAM roles).",
					},
					"aws_secret_access_key": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "AWS secret access key for authentication.",
					},
					"aws_session_token": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "AWS session token for temporary credentials.",
					},
					"assume_role_arn": schema.StringAttribute{
						Optional:    true,
						Description: "IAM role ARN to assume for cross-account access.",
					},
					"external_id": schema.StringAttribute{
						Optional:    true,
						Description: "External ID for role assumption.",
					},
					"endpoint_url": schema.StringAttribute{
						Optional:    true,
						Description: "Custom endpoint URL (e.g., for LocalStack).",
					},
					"kms_key_alias_prefix": schema.StringAttribute{
						Optional:    true,
						Description: "Prefix for KMS key aliases (default: chainlaunch/).",
					},
				},
			},
			"vault_config": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Vault configuration. Required when type is VAULT.",
				Attributes: map[string]schema.Attribute{
					"operation": schema.StringAttribute{
						Required:    true,
						Description: "Operation mode: IMPORT (use existing Vault) or CREATE (create new Vault instance).",
					},
					// IMPORT mode fields
					"address": schema.StringAttribute{
						Optional:    true,
						Description: "Vault server address (required for IMPORT operation).",
					},
					"token": schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Vault authentication token (required for IMPORT operation).",
					},
					"mount": schema.StringAttribute{
						Optional:    true,
						Description: "Vault mount path for secrets (KV mount).",
					},
					"ca_cert": schema.StringAttribute{
						Optional:    true,
						Description: "CA certificate for Vault TLS verification (optional for IMPORT).",
					},
					"namespace": schema.StringAttribute{
						Optional:    true,
						Description: "Vault namespace (optional).",
					},
					// CREATE mode fields
					"mode": schema.StringAttribute{
						Optional:    true,
						Description: "Deployment mode for CREATE operation (e.g., 'dev', 'prod').",
					},
					"network": schema.StringAttribute{
						Optional:    true,
						Description: "Network mode for CREATE operation: 'host' or 'bridge'.",
					},
					"port": schema.Int64Attribute{
						Optional:    true,
						Description: "Port number for Vault server (used in CREATE mode).",
					},
					"version": schema.StringAttribute{
						Optional:    true,
						Description: "Vault version for CREATE operation (e.g., '1.15.6'). Required for CREATE mode.",
					},
					// PKI configuration
					"pki_mount": schema.StringAttribute{
						Optional:    true,
						Description: "PKI mount path (default: 'pki').",
					},
					"kv_mount": schema.StringAttribute{
						Optional:    true,
						Description: "KV secrets mount path (default: 'secret').",
					},
					"default_cert_ttl": schema.StringAttribute{
						Optional:    true,
						Description: "Default TTL for certificates (e.g., '8760h' for 1 year).",
					},
					"max_cert_ttl": schema.StringAttribute{
						Optional:    true,
						Description: "Maximum TTL for certificates (e.g., '87600h' for 10 years).",
					},
					"default_ca_cert_ttl": schema.StringAttribute{
						Optional:    true,
						Description: "Default TTL for CA certificates (e.g., '87600h' for 10 years).",
					},
					"max_ca_cert_ttl": schema.StringAttribute{
						Optional:    true,
						Description: "Maximum TTL for CA certificates (e.g., '175200h' for 20 years).",
					},
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the key provider was created.",
			},
		},
	}
}

func (r *KeyProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *KeyProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data KeyProviderResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the request payload
	createReq := map[string]interface{}{
		"name": data.Name.ValueString(),
		"type": data.Type.ValueString(),
	}

	isDefault := 0
	if !data.IsDefault.IsNull() && data.IsDefault.ValueBool() {
		isDefault = 1
	}
	createReq["isDefault"] = isDefault

	// Add configuration based on type
	config := make(map[string]interface{})

	if !data.AWSKMSConfig.IsNull() {
		var awsConfig AWSKMSConfigModel
		diags := data.AWSKMSConfig.As(ctx, &awsConfig, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		awsKmsConfig := map[string]interface{}{
			"operation": awsConfig.Operation.ValueString(),
			"awsRegion": awsConfig.AWSRegion.ValueString(),
		}

		if !awsConfig.AWSAccessKeyID.IsNull() {
			awsKmsConfig["awsAccessKeyId"] = awsConfig.AWSAccessKeyID.ValueString()
		}
		if !awsConfig.AWSSecretAccessKey.IsNull() {
			awsKmsConfig["awsSecretAccessKey"] = awsConfig.AWSSecretAccessKey.ValueString()
		}
		if !awsConfig.AWSSessionToken.IsNull() {
			awsKmsConfig["awsSessionToken"] = awsConfig.AWSSessionToken.ValueString()
		}
		if !awsConfig.AssumeRoleARN.IsNull() {
			awsKmsConfig["assumeRoleArn"] = awsConfig.AssumeRoleARN.ValueString()
		}
		if !awsConfig.ExternalID.IsNull() {
			awsKmsConfig["externalId"] = awsConfig.ExternalID.ValueString()
		}
		if !awsConfig.EndpointURL.IsNull() {
			awsKmsConfig["endpointUrl"] = awsConfig.EndpointURL.ValueString()
		}
		if !awsConfig.KMSKeyAliasPrefix.IsNull() {
			awsKmsConfig["kmsKeyAliasPrefix"] = awsConfig.KMSKeyAliasPrefix.ValueString()
		}

		config["awsKms"] = awsKmsConfig
	}

	if !data.VaultConfig.IsNull() {
		var vaultConfig VaultConfigModel
		diags := data.VaultConfig.As(ctx, &vaultConfig, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		vaultCfg := map[string]interface{}{
			"operation": vaultConfig.Operation.ValueString(),
		}

		// IMPORT mode fields
		if !vaultConfig.Address.IsNull() {
			vaultCfg["address"] = vaultConfig.Address.ValueString()
		}
		if !vaultConfig.Token.IsNull() {
			vaultCfg["token"] = vaultConfig.Token.ValueString()
		}
		if !vaultConfig.Mount.IsNull() {
			vaultCfg["mount"] = vaultConfig.Mount.ValueString()
		}
		if !vaultConfig.CACert.IsNull() {
			vaultCfg["caCert"] = vaultConfig.CACert.ValueString()
		}
		if !vaultConfig.Namespace.IsNull() {
			vaultCfg["namespace"] = vaultConfig.Namespace.ValueString()
		}

		// CREATE mode fields
		if !vaultConfig.Mode.IsNull() {
			vaultCfg["mode"] = vaultConfig.Mode.ValueString()
		}
		if !vaultConfig.Network.IsNull() {
			vaultCfg["network"] = vaultConfig.Network.ValueString()
		}
		if !vaultConfig.Port.IsNull() {
			vaultCfg["port"] = vaultConfig.Port.ValueInt64()
		}
		if !vaultConfig.Version.IsNull() {
			vaultCfg["version"] = vaultConfig.Version.ValueString()
		}

		// PKI configuration
		if !vaultConfig.PKIMount.IsNull() {
			vaultCfg["pkiMount"] = vaultConfig.PKIMount.ValueString()
		}
		if !vaultConfig.KVMount.IsNull() {
			vaultCfg["kvMount"] = vaultConfig.KVMount.ValueString()
		}
		if !vaultConfig.DefaultCertTTL.IsNull() {
			vaultCfg["defaultCertTTL"] = vaultConfig.DefaultCertTTL.ValueString()
		}
		if !vaultConfig.MaxCertTTL.IsNull() {
			vaultCfg["maxCertTTL"] = vaultConfig.MaxCertTTL.ValueString()
		}
		if !vaultConfig.DefaultCACertTTL.IsNull() {
			vaultCfg["defaultCACertTTL"] = vaultConfig.DefaultCACertTTL.ValueString()
		}
		if !vaultConfig.MaxCACertTTL.IsNull() {
			vaultCfg["maxCACertTTL"] = vaultConfig.MaxCACertTTL.ValueString()
		}

		config["vault"] = vaultCfg
	}

	createReq["config"] = config

	// Debug: Log the request payload
	requestJSON, _ := json.MarshalIndent(createReq, "", "  ")
	resp.Diagnostics.AddWarning(
		"DEBUG: Request Payload",
		fmt.Sprintf("Sending to POST /key-providers:\n%s", string(requestJSON)),
	)

	body, err := r.client.DoRequest("POST", "/key-providers", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create key provider, got error: %s", err))
		return
	}

	var providerResp struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Type      string `json:"type"`
		IsDefault int    `json:"isDefault"`
		CreatedAt string `json:"createdAt"`
	}

	if err := json.Unmarshal(body, &providerResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse key provider response, got error: %s\nResponse body: %s", err, string(body)))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", providerResp.ID))
	data.Name = types.StringValue(providerResp.Name)
	data.Type = types.StringValue(providerResp.Type)
	data.IsDefault = types.BoolValue(providerResp.IsDefault == 1)
	if providerResp.CreatedAt != "" {
		data.CreatedAt = types.StringValue(providerResp.CreatedAt)
	}

	// Wait for provider to be ready
	if data.Type.ValueString() == "VAULT" && !data.VaultConfig.IsNull() {
		// Vault status checking (only for CREATE operation)
		var vaultConfig VaultConfigModel
		diags := data.VaultConfig.As(ctx, &vaultConfig, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		if vaultConfig.Operation.ValueString() == "CREATE" {
			if err := r.waitForVaultReady(ctx, providerResp.ID); err != nil {
				resp.Diagnostics.AddWarning(
					"Provider Status Check",
					fmt.Sprintf("Vault provider created but status check failed: %s. The provider may not be fully ready yet.", err),
				)
			}
		}
	} else if data.Type.ValueString() == "AWS_KMS" && !data.AWSKMSConfig.IsNull() {
		// AWS KMS status checking
		if err := r.waitForAWSKMSReady(ctx, providerResp.ID); err != nil {
			resp.Diagnostics.AddWarning(
				"Provider Status Check",
				fmt.Sprintf("AWS KMS provider created but status check failed: %s. The provider may not be fully ready yet.", err),
			)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeyProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data KeyProviderResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.DoRequest("GET", fmt.Sprintf("/key-providers/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read key provider, got error: %s", err))
		return
	}

	var providerResp struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Type      string `json:"type"`
		IsDefault int    `json:"isDefault"`
		CreatedAt string `json:"createdAt"`
	}

	if err := json.Unmarshal(body, &providerResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse key provider response, got error: %s", err))
		return
	}

	data.Name = types.StringValue(providerResp.Name)
	data.Type = types.StringValue(providerResp.Type)
	data.IsDefault = types.BoolValue(providerResp.IsDefault == 1)
	if providerResp.CreatedAt != "" {
		data.CreatedAt = types.StringValue(providerResp.CreatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeyProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data KeyProviderResourceModel
	var state KeyProviderResourceModel

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

	// For now, we'll just read back the resource since updates might not be fully supported
	resp.Diagnostics.AddWarning(
		"Update Not Fully Implemented",
		"Key provider updates may have limited support. Most changes require resource replacement.",
	)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *KeyProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data KeyProviderResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DoRequest("DELETE", fmt.Sprintf("/key-providers/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete key provider, got error: %s", err))
		return
	}
}

func (r *KeyProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper function to get AWS KMS config attribute types

// Helper function to get Vault config attribute types

// waitForVaultReady polls the Vault status endpoint until the vault is ready or timeout
func (r *KeyProviderResource) waitForVaultReady(ctx context.Context, providerID int64) error {
	maxAttempts := 30 // 30 attempts
	delaySeconds := 2 // 2 seconds between attempts

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for Vault to be ready")
		default:
		}

		// Call the status endpoint
		body, err := r.client.DoRequest("GET", fmt.Sprintf("/key-providers/%d/vault/status", providerID), nil)
		if err != nil {
			// If it's just not ready yet, continue waiting
			if attempt < maxAttempts {
				time.Sleep(time.Duration(delaySeconds) * time.Second)
				continue
			}
			return fmt.Errorf("failed to get vault status after %d attempts: %s", maxAttempts, err)
		}

		var statusResp struct {
			VaultReachable   bool   `json:"vault_reachable"`
			VaultInitialized bool   `json:"vault_initialized"`
			Sealed           bool   `json:"sealed"`
			ContainerRunning bool   `json:"container_running"`
			VaultStatus      string `json:"vault_status"`
		}

		if err := json.Unmarshal(body, &statusResp); err != nil {
			return fmt.Errorf("failed to parse vault status response: %s", err)
		}

		// Debug logging
		statusJSON, _ := json.MarshalIndent(statusResp, "", "  ")
		fmt.Printf("[DEBUG] Vault status (attempt %d/%d):\n%s\n", attempt, maxAttempts, string(statusJSON))

		// Check if vault is ready
		if statusResp.VaultReachable && statusResp.VaultInitialized && !statusResp.Sealed && statusResp.ContainerRunning {
			return nil // Vault is ready!
		}

		// Not ready yet, wait and try again
		if attempt < maxAttempts {
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}
	}

	return fmt.Errorf("vault did not become ready after %d attempts (%d seconds)", maxAttempts, maxAttempts*delaySeconds)
}

// waitForAWSKMSReady polls the AWS KMS status endpoint until KMS is ready or timeout
func (r *KeyProviderResource) waitForAWSKMSReady(ctx context.Context, providerID int64) error {
	maxAttempts := 10 // 10 attempts
	delaySeconds := 2 // 2 seconds between attempts

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for AWS KMS to be ready")
		default:
		}

		// Call the status endpoint
		body, err := r.client.DoRequest("GET", fmt.Sprintf("/key-providers/%d/awskms/status", providerID), nil)
		if err != nil {
			// If it's just not ready yet, continue waiting
			if attempt < maxAttempts {
				time.Sleep(time.Duration(delaySeconds) * time.Second)
				continue
			}
			return fmt.Errorf("failed to get AWS KMS status after %d attempts: %s", maxAttempts, err)
		}

		var statusResp struct {
			KMSReachable    bool   `json:"kms_reachable"`
			KMSStatus       string `json:"kms_status"`
			HasCredentials  bool   `json:"has_credentials"`
			ConnectionError string `json:"connection_error"`
		}

		if err := json.Unmarshal(body, &statusResp); err != nil {
			return fmt.Errorf("failed to parse AWS KMS status response: %s", err)
		}

		// Check if KMS is ready
		if statusResp.KMSReachable && statusResp.HasCredentials && statusResp.KMSStatus == "available" {
			return nil // AWS KMS is ready!
		}

		// If there's a connection error, report it
		if statusResp.ConnectionError != "" && attempt == maxAttempts {
			return fmt.Errorf("AWS KMS connection error: %s", statusResp.ConnectionError)
		}

		// Not ready yet, wait and try again
		if attempt < maxAttempts {
			time.Sleep(time.Duration(delaySeconds) * time.Second)
		}
	}

	return fmt.Errorf("AWS KMS did not become ready after %d attempts (%d seconds)", maxAttempts, maxAttempts*delaySeconds)
}
