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
var _ datasource.DataSource = &FabricOrdererDataSource{}

func NewFabricOrdererDataSource() datasource.DataSource {
	return &FabricOrdererDataSource{}
}

// FabricOrdererDataSource defines the data source implementation.
type FabricOrdererDataSource struct {
	client *Client
}

// FabricOrdererDataSourceModel describes the data source data model.
type FabricOrdererDataSourceModel struct {
	ID                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	OrganizationID          types.Int64  `tfsdk:"organization_id"`
	Mode                    types.String `tfsdk:"mode"`
	Version                 types.String `tfsdk:"version"`
	ExternalEndpoint        types.String `tfsdk:"external_endpoint"`
	ListenAddress           types.String `tfsdk:"listen_address"`
	AdminAddress            types.String `tfsdk:"admin_address"`
	OperationsListenAddress types.String `tfsdk:"operations_listen_address"`
	Status                  types.String `tfsdk:"status"`
	CreatedAt               types.String `tfsdk:"created_at"`
	UpdatedAt               types.String `tfsdk:"updated_at"`
}

func (d *FabricOrdererDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_orderer"
}

func (d *FabricOrdererDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details about a specific Hyperledger Fabric orderer node from Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The unique identifier of the orderer node. Either id or name must be specified.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the orderer node. Either id or name must be specified.",
			},
			"organization_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the organization that owns this orderer.",
			},
			"mode": schema.StringAttribute{
				Computed:    true,
				Description: "Deployment mode (e.g., 'docker').",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Hyperledger Fabric orderer version.",
			},
			"external_endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "External endpoint for the orderer (e.g., orderer0.example.com:7050).",
			},
			"listen_address": schema.StringAttribute{
				Computed:    true,
				Description: "Orderer listen address.",
			},
			"admin_address": schema.StringAttribute{
				Computed:    true,
				Description: "Admin listen address.",
			},
			"operations_listen_address": schema.StringAttribute{
				Computed:    true,
				Description: "Operations/metrics listen address.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the orderer (e.g., RUNNING, CREATING, ERROR).",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the orderer was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the orderer was last updated.",
			},
		},
	}
}

func (d *FabricOrdererDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FabricOrdererDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FabricOrdererDataSourceModel

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
			"Either 'id' or 'name' must be specified to look up a fabric orderer.",
		)
		return
	}

	// Use NodeResponse which has the fabricOrderer nested object
	var nodeResp struct {
		ID            int64                  `json:"id"`
		Name          string                 `json:"name"`
		Platform      string                 `json:"blockchainPlatform,omitempty"`
		Type          string                 `json:"type,omitempty"`
		Status        string                 `json:"status,omitempty"`
		CreatedAt     string                 `json:"createdAt,omitempty"`
		UpdatedAt     string                 `json:"updatedAt,omitempty"`
		Config        map[string]interface{} `json:"config,omitempty"`
		FabricOrderer map[string]interface{} `json:"fabricOrderer,omitempty"`
	}

	if hasID {
		// Lookup by ID
		body, err := d.client.DoRequest("GET", fmt.Sprintf("/nodes/%s", data.ID.ValueString()), nil)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read fabric orderer, got error: %s", err))
			return
		}

		if err := json.Unmarshal(body, &nodeResp); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse orderer response, got error: %s", err))
			return
		}
	} else {
		// Lookup by name using query parameter (API-side filtering)
		body, err := d.client.DoRequest("GET", fmt.Sprintf("/nodes?name=%s", data.Name.ValueString()), nil)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to search for fabric orderer by name, got error: %s", err))
			return
		}

		// Parse paginated response
		var paginatedResp struct {
			Items []struct {
				ID            int64                  `json:"id"`
				Name          string                 `json:"name"`
				Platform      string                 `json:"blockchainPlatform,omitempty"`
				Type          string                 `json:"type,omitempty"`
				Status        string                 `json:"status,omitempty"`
				CreatedAt     string                 `json:"createdAt,omitempty"`
				UpdatedAt     string                 `json:"updatedAt,omitempty"`
				Config        map[string]interface{} `json:"config,omitempty"`
				FabricOrderer map[string]interface{} `json:"fabricOrderer,omitempty"`
			} `json:"items"`
			Total int `json:"total"`
		}

		if err := json.Unmarshal(body, &paginatedResp); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse paginated nodes response, got error: %s", err))
			return
		}

		// Find the orderer with matching name
		found := false
		for _, node := range paginatedResp.Items {
			if node.Name == data.Name.ValueString() && node.FabricOrderer != nil {
				nodeResp = node
				found = true
				break
			}
		}

		if !found {
			resp.Diagnostics.AddError(
				"Orderer Not Found",
				fmt.Sprintf("No fabric orderer found with name: %s", data.Name.ValueString()),
			)
			return
		}
	}

	// Set ID and Name
	data.ID = types.StringValue(fmt.Sprintf("%d", nodeResp.ID))
	data.Name = types.StringValue(nodeResp.Name)

	// Extract data from fabricOrderer nested object if present
	if nodeResp.FabricOrderer != nil {
		if orgID, ok := nodeResp.FabricOrderer["organizationId"].(float64); ok {
			data.OrganizationID = types.Int64Value(int64(orgID))
		}
		if mode, ok := nodeResp.FabricOrderer["mode"].(string); ok {
			data.Mode = types.StringValue(mode)
		}
		if version, ok := nodeResp.FabricOrderer["version"].(string); ok {
			data.Version = types.StringValue(version)
		}
		if externalEndpoint, ok := nodeResp.FabricOrderer["externalEndpoint"].(string); ok {
			data.ExternalEndpoint = types.StringValue(externalEndpoint)
		}
		if listenAddr, ok := nodeResp.FabricOrderer["listenAddress"].(string); ok {
			data.ListenAddress = types.StringValue(listenAddr)
		}
		if adminAddr, ok := nodeResp.FabricOrderer["adminAddress"].(string); ok {
			data.AdminAddress = types.StringValue(adminAddr)
		}
		if opsAddr, ok := nodeResp.FabricOrderer["operationsListenAddress"].(string); ok {
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
