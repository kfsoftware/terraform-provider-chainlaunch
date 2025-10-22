package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NetworkDataSource{}

func NewNetworkDataSource() datasource.DataSource {
	return &NetworkDataSource{}
}

type NetworkDataSource struct {
	client *Client
}

type NetworkDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Status    types.String `tfsdk:"status"`
	Config    types.String `tfsdk:"config"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *NetworkDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (d *NetworkDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details about a specific blockchain network from Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the network.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the network.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of network (e.g., fabric, besu).",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the network.",
			},
			"config": schema.StringAttribute{
				Computed:    true,
				Description: "JSON configuration for the network.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the network was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the network was last updated.",
			},
		},
	}
}

func (d *NetworkDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NetworkDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/networks/%s/%s", data.Type.ValueString(), data.ID.ValueString())

	body, err := d.client.DoRequest("GET", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read network, got error: %s", err))
		return
	}

	var network Network
	if err := json.Unmarshal(body, &network); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse network response, got error: %s", err))
		return
	}

	data.Name = types.StringValue(network.Name)
	data.Type = types.StringValue(network.Type)
	if network.Status != "" {
		data.Status = types.StringValue(network.Status)
	}
	if network.Config != nil {
		configJSON, err := json.Marshal(network.Config)
		if err == nil {
			data.Config = types.StringValue(string(configJSON))
		}
	}
	if network.CreatedAt != "" {
		data.CreatedAt = types.StringValue(network.CreatedAt)
	}
	if network.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(network.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
