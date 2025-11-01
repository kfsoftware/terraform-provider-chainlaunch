package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SyncAllExternalNodesResource{}

func NewSyncAllExternalNodesResource() resource.Resource {
	return &SyncAllExternalNodesResource{}
}

type SyncAllExternalNodesResource struct {
	client *Client
}

type SyncAllExternalNodesResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	LastSyncAt            types.String `tfsdk:"last_sync_at"`
	PeerNodeIDs           types.List   `tfsdk:"peer_node_ids"`
	OrganizationsAdded    types.Int64  `tfsdk:"organizations_added"`
	FabricPeersAdded      types.Int64  `tfsdk:"fabric_peers_added"`
	FabricPeersDeleted    types.Int64  `tfsdk:"fabric_peers_deleted"`
	FabricOrderersAdded   types.Int64  `tfsdk:"fabric_orderers_added"`
	FabricOrderersDeleted types.Int64  `tfsdk:"fabric_orderers_deleted"`
	BesuNodesAdded        types.Int64  `tfsdk:"besu_nodes_added"`
	BesuNodesDeleted      types.Int64  `tfsdk:"besu_nodes_deleted"`
	SyncResults           types.List   `tfsdk:"sync_results"`
}

type SyncResultModel struct {
	PeerNodeID            types.String `tfsdk:"peer_node_id"`
	Success               types.Bool   `tfsdk:"success"`
	Error                 types.String `tfsdk:"error"`
	OrganizationsAdded    types.Int64  `tfsdk:"organizations_added"`
	FabricPeersAdded      types.Int64  `tfsdk:"fabric_peers_added"`
	FabricPeersDeleted    types.Int64  `tfsdk:"fabric_peers_deleted"`
	FabricOrderersAdded   types.Int64  `tfsdk:"fabric_orderers_added"`
	FabricOrderersDeleted types.Int64  `tfsdk:"fabric_orderers_deleted"`
	BesuNodesAdded        types.Int64  `tfsdk:"besu_nodes_added"`
	BesuNodesDeleted      types.Int64  `tfsdk:"besu_nodes_deleted"`
}

func (r *SyncAllExternalNodesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sync_all_external_nodes"
}

