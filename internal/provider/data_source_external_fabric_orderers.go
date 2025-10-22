package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ExternalFabricOrderersDataSource{}

func NewExternalFabricOrderersDataSource() datasource.DataSource {
	return &ExternalFabricOrderersDataSource{}
}

type ExternalFabricOrderersDataSource struct {
	client *Client
}

type ExternalFabricOrderersDataSourceModel struct {
	Orderers []ExternalFabricOrdererModel `tfsdk:"orderers"`
}

type ExternalFabricOrdererModel struct {
	ID               types.Int64  `tfsdk:"id"`
	ExternalNodeID   types.Int64  `tfsdk:"external_node_id"`
	Name             types.String `tfsdk:"name"`
	MSPID            types.String `tfsdk:"msp_id"`
	ExternalEndpoint types.String `tfsdk:"external_endpoint"`
	Version          types.String `tfsdk:"version"`
	SignCertificate  types.String `tfsdk:"sign_certificate"`
	TLSCertificate   types.String `tfsdk:"tls_certificate"`
}

func (d *ExternalFabricOrderersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_fabric_orderers"
}

func (d *ExternalFabricOrderersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "List all external Fabric orderers that have been synced from remote Chainlaunch instances.",
		MarkdownDescription: "List all external Fabric orderers that have been synced from remote Chainlaunch instances.",

		Attributes: map[string]schema.Attribute{
			"orderers": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "List of external Fabric orderers",
				MarkdownDescription: "List of external Fabric orderers",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:            true,
							Description:         "External orderer ID",
							MarkdownDescription: "External orderer ID",
						},
						"external_node_id": schema.Int64Attribute{
							Computed:            true,
							Description:         "ID of the external node this orderer belongs to",
							MarkdownDescription: "ID of the external node this orderer belongs to",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							Description:         "Orderer name",
							MarkdownDescription: "Orderer name",
						},
						"msp_id": schema.StringAttribute{
							Computed:            true,
							Description:         "MSP ID of the orderer's organization",
							MarkdownDescription: "MSP ID of the orderer's organization",
						},
						"external_endpoint": schema.StringAttribute{
							Computed:            true,
							Description:         "External endpoint address",
							MarkdownDescription: "External endpoint address",
						},
						"version": schema.StringAttribute{
							Computed:            true,
							Description:         "Fabric version",
							MarkdownDescription: "Fabric version",
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

func (d *ExternalFabricOrderersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ExternalFabricOrderersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExternalFabricOrderersDataSourceModel

	body, err := d.client.DoRequest("GET", "/external-nodes/fabric-orderers", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read external Fabric orderers, got error: %s", err))
		return
	}

	var orderers []map[string]interface{}
	if err := json.Unmarshal(body, &orderers); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse orderers: %s", err))
		return
	}

	data.Orderers = make([]ExternalFabricOrdererModel, 0, len(orderers))
	for _, orderer := range orderers {
		ordererModel := ExternalFabricOrdererModel{}

		if id, ok := orderer["id"].(float64); ok {
			ordererModel.ID = types.Int64Value(int64(id))
		}

		if externalNodeID, ok := orderer["externalNodeId"].(float64); ok {
			ordererModel.ExternalNodeID = types.Int64Value(int64(externalNodeID))
		}

		if name, ok := orderer["name"].(string); ok {
			ordererModel.Name = types.StringValue(name)
		}

		if mspID, ok := orderer["mspId"].(string); ok {
			ordererModel.MSPID = types.StringValue(mspID)
		}

		if endpoint, ok := orderer["externalEndpoint"].(string); ok {
			ordererModel.ExternalEndpoint = types.StringValue(endpoint)
		}

		if version, ok := orderer["version"].(string); ok {
			ordererModel.Version = types.StringValue(version)
		}

		if signCert, ok := orderer["signCertificate"].(string); ok {
			ordererModel.SignCertificate = types.StringValue(signCert)
		}

		if tlsCert, ok := orderer["tlsCertificate"].(string); ok {
			ordererModel.TLSCertificate = types.StringValue(tlsCert)
		}

		data.Orderers = append(data.Orderers, ordererModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
