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
var _ datasource.DataSource = &OrganizationDataSource{}

func NewOrganizationDataSource() datasource.DataSource {
	return &OrganizationDataSource{}
}

// OrganizationDataSource defines the data source implementation.
type OrganizationDataSource struct {
	client *Client
}

// OrganizationDataSourceModel describes the data source data model.
type OrganizationDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	MSPID       types.String `tfsdk:"msp_id"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (d *OrganizationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_organization"
}

func (d *OrganizationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details about a specific Hyperledger Fabric organization from Chainlaunch.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The unique identifier of the organization. Either id or msp_id must be specified.",
			},
			"msp_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The MSP ID of the organization. Either id or msp_id must be specified.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "A description of the organization.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the organization was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the organization was last updated.",
			},
		},
	}
}

func (d *OrganizationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OrganizationDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either ID or MSP ID is provided
	hasID := !data.ID.IsNull() && data.ID.ValueString() != ""
	hasMSPID := !data.MSPID.IsNull() && data.MSPID.ValueString() != ""

	if !hasID && !hasMSPID {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either 'id' or 'msp_id' must be specified to look up an organization.",
		)
		return
	}

	var org Organization

	if hasID {
		// Lookup by ID - direct GET
		body, err := d.client.DoRequest("GET", fmt.Sprintf("/organizations/%s", data.ID.ValueString()), nil)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read organization, got error: %s", err))
			return
		}

		if err := json.Unmarshal(body, &org); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse organization response, got error: %s", err))
			return
		}
	} else {
		// Lookup by MSP ID using query parameter (API-side filtering)
		body, err := d.client.DoRequest("GET", fmt.Sprintf("/organizations?mspId=%s", data.MSPID.ValueString()), nil)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to search for organization by msp_id, got error: %s", err))
			return
		}

		// Parse paginated response
		var paginatedResp struct {
			Items []Organization `json:"items"`
			Count int            `json:"count"`
		}

		if err := json.Unmarshal(body, &paginatedResp); err != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse paginated organizations response, got error: %s", err))
			return
		}

		// Find the organization with matching MSP ID
		found := false
		for _, item := range paginatedResp.Items {
			if item.MSPID == data.MSPID.ValueString() {
				org = item
				found = true
				break
			}
		}

		if !found {
			resp.Diagnostics.AddError(
				"Organization Not Found",
				fmt.Sprintf("No organization found with msp_id: %s", data.MSPID.ValueString()),
			)
			return
		}
	}

	// Set all fields from the organization
	data.ID = types.StringValue(fmt.Sprintf("%d", org.ID))
	data.MSPID = types.StringValue(org.MSPID)
	if org.Description != "" {
		data.Description = types.StringValue(org.Description)
	}
	if org.CreatedAt != "" {
		data.CreatedAt = types.StringValue(org.CreatedAt)
	}
	if org.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(org.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
