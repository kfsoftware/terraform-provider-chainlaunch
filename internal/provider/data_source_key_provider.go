package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &KeyProviderDataSource{}

func NewKeyProviderDataSource() datasource.DataSource {
	return &KeyProviderDataSource{}
}

type KeyProviderDataSource struct {
	client *Client
}

type KeyProviderDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	Type      types.String `tfsdk:"type"`
	Status    types.String `tfsdk:"status"`
	Config    types.String `tfsdk:"config"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

func (d *KeyProviderDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key_provider"
}

func (d *KeyProviderDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details about a specific key provider from Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The unique identifier of the key provider.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the key provider.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of key provider.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The current status of the key provider.",
			},
			"config": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "JSON configuration for the key provider.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the key provider was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the key provider was last updated.",
			},
		},
	}
}

func (d *KeyProviderDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KeyProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KeyProviderDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := d.client.DoRequest("GET", fmt.Sprintf("/key-providers/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read key provider, got error: %s", err))
		return
	}

	var keyProvider KeyProvider
	if err := json.Unmarshal(body, &keyProvider); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse key provider response, got error: %s", err))
		return
	}

	data.Name = types.StringValue(keyProvider.Name)
	data.Type = types.StringValue(keyProvider.Type)
	if keyProvider.Status != "" {
		data.Status = types.StringValue(keyProvider.Status)
	}
	if keyProvider.Config != nil {
		configJSON, err := json.Marshal(keyProvider.Config)
		if err == nil {
			data.Config = types.StringValue(string(configJSON))
		}
	}
	if keyProvider.CreatedAt != "" {
		data.CreatedAt = types.StringValue(keyProvider.CreatedAt)
	}
	if keyProvider.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(keyProvider.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