func (r *SyncAllExternalNodesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Automatically synchronizes external nodes from ALL connected remote Chainlaunch instances.

This resource discovers all connected peers and syncs their nodes automatically - no need to specify individual peer_node_ids. It always runs on every terraform apply to keep external nodes synchronized.

The resource will:
- Discover all connected peer nodes automatically
- Sync nodes from each connected peer in parallel
- Aggregate results across all peers
- Track per-peer sync results and errors

Use this instead of 'chainlaunch_external_nodes_sync' when you want to sync from all connected peers without manually specifying each peer_node_id.`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier (timestamp of sync)",
			},
			"last_sync_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp of the last successful sync",
			},
			"peer_node_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of peer node IDs that were synced",
			},
			"organizations_added": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Total number of organizations added across all peers",
			},
			"fabric_peers_added": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Total number of Fabric peers added across all peers",
			},
			"fabric_peers_deleted": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Total number of Fabric peers deleted across all peers",
			},
			"fabric_orderers_added": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Total number of Fabric orderers added across all peers",
			},
			"fabric_orderers_deleted": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Total number of Fabric orderers deleted across all peers",
			},
			"besu_nodes_added": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Total number of Besu nodes added across all peers",
			},
			"besu_nodes_deleted": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Total number of Besu nodes deleted across all peers",
			},
			"sync_results": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Detailed sync results for each peer",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"peer_node_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The peer node ID that was synced",
						},
						"success": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the sync was successful",
						},
						"error": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Error message if sync failed",
						},
						"organizations_added": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of organizations added from this peer",
						},
						"fabric_peers_added": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of Fabric peers added from this peer",
						},
						"fabric_peers_deleted": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of Fabric peers deleted from this peer",
						},
						"fabric_orderers_added": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of Fabric orderers added from this peer",
						},
						"fabric_orderers_deleted": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of Fabric orderers deleted from this peer",
						},
						"besu_nodes_added": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of Besu nodes added from this peer",
						},
						"besu_nodes_deleted": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of Besu nodes deleted from this peer",
						},
					},
				},
			},
		},
	}
}

func (r *SyncAllExternalNodesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *SyncAllExternalNodesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SyncAllExternalNodesResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Perform sync
	if err := r.performSyncAll(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Sync Error", fmt.Sprintf("Unable to sync external nodes from all peers: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SyncAllExternalNodesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SyncAllExternalNodesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Always trigger sync on read by performing sync again
	// This ensures "terraform refresh" will sync external nodes
	if err := r.performSyncAll(ctx, &data); err != nil {
		resp.Diagnostics.AddWarning("Sync Warning", fmt.Sprintf("Unable to sync external nodes from all peers: %s", err))
		// Don't fail, just warn - keep existing state
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SyncAllExternalNodesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SyncAllExternalNodesResourceModel
	var state SyncAllExternalNodesResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Perform sync
	if err := r.performSyncAll(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Sync Error", fmt.Sprintf("Unable to sync external nodes from all peers: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SyncAllExternalNodesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// When resource is deleted, we don't delete the external nodes
	// They remain in the system as external nodes
	// Just remove from Terraform state
}

func (r *SyncAllExternalNodesResource) performSyncAll(ctx context.Context, data *SyncAllExternalNodesResourceModel) error {
	// Step 1: Get all connected peers
	body, err := r.client.DoRequest("GET", "/node/connected-peers", nil)
	if err != nil {
		return fmt.Errorf("unable to get connected peers: %w", err)
	}

	var connectedPeersResp struct {
		ConnectedPeers []struct {
			NodeID string `json:"node_id"`
		} `json:"connected_peers"`
	}
	if err := json.Unmarshal(body, &connectedPeersResp); err != nil {
		return fmt.Errorf("unable to parse connected peers response: %w", err)
	}

	if len(connectedPeersResp.ConnectedPeers) == 0 {
		// No connected peers, set empty state
		data.ID = types.StringValue(fmt.Sprintf("sync-%d", time.Now().Unix()))
		data.LastSyncAt = types.StringValue(time.Now().Format(time.RFC3339))
		emptyList, _ := types.ListValueFrom(ctx, types.StringType, []string{})
		data.PeerNodeIDs = emptyList
		data.OrganizationsAdded = types.Int64Value(0)
		data.FabricPeersAdded = types.Int64Value(0)
		data.FabricPeersDeleted = types.Int64Value(0)
		data.FabricOrderersAdded = types.Int64Value(0)
		data.FabricOrderersDeleted = types.Int64Value(0)
		data.BesuNodesAdded = types.Int64Value(0)
		data.BesuNodesDeleted = types.Int64Value(0)
		emptySyncResults, _ := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"peer_node_id":             types.StringType,
				"success":                  types.BoolType,
				"error":                    types.StringType,
				"organizations_added":      types.Int64Type,
				"fabric_peers_added":       types.Int64Type,
				"fabric_peers_deleted":     types.Int64Type,
				"fabric_orderers_added":    types.Int64Type,
				"fabric_orderers_deleted":  types.Int64Type,
				"besu_nodes_added":         types.Int64Type,
				"besu_nodes_deleted":       types.Int64Type,
			},
		}, []SyncResultModel{})
		data.SyncResults = emptySyncResults
		return nil
	}

	// Step 2: Sync from each peer
	var (
		totalOrgsAdded         int64
		totalPeersAdded        int64
		totalPeersDeleted      int64
		totalOrderersAdded     int64
		totalOrderersDeleted   int64
		totalBesuNodesAdded    int64
		totalBesuNodesDeleted  int64
		peerNodeIDs            []string
		syncResults            []SyncResultModel
	)

	for _, peer := range connectedPeersResp.ConnectedPeers {
		syncReq := map[string]interface{}{
			"peer_node_id": peer.NodeID,
		}

		syncBody, syncErr := r.client.DoRequest("POST", "/node/sync-external-nodes", syncReq)

		result := SyncResultModel{
			PeerNodeID: types.StringValue(peer.NodeID),
		}

		if syncErr != nil {
			// Record failure
			result.Success = types.BoolValue(false)
			result.Error = types.StringValue(syncErr.Error())
			result.OrganizationsAdded = types.Int64Value(0)
			result.FabricPeersAdded = types.Int64Value(0)
			result.FabricPeersDeleted = types.Int64Value(0)
			result.FabricOrderersAdded = types.Int64Value(0)
			result.FabricOrderersDeleted = types.Int64Value(0)
			result.BesuNodesAdded = types.Int64Value(0)
			result.BesuNodesDeleted = types.Int64Value(0)
		} else {
			var syncResp map[string]interface{}
			if err := json.Unmarshal(syncBody, &syncResp); err != nil {
				result.Success = types.BoolValue(false)
				result.Error = types.StringValue(fmt.Sprintf("unable to parse sync response: %v", err))
				result.OrganizationsAdded = types.Int64Value(0)
				result.FabricPeersAdded = types.Int64Value(0)
				result.FabricPeersDeleted = types.Int64Value(0)
				result.FabricOrderersAdded = types.Int64Value(0)
				result.FabricOrderersDeleted = types.Int64Value(0)
				result.BesuNodesAdded = types.Int64Value(0)
				result.BesuNodesDeleted = types.Int64Value(0)
			} else {
				// Success - extract results
				result.Success = types.BoolValue(true)
				result.Error = types.StringValue("")

				orgsAdded := getInt64FromResponse(syncResp, "organizations_added")
				peersAdded := getInt64FromResponse(syncResp, "fabric_peers_added")
				peersDeleted := getInt64FromResponse(syncResp, "fabric_peers_deleted")
				orderersAdded := getInt64FromResponse(syncResp, "fabric_orderers_added")
				orderersDeleted := getInt64FromResponse(syncResp, "fabric_orderers_deleted")
				besuAdded := getInt64FromResponse(syncResp, "besu_nodes_added")
				besuDeleted := getInt64FromResponse(syncResp, "besu_nodes_deleted")

				result.OrganizationsAdded = types.Int64Value(orgsAdded)
				result.FabricPeersAdded = types.Int64Value(peersAdded)
				result.FabricPeersDeleted = types.Int64Value(peersDeleted)
				result.FabricOrderersAdded = types.Int64Value(orderersAdded)
				result.FabricOrderersDeleted = types.Int64Value(orderersDeleted)
				result.BesuNodesAdded = types.Int64Value(besuAdded)
				result.BesuNodesDeleted = types.Int64Value(besuDeleted)

				// Aggregate totals
				totalOrgsAdded += orgsAdded
				totalPeersAdded += peersAdded
				totalPeersDeleted += peersDeleted
				totalOrderersAdded += orderersAdded
				totalOrderersDeleted += orderersDeleted
				totalBesuNodesAdded += besuAdded
				totalBesuNodesDeleted += besuDeleted
			}
		}

		peerNodeIDs = append(peerNodeIDs, peer.NodeID)
		syncResults = append(syncResults, result)
	}

	// Step 3: Update state
	data.ID = types.StringValue(fmt.Sprintf("sync-%d", time.Now().Unix()))
	data.LastSyncAt = types.StringValue(time.Now().Format(time.RFC3339))

	peerNodeIDsList, diags := types.ListValueFrom(ctx, types.StringType, peerNodeIDs)
	if diags.HasError() {
		return fmt.Errorf("unable to convert peer node IDs list")
	}
	data.PeerNodeIDs = peerNodeIDsList

	data.OrganizationsAdded = types.Int64Value(totalOrgsAdded)
	data.FabricPeersAdded = types.Int64Value(totalPeersAdded)
	data.FabricPeersDeleted = types.Int64Value(totalPeersDeleted)
	data.FabricOrderersAdded = types.Int64Value(totalOrderersAdded)
	data.FabricOrderersDeleted = types.Int64Value(totalOrderersDeleted)
	data.BesuNodesAdded = types.Int64Value(totalBesuNodesAdded)
	data.BesuNodesDeleted = types.Int64Value(totalBesuNodesDeleted)

	syncResultsList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"peer_node_id":             types.StringType,
			"success":                  types.BoolType,
			"error":                    types.StringType,
			"organizations_added":      types.Int64Type,
			"fabric_peers_added":       types.Int64Type,
			"fabric_peers_deleted":     types.Int64Type,
			"fabric_orderers_added":    types.Int64Type,
			"fabric_orderers_deleted":  types.Int64Type,
			"besu_nodes_added":         types.Int64Type,
			"besu_nodes_deleted":       types.Int64Type,
		},
	}, syncResults)
	if diags.HasError() {
		return fmt.Errorf("unable to convert sync results list")
	}
	data.SyncResults = syncResultsList

	return nil
}

func getInt64FromResponse(resp map[string]interface{}, key string) int64 {
	if val, ok := resp[key].(float64); ok {
		return int64(val)
	}
	return 0
}
