package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &FabricChaincodeDataSource{}

func NewFabricChaincodeDataSource() datasource.DataSource {
	return &FabricChaincodeDataSource{}
}

type FabricChaincodeDataSource struct {
	client *Client
}

type FabricChaincodeDataSourceModel struct {
	ID              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	NetworkID       types.Int64  `tfsdk:"network_id"`
	NetworkName     types.String `tfsdk:"network_name"`
	NetworkPlatform types.String `tfsdk:"network_platform"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

func (d *FabricChaincodeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_chaincode"
}

func (d *FabricChaincodeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a Fabric chaincode by name and network ID.",

		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "The unique identifier for the chaincode.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the chaincode to look up.",
			},
			"network_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the Fabric network (channel) this chaincode belongs to.",
			},
			"network_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the network (channel).",
			},
			"network_platform": schema.StringAttribute{
				Computed:    true,
				Description: "The network platform (e.g., 'FABRIC').",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the chaincode was created.",
			},
		},
	}
}

func (d *FabricChaincodeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FabricChaincodeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FabricChaincodeDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// List all chaincodes and find the one matching name and network_id
	body, err := d.client.DoRequest("GET", "/sc/fabric/chaincodes", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list chaincodes: %s", err))
		return
	}

	var listResp struct {
		Chaincodes []struct {
			ID              int64  `json:"id"`
			Name            string `json:"name"`
			NetworkID       int64  `json:"network_id"`
			NetworkName     string `json:"network_name"`
			NetworkPlatform string `json:"network_platform"`
			CreatedAt       string `json:"created_at"`
		} `json:"chaincodes"`
	}
	if err := json.Unmarshal(body, &listResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Find matching chaincode
	var found bool
	for _, cc := range listResp.Chaincodes {
		if cc.Name == data.Name.ValueString() && cc.NetworkID == data.NetworkID.ValueInt64() {
			data.ID = types.Int64Value(cc.ID)
			data.Name = types.StringValue(cc.Name)
			data.NetworkID = types.Int64Value(cc.NetworkID)
			data.NetworkName = types.StringValue(cc.NetworkName)
			data.NetworkPlatform = types.StringValue(cc.NetworkPlatform)
			data.CreatedAt = types.StringValue(cc.CreatedAt)
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"Chaincode Not Found",
			fmt.Sprintf("No chaincode found with name '%s' on network ID %d", data.Name.ValueString(), data.NetworkID.ValueInt64()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
