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
var _ datasource.DataSource = &KeyProvidersDataSource{}

func NewKeyProvidersDataSource() datasource.DataSource {
	return &KeyProvidersDataSource{}
}

// KeyProvidersDataSource defines the data source implementation.
type KeyProvidersDataSource struct {
	client *Client
}

// KeyProviderItem represents a single key provider in the list
type KeyProviderItem struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Type   types.String `tfsdk:"type"`
	Status types.String `tfsdk:"status"`
}

// KeyProvidersDataSourceModel describes the data source data model.
type KeyProvidersDataSourceModel struct {
	ID                  types.String      `tfsdk:"id"`
	NameFilter          types.String      `tfsdk:"name_filter"`
	TypeFilter          types.String      `tfsdk:"type_filter"`
	Providers           []KeyProviderItem `tfsdk:"providers"`
	DefaultProviderID   types.Int64       `tfsdk:"default_provider_id"`
	DefaultProviderName types.String      `tfsdk:"default_provider_name"`
	DefaultProviderType types.String      `tfsdk:"default_provider_type"`
}

func (d *KeyProvidersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_key_providers"
}

func (d *KeyProvidersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a list of key providers from Chainlaunch. Can filter by name or type, and automatically identifies the default provider.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Placeholder identifier for the data source.",
			},
			"name_filter": schema.StringAttribute{
				Optional:    true,
				Description: "Filter providers by name (case-insensitive partial match).",
			},
			"type_filter": schema.StringAttribute{
				Optional:    true,
				Description: "Filter providers by type (e.g., 'database', 'vault', 'aws-kms').",
			},
			"providers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of key providers matching the filters.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
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
					},
				},
			},
			"default_provider_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The ID of the default/database key provider (if found).",
			},
			"default_provider_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the default/database key provider (if found).",
			},
			"default_provider_type": schema.StringAttribute{
				Computed:    true,
				Description: "The type of the default/database key provider (if found).",
			},
		},
	}
}

func (d *KeyProvidersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *KeyProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data KeyProvidersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch all key providers
	body, err := d.client.DoRequest("GET", "/key-providers", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read key providers, got error: %s", err))
		return
	}

	var keyProviders []KeyProvider
	if err := json.Unmarshal(body, &keyProviders); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse key providers response, got error: %s", err))
		return
	}

	// Filter and process providers
	var filteredProviders []KeyProviderItem
	var defaultProvider *KeyProvider

	nameFilter := data.NameFilter.ValueString()
	typeFilter := data.TypeFilter.ValueString()

	for _, provider := range keyProviders {
		// Check if this is the default provider (case-insensitive)
		providerTypeLower := toLower(provider.Type)
		if providerTypeLower == "database" || provider.Name == "Default Database Provider" {
			if defaultProvider == nil {
				defaultProvider = &provider
			}
		}

		// Apply filters (case-insensitive for type)
		matchesName := nameFilter == "" || containsIgnoreCase(provider.Name, nameFilter)
		matchesType := typeFilter == "" || toLower(provider.Type) == toLower(typeFilter)

		if matchesName && matchesType {
			item := KeyProviderItem{
				ID:   types.StringValue(fmt.Sprintf("%d", provider.ID)),
				Name: types.StringValue(provider.Name),
				Type: types.StringValue(provider.Type),
			}
			if provider.Status != "" {
				item.Status = types.StringValue(provider.Status)
			} else {
				item.Status = types.StringValue("unknown")
			}
			filteredProviders = append(filteredProviders, item)
		}
	}

	// Set the filtered providers
	data.Providers = filteredProviders

	// Set default provider info if found
	if defaultProvider != nil {
		data.DefaultProviderID = types.Int64Value(defaultProvider.ID)
		data.DefaultProviderName = types.StringValue(defaultProvider.Name)
		data.DefaultProviderType = types.StringValue(defaultProvider.Type)
	} else {
		data.DefaultProviderID = types.Int64Null()
		data.DefaultProviderName = types.StringNull()
		data.DefaultProviderType = types.StringNull()
	}

	// Set placeholder ID
	data.ID = types.StringValue("key-providers")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return contains(s, substr)
}

func toLower(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			result[i] = r + ('a' - 'A')
		} else {
			result[i] = r
		}
	}
	return string(result)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
