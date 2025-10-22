package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ExternalFabricOrganizationsDataSource{}

func NewExternalFabricOrganizationsDataSource() datasource.DataSource {
	return &ExternalFabricOrganizationsDataSource{}
}

type ExternalFabricOrganizationsDataSource struct {
	client *Client
}

type ExternalFabricOrganizationsDataSourceModel struct {
	Organizations []ExternalFabricOrganizationModel `tfsdk:"organizations"`
}

type ExternalFabricOrganizationModel struct {
	ID              types.Int64  `tfsdk:"id"`
	MSPID           types.String `tfsdk:"msp_id"`
	SignCertificate types.String `tfsdk:"sign_certificate"`
	TLSCertificate  types.String `tfsdk:"tls_certificate"`
}

func (d *ExternalFabricOrganizationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_fabric_organizations"
}

func (d *ExternalFabricOrganizationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "List all external Fabric organizations that have been synced from remote Chainlaunch instances.",
		MarkdownDescription: "List all external Fabric organizations that have been synced from remote Chainlaunch instances.",

		Attributes: map[string]schema.Attribute{
			"organizations": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "List of external Fabric organizations",
				MarkdownDescription: "List of external Fabric organizations",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:            true,
							Description:         "External organization ID",
							MarkdownDescription: "External organization ID",
						},
						"msp_id": schema.StringAttribute{
							Computed:            true,
							Description:         "MSP ID of the organization",
							MarkdownDescription: "MSP ID of the organization",
						},
						"sign_certificate": schema.StringAttribute{
							Computed:            true,
							Description:         "Signing certificate",
							MarkdownDescription: "Signing certificate",
						},
						"tls_certificate": schema.StringAttribute{
							Computed:            true,
							Description:         "TLS certificate",
							MarkdownDescription: "TLS certificate",
						},
					},
				},
			},
		},
	}
}

func (d *ExternalFabricOrganizationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ExternalFabricOrganizationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExternalFabricOrganizationsDataSourceModel

	body, err := d.client.DoRequest("GET", "/external-nodes/fabric-organizations", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read external Fabric organizations, got error: %s", err))
		return
	}

	var orgs []map[string]interface{}
	if err := json.Unmarshal(body, &orgs); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse organizations: %s", err))
		return
	}

	data.Organizations = make([]ExternalFabricOrganizationModel, 0, len(orgs))
	for _, org := range orgs {
		orgModel := ExternalFabricOrganizationModel{}

		if id, ok := org["id"].(float64); ok {
			orgModel.ID = types.Int64Value(int64(id))
		}

		if mspID, ok := org["mspId"].(string); ok {
			orgModel.MSPID = types.StringValue(mspID)
		}

		if signCert, ok := org["signCertificate"].(string); ok {
			orgModel.SignCertificate = types.StringValue(signCert)
		}

		if tlsCert, ok := org["tlsCertificate"].(string); ok {
			orgModel.TLSCertificate = types.StringValue(tlsCert)
		}

		data.Organizations = append(data.Organizations, orgModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
