package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &FabricNetworkConfigDataSource{}

func NewFabricNetworkConfigDataSource() datasource.DataSource {
	return &FabricNetworkConfigDataSource{}
}

type FabricNetworkConfigDataSource struct {
	client *Client
}

type FabricNetworkConfigDataSourceModel struct {
	NetworkID      types.Int64  `tfsdk:"network_id"`
	OrganizationID types.Int64  `tfsdk:"organization_id"`
	Config         types.String `tfsdk:"config"`
}

func (d *FabricNetworkConfigDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_network_config"
}

func (d *FabricNetworkConfigDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Fetches the network configuration (connection profile) for a Fabric network channel.

This data source retrieves the network configuration YAML for a specific organization in a Fabric network.
The configuration includes connection information for peers, orderers, and certificate authorities,
which can be used by Fabric SDK clients to connect to the network.

## Example Usage

` + "```hcl" + `
# Get network config for Org1
data "chainlaunch_fabric_network_config" "org1_config" {
  network_id      = chainlaunch_fabric_network.mychannel.id
  organization_id = chainlaunch_fabric_organization.org1.id
}

# Write config to file for SDK use
resource "local_file" "connection_profile" {
  content  = data.chainlaunch_fabric_network_config.org1_config.config
  filename = "${path.module}/connection-profile.yaml"
}

# Output the config
output "network_config" {
  value     = data.chainlaunch_fabric_network_config.org1_config.config
  sensitive = true
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"network_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The ID of the Fabric network (channel) to get the configuration for.",
			},
			"organization_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "The ID of the organization to get the network configuration for. The configuration will include connection details relevant to this organization.",
			},
			"config": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The network configuration in YAML format. This is a connection profile that can be used by Fabric SDK clients.",
			},
		},
	}
}

func (d *FabricNetworkConfigDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FabricNetworkConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FabricNetworkConfigDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call the API to get network configuration
	// GET /networks/fabric/{id}/organizations/{orgId}/network-config
	path := fmt.Sprintf("/networks/fabric/%d/organizations/%d/network-config",
		data.NetworkID.ValueInt64(),
		data.OrganizationID.ValueInt64(),
	)

	tflog.Debug(ctx, "Fetching Fabric network config", map[string]interface{}{
		"path":            path,
		"network_id":      data.NetworkID.ValueInt64(),
		"organization_id": data.OrganizationID.ValueInt64(),
		"full_url":        fmt.Sprintf("%s/api/v1%s", d.client.BaseURL, path),
	})

	body, err := d.client.DoRequest("GET", path, nil)
	if err != nil {
		tflog.Error(ctx, "Failed to get Fabric network config", map[string]interface{}{
			"error": err.Error(),
			"path":  path,
		})
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Unable to get Fabric network configuration (path: %s): %s", path, err),
		)
		return
	}

	tflog.Debug(ctx, "Successfully fetched Fabric network config", map[string]interface{}{
		"response_length": len(body),
	})

	// The API returns YAML content directly as text
	data.Config = types.StringValue(string(body))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
