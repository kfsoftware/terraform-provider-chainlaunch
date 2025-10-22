package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &FabricNetworkDataSource{}

func NewFabricNetworkDataSource() datasource.DataSource {
	return &FabricNetworkDataSource{}
}

type FabricNetworkDataSource struct {
	client *Client
}

type FabricNetworkDataSourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Platform    types.String `tfsdk:"platform"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

func (d *FabricNetworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_network"
}

func (d *FabricNetworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a Fabric network (channel) by name.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier for the network.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Fabric network (channel) to look up.",
			},
			"platform": schema.StringAttribute{
				Computed:    true,
				Description: "The blockchain platform (e.g., fabric).",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the network.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the network.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the network was created.",
			},
		},
	}
}

func (d *FabricNetworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FabricNetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FabricNetworkDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// List all Fabric networks and find the one matching the name
	body, err := d.client.DoRequest("GET", "/networks/fabric", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list Fabric networks: %s", err))
		return
	}

	var listResp struct {
		Networks []struct {
			ID          int64  `json:"id"`
			Name        string `json:"name"`
			Platform    string `json:"platform"`
			Description string `json:"description"`
			Status      string `json:"status"`
			CreatedAt   string `json:"createdAt"`
		} `json:"networks"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Find matching network
	var found bool
	for _, network := range listResp.Networks {
		if network.Name == data.Name.ValueString() {
			data.ID = types.Int64Value(network.ID)
			data.Name = types.StringValue(network.Name)
			data.Platform = types.StringValue(network.Platform)
			data.Description = types.StringValue(network.Description)
			data.Status = types.StringValue(network.Status)
			data.CreatedAt = types.StringValue(network.CreatedAt)
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"Fabric Network Not Found",
			fmt.Sprintf("No Fabric network found with name '%s'", data.Name.ValueString()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
