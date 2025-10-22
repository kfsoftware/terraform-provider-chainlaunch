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
var _ datasource.DataSource = &FabricPeerDataSource{}

func NewFabricPeerDataSource() datasource.DataSource {
	return &FabricPeerDataSource{}
}

// FabricPeerDataSource defines the data source implementation.
type FabricPeerDataSource struct {
	client *Client
}

// FabricPeerDataSourceModel describes the data source data model.
type FabricPeerDataSourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	OrganizationID          types.Int64  `tfsdk:"organization_id"`
	Mode                    types.String `tfsdk:"mode"`
	Version                 types.String `tfsdk:"version"`
	ExternalEndpoint        types.String `tfsdk:"external_endpoint"`
	ListenAddress           types.String `tfsdk:"listen_address"`
	ChaincodeAddress        types.String `tfsdk:"chaincode_address"`
	EventsAddress           types.String `tfsdk:"events_address"`
	OperationsListenAddress types.String `tfsdk:"operations_listen_address"`
	Status                  types.String `tfsdk:"status"`
	CreatedAt               types.String `tfsdk:"created_at"`
	UpdatedAt               types.String `tfsdk:"updated_at"`
}

func (d *FabricPeerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_peer"
}

func (d *FabricPeerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details about a specific Hyperledger Fabric peer node from Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The unique identifier of the peer node. Either id or name must be specified.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the peer node. Either id or name must be specified.",
			},
			"organization_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the organization that owns this peer.",
			},
			"mode": schema.StringAttribute{
				Computed:    true,
				Description: "Deployment mode (e.g., 'docker').",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Hyperledger Fabric peer version.",
			},
			"external_endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "External endpoint for the peer (e.g., peer0.org1.example.com:7051).",
			},
			"listen_address": schema.StringAttribute{
				Computed:    true,
				Description: "Peer listen address.",
			},
			"chaincode_address": schema.StringAttribute{
				Computed:    true,
				Description: "Chaincode listen address.",
			},
			"events_address": schema.StringAttribute{
				Computed:    true,
				Description: "Events listen address.",
			},
			"operations_listen_address": schema.StringAttribute{
				Computed:    true,
				Description: "Operations/metrics listen address.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the peer (e.g., RUNNING, CREATING, ERROR).",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the peer was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the peer was last updated.",
			},
		},
	}
}

func (d *FabricPeerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FabricPeerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FabricPeerDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either ID or name is provided
	hasID := !data.ID.IsNull() && data.ID.ValueString() != ""
	hasName := !data.Name.IsNull() && data.Name.ValueString() != ""

	if !hasID && !hasName {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'name' must be specified to look up a fabric peer.",
		)
		return
	}

	// Use NodeResponse which has the fabricPeer nested object
	var nodeResp struct {
		ID         int64                  `json:"id"`
		Name       string                 `json:"name"`
		Platform   string                 `json:"blockchainPlatform,omitempty"`
		Type       string                 `json:"type,omitempty"`
		Status     string                 `json:"status,omitempty"`
		CreatedAt  string                 `json:"createdAt,omitempty"`
		UpdatedAt  string                 `json:"updatedAt,omitempty"`
		Config     map[string]interface{} `json:"config,omitempty"`
		FabricPeer map[string]interface{} `json:"fabricPeer,omitempty"`
	}

	if hasID {
		// Lookup by ID
		body, err := d.client.DoRequest("GET", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), nil)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read fabric peer, got error: %s", err))
			return
		}

		if err := json.Unmarshal(body, &nodeResp); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse peer response, got error: %s", err))
			return
		}
	} else {
		// Lookup by name using query parameter (API-side filtering)
		body, err := d.client.DoRequest("GET", fmt.Sprintf("/nodes?name=%s", data.Name.ValueString()), nil)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to search for fabric peer by name, got error: %s", err))
			return
		}

		// Parse paginated response
		var paginatedResp struct {
			Items []struct {
				ID         int64                  `json:"id"`
				Name       string                 `json:"name"`
				Platform   string                 `json:"blockchainPlatform,omitempty"`
				Type       string                 `json:"type,omitempty"`
				Status     string                 `json:"status,omitempty"`
				CreatedAt  string                 `json:"createdAt,omitempty"`
				UpdatedAt  string                 `json:"updatedAt,omitempty"`
				Config     map[string]interface{} `json:"config,omitempty"`
				FabricPeer map[string]interface{} `json:"fabricPeer,omitempty"`
			} `json:"items"`
			Total int `json:"total"`
		}

		if err := json.Unmarshal(body, &paginatedResp); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse paginated nodes response, got error: %s", err))
			return
		}

		// Find the peer with matching name
		found := false
		for _, node := range paginatedResp.Items {
			if node.Name == data.Name.ValueString() && node.FabricPeer != nil {
				nodeResp = node
				found = true
				break
			}
		}

		if !found {
			resp.Diagnostics.AddError(
				"Peer Not Found",
				fmt.Sprintf("No fabric peer found with name: %s", data.Name.ValueString()),
			)
			return
		}
	}

	// Set ID and Name
	data.ID = types.StringValue(fmt.Sprintf("%d", nodeResp.ID))
	data.Name = types.StringValue(nodeResp.Name)

	// Extract data from fabricPeer nested object if present
	if nodeResp.FabricPeer != nil {
		if orgID, ok := nodeResp.FabricPeer["organizationId"].(float64); ok {
			data.OrganizationID = types.Int64Value(int64(orgID))
		}
		if mode, ok := nodeResp.FabricPeer["mode"].(string); ok {
			data.Mode = types.StringValue(mode)
		}
		if version, ok := nodeResp.FabricPeer["version"].(string); ok {
			data.Version = types.StringValue(version)
		}
		if externalEndpoint, ok := nodeResp.FabricPeer["externalEndpoint"].(string); ok {
			data.ExternalEndpoint = types.StringValue(externalEndpoint)
		}
		if listenAddr, ok := nodeResp.FabricPeer["listenAddress"].(string); ok {
			data.ListenAddress = types.StringValue(listenAddr)
		}
		if chaincodeAddr, ok := nodeResp.FabricPeer["chaincodeAddress"].(string); ok {
			data.ChaincodeAddress = types.StringValue(chaincodeAddr)
		}
		if eventsAddr, ok := nodeResp.FabricPeer["eventsAddress"].(string); ok {
			data.EventsAddress = types.StringValue(eventsAddr)
		}
		if opsAddr, ok := nodeResp.FabricPeer["operationsListenAddress"].(string); ok {
			data.OperationsListenAddress = types.StringValue(opsAddr)
		}
	}

	if nodeResp.Status != "" {
		data.Status = types.StringValue(nodeResp.Status)
	}
	if nodeResp.CreatedAt != "" {
		data.CreatedAt = types.StringValue(nodeResp.CreatedAt)
	}
	if nodeResp.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(nodeResp.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
