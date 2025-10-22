package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ExternalNodesSyncResource{}

func NewExternalNodesSyncResource() resource.Resource {
	return &ExternalNodesSyncResource{}
}

type ExternalNodesSyncResource struct {
	client *Client
}

type ExternalNodesSyncResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	PeerNodeID            types.String `tfsdk:"peer_node_id"`
	LastSyncAt            types.String `tfsdk:"last_sync_at"`
	OrganizationsAdded    types.Int64  `tfsdk:"organizations_added"`
	FabricPeersAdded      types.Int64  `tfsdk:"fabric_peers_added"`
	FabricPeersDeleted    types.Int64  `tfsdk:"fabric_peers_deleted"`
	FabricOrderersAdded   types.Int64  `tfsdk:"fabric_orderers_added"`
	FabricOrderersDeleted types.Int64  `tfsdk:"fabric_orderers_deleted"`
	BesuNodesAdded        types.Int64  `tfsdk:"besu_nodes_added"`
	BesuNodesDeleted      types.Int64  `tfsdk:"besu_nodes_deleted"`
	Errors                types.List   `tfsdk:"errors"`
}

func (r *ExternalNodesSyncResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_nodes_sync"
}

func (r *ExternalNodesSyncResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Synchronizes external nodes from a remote Chainlaunch instance.

This resource always runs on every terraform apply to keep external nodes synchronized with the remote instance. Use this after accepting a node invitation to automatically import all nodes from the remote instance.

The resource will:
- Fetch all nodes (peers, orderers, Besu nodes) from the remote instance
- Store them locally as external nodes
- Remove external nodes that no longer exist remotely
- Track what was added/deleted in each sync`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier (combination of peer_node_id and timestamp)",
			},
			"peer_node_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The remote peer node ID to sync from (typically 'node1', 'node2', etc.)",
			},
			"last_sync_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp of the last successful sync",
			},
			"organizations_added": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of organizations added in last sync",
			},
			"fabric_peers_added": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of Fabric peers added in last sync",
			},
			"fabric_peers_deleted": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of Fabric peers deleted in last sync",
			},
			"fabric_orderers_added": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of Fabric orderers added in last sync",
			},
			"fabric_orderers_deleted": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of Fabric orderers deleted in last sync",
			},
			"besu_nodes_added": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of Besu nodes added in last sync",
			},
			"besu_nodes_deleted": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of Besu nodes deleted in last sync",
			},
			"errors": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of errors encountered during sync (if any)",
			},
		},
	}
}

func (r *ExternalNodesSyncResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ExternalNodesSyncResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ExternalNodesSyncResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Perform sync
	if err := r.performSync(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Sync Error", fmt.Sprintf("Unable to sync external nodes: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExternalNodesSyncResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ExternalNodesSyncResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Always trigger sync on read by performing sync again
	// This ensures "terraform refresh" will sync external nodes
	if err := r.performSync(ctx, &data); err != nil {
		resp.Diagnostics.AddWarning("Sync Warning", fmt.Sprintf("Unable to sync external nodes: %s", err))
		// Don't fail, just warn - keep existing state
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExternalNodesSyncResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ExternalNodesSyncResourceModel
	var state ExternalNodesSyncResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Perform sync
	if err := r.performSync(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Sync Error", fmt.Sprintf("Unable to sync external nodes: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ExternalNodesSyncResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// When resource is deleted, we don't delete the external nodes
	// They remain in the system as external nodes
	// Just remove from Terraform state
}

func (r *ExternalNodesSyncResource) performSync(ctx context.Context, data *ExternalNodesSyncResourceModel) error {
	syncReq := map[string]interface{}{
		"peer_node_id": data.PeerNodeID.ValueString(),
	}

	body, err := r.client.DoRequest("POST", "/node/sync-external-nodes", syncReq)
	if err != nil {
		return err
	}

	var syncResp map[string]interface{}
	if err := json.Unmarshal(body, &syncResp); err != nil {
		return fmt.Errorf("unable to parse sync response: %w", err)
	}

	// Update state with sync results
	data.ID = types.StringValue(fmt.Sprintf("%s-%d", data.PeerNodeID.ValueString(), time.Now().Unix()))
	data.LastSyncAt = types.StringValue(time.Now().Format(time.RFC3339))

	if orgsAdded, ok := syncResp["organizations_added"].(float64); ok {
		data.OrganizationsAdded = types.Int64Value(int64(orgsAdded))
	} else {
		data.OrganizationsAdded = types.Int64Value(0)
	}

	if peersAdded, ok := syncResp["fabric_peers_added"].(float64); ok {
		data.FabricPeersAdded = types.Int64Value(int64(peersAdded))
	} else {
		data.FabricPeersAdded = types.Int64Value(0)
	}

	if peersDeleted, ok := syncResp["fabric_peers_deleted"].(float64); ok {
		data.FabricPeersDeleted = types.Int64Value(int64(peersDeleted))
	} else {
		data.FabricPeersDeleted = types.Int64Value(0)
	}

	if orderersAdded, ok := syncResp["fabric_orderers_added"].(float64); ok {
		data.FabricOrderersAdded = types.Int64Value(int64(orderersAdded))
	} else {
		data.FabricOrderersAdded = types.Int64Value(0)
	}

	if orderersDeleted, ok := syncResp["fabric_orderers_deleted"].(float64); ok {
		data.FabricOrderersDeleted = types.Int64Value(int64(orderersDeleted))
	} else {
		data.FabricOrderersDeleted = types.Int64Value(0)
	}

	if besuAdded, ok := syncResp["besu_nodes_added"].(float64); ok {
		data.BesuNodesAdded = types.Int64Value(int64(besuAdded))
	} else {
		data.BesuNodesAdded = types.Int64Value(0)
	}

	if besuDeleted, ok := syncResp["besu_nodes_deleted"].(float64); ok {
		data.BesuNodesDeleted = types.Int64Value(int64(besuDeleted))
	} else {
		data.BesuNodesDeleted = types.Int64Value(0)
	}

	// Handle errors list
	if errors, ok := syncResp["errors"].([]interface{}); ok && len(errors) > 0 {
		errorStrings := make([]string, len(errors))
		for i, e := range errors {
			if errStr, ok := e.(string); ok {
				errorStrings[i] = errStr
			}
		}
		errorList, diags := types.ListValueFrom(ctx, types.StringType, errorStrings)
		if diags.HasError() {
			return fmt.Errorf("unable to convert errors list")
		}
		data.Errors = errorList
	} else {
		emptyList, _ := types.ListValueFrom(ctx, types.StringType, []string{})
		data.Errors = emptyList
	}

	return nil
}
