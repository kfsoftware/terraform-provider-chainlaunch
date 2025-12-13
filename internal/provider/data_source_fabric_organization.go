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
var _ datasource.DataSource = &OrganizationDataSource{}

func NewOrganizationDataSource() datasource.DataSource {
	return &OrganizationDataSource{}
}

// OrganizationDataSource defines the data source implementation.
type OrganizationDataSource struct {
	client *Client
}

// OrganizationDataSourceModel describes the data source data model.
type OrganizationDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	MSPID           types.String `tfsdk:"msp_id"`
	Description     types.String `tfsdk:"description"`
	CACertValidFor  types.String `tfsdk:"ca_cert_valid_for"`
	CertValidFor    types.String `tfsdk:"cert_valid_for"`
	SignCAKeyId     types.Int64  `tfsdk:"sign_ca_key_id"`
	TlsCAKeyId      types.Int64  `tfsdk:"tls_ca_key_id"`
	AdminTlsKeyId   types.Int64  `tfsdk:"admin_tls_key_id"`
	AdminSignKeyId  types.Int64  `tfsdk:"admin_sign_key_id"`
	ClientSignKeyId types.Int64  `tfsdk:"client_sign_key_id"`
	SignPublicKey   types.String `tfsdk:"sign_public_key"`
	SignCertificate types.String `tfsdk:"sign_certificate"`
	TlsPublicKey    types.String `tfsdk:"tls_public_key"`
	TlsCertificate  types.String `tfsdk:"tls_certificate"`
	ProviderName    types.String `tfsdk:"provider_name"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (d *OrganizationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_organization"
}

func (d *OrganizationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details about a specific Hyperledger Fabric organization from Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The unique identifier of the organization. Either id or msp_id must be specified.",
			},
			"msp_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The MSP ID of the organization. Either id or msp_id must be specified.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "A description of the organization.",
			},
			"ca_cert_valid_for": schema.StringAttribute{
				Computed:    true,
				Description: "Certificate validity configuration for CA certificates in Go duration format.",
			},
			"cert_valid_for": schema.StringAttribute{
				Computed:    true,
				Description: "Certificate validity for admin/client/peer certificates in Go duration format.",
			},
			"sign_ca_key_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the signing CA key for this organization.",
			},
			"tls_ca_key_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the TLS CA key for this organization.",
			},
			"admin_tls_key_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the admin TLS key created during import.",
			},
			"admin_sign_key_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the admin signing key created during import.",
			},
			"client_sign_key_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the client signing key created during import.",
			},
			"sign_public_key": schema.StringAttribute{
				Computed:    true,
				Description: "The signing CA public key in PEM format.",
			},
			"sign_certificate": schema.StringAttribute{
				Computed:    true,
				Description: "The signing CA certificate in PEM format.",
			},
			"tls_public_key": schema.StringAttribute{
				Computed:    true,
				Description: "The TLS CA public key in PEM format.",
			},
			"tls_certificate": schema.StringAttribute{
				Computed:    true,
				Description: "The TLS CA certificate in PEM format.",
			},
			"provider_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the key provider for this organization.",
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

func (d *OrganizationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either ID or MSP ID is provided
	hasID := !data.ID.IsNull() && data.ID.ValueString() != ""
	hasMSPID := !data.MSPID.IsNull() && data.MSPID.ValueString() != ""

	if !hasID && !hasMSPID {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'msp_id' must be specified to look up an organization.",
		)
		return
	}

	var org Organization

	if hasID {
		// Lookup by ID - direct GET
		body, err := d.client.DoRequest("GET", fmt.Sprintf("/organizations/%s", data.ID.ValueString()), nil)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization, got error: %s", err))
			return
		}

		if err := json.Unmarshal(body, &org); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse organization response, got error: %s", err))
			return
		}
	} else {
		// Lookup by MSP ID using query parameter (API-side filtering)
		body, err := d.client.DoRequest("GET", fmt.Sprintf("/organizations?mspId=%s", data.MSPID.ValueString()), nil)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to search for organization by msp_id, got error: %s", err))
			return
		}

		// Parse paginated response
		var paginatedResp struct {
			Items []Organization `json:"items"`
			Count int            `json:"count"`
		}

		if err := json.Unmarshal(body, &paginatedResp); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse paginated organizations response, got error: %s", err))
			return
		}

		// Find the organization with matching MSP ID
		found := false
		for _, item := range paginatedResp.Items {
			if item.MSPID == data.MSPID.ValueString() {
				org = item
				found = true
				break
			}
		}

		if !found {
			resp.Diagnostics.AddError(
				"Organization Not Found",
				fmt.Sprintf("No organization found with msp_id: %s", data.MSPID.ValueString()),
			)
			return
		}
	}

	// Set all fields from the organization
	data.ID = types.StringValue(fmt.Sprintf("%d", org.ID))
	data.MSPID = types.StringValue(org.MSPID)
	if org.Description != "" {
		data.Description = types.StringValue(org.Description)
	}
	if org.CACertValidFor != "" {
		data.CACertValidFor = types.StringValue(org.CACertValidFor)
	}
	if org.CertValidFor != "" {
		data.CertValidFor = types.StringValue(org.CertValidFor)
	}
	if org.SignCAKeyId != 0 {
		data.SignCAKeyId = types.Int64Value(org.SignCAKeyId)
	}
	if org.TlsCAKeyId != 0 {
		data.TlsCAKeyId = types.Int64Value(org.TlsCAKeyId)
	}
	if org.AdminTlsKeyId != 0 {
		data.AdminTlsKeyId = types.Int64Value(org.AdminTlsKeyId)
	}
	if org.AdminSignKeyId != 0 {
		data.AdminSignKeyId = types.Int64Value(org.AdminSignKeyId)
	}
	if org.ClientSignKeyId != 0 {
		data.ClientSignKeyId = types.Int64Value(org.ClientSignKeyId)
	}
	if org.SignPublicKey != "" {
		data.SignPublicKey = types.StringValue(org.SignPublicKey)
	}
	if org.SignCertificate != "" {
		data.SignCertificate = types.StringValue(org.SignCertificate)
	}
	if org.TlsPublicKey != "" {
		data.TlsPublicKey = types.StringValue(org.TlsPublicKey)
	}
	if org.TlsCertificate != "" {
		data.TlsCertificate = types.StringValue(org.TlsCertificate)
	}
	if org.ProviderName != "" {
		data.ProviderName = types.StringValue(org.ProviderName)
	}
	if org.CreatedAt != "" {
		data.CreatedAt = types.StringValue(org.CreatedAt)
	}
	if org.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(org.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
