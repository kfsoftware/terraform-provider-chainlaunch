package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NodeDataSource{}

func NewNodeDataSource() datasource.DataSource {
	return &NodeDataSource{}
}

type NodeDataSource struct {
	client *Client
}

type NodeDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Platform  types.String `tfsdk:"platform"`
	Type      types.String `tfsdk:"type"`
	Status    types.String `tfsdk:"status"`
	Config    types.String `tfsdk:"config"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *NodeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node"
}

func (d *NodeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details about a specific blockchain node from Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the node.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the node.",
			},
			"platform": schema.StringAttribute{
				Computed:    true,
				Description: "The blockchain platform.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of node.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the node.",
			},
			"config": schema.StringAttribute{
				Computed:    true,
				Description: "JSON configuration for the node.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the node was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the node was last updated.",
			},
		},
	}
}

func (d *NodeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NodeDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.DoRequest("GET", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read node, got error: %s", err))
		return
	}

	var node Node
	if err := json.Unmarshal(body, &node); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse node response, got error: %s", err))
		return
	}

	data.Name = types.StringValue(node.Name)
	data.Platform = types.StringValue(node.Platform)
	data.Type = types.StringValue(node.Type)
	if node.Status != "" {
		data.Status = types.StringValue(node.Status)
	}
	if node.Config != nil {
		configJSON, err := json.Marshal(node.Config)
		if err == nil {
			data.Config = types.StringValue(string(configJSON))
		}
	}
	if node.CreatedAt != "" {
		data.CreatedAt = types.StringValue(node.CreatedAt)
	}
	if node.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(node.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
