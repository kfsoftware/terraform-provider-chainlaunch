package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ChainlaunchProvider satisfies various provider interfaces.
var _ provider.Provider = &ChainlaunchProvider{}

// ChainlaunchProvider defines the provider implementation.
type ChainlaunchProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ChainlaunchProviderModel describes the provider data model.
type ChainlaunchProviderModel struct {
	URL      types.String `tfsdk:"url"`
	APIKey   types.String `tfsdk:"api_key"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

func (p *ChainlaunchProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "chainlaunch"
	resp.Version = p.version
}

func (p *ChainlaunchProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Chainlaunch to manage Hyperledger Fabric organizations, nodes, networks, and key providers. " +
			"Supports authentication via API key or username/password.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "The Chainlaunch API URL. Can also be set via the CHAINLAUNCH_URL environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The Chainlaunch API key for authentication. Can also be set via the CHAINLAUNCH_API_KEY environment variable. " +
					"Use either api_key or username/password for authentication.",
				Optional:  true,
				Sensitive: true,
			},
			"username": schema.StringAttribute{
				Description: "The Chainlaunch username for basic authentication. Can also be set via the CHAINLAUNCH_USERNAME environment variable. " +
					"Use either api_key or username/password for authentication.",
				Optional: true,
			},
			"password": schema.StringAttribute{
				Description: "The Chainlaunch password for basic authentication. Can also be set via the CHAINLAUNCH_PASSWORD environment variable. " +
					"Required when username is provided.",
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (p *ChainlaunchProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ChainlaunchProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	url := os.Getenv("CHAINLAUNCH_URL")
	apiKey := os.Getenv("CHAINLAUNCH_API_KEY")
	username := os.Getenv("CHAINLAUNCH_USERNAME")
	password := os.Getenv("CHAINLAUNCH_PASSWORD")

	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}

	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// Validate configuration

	if url == "" {
		resp.Diagnostics.AddError(
			"Missing Chainlaunch API URL",
			"The provider cannot create the Chainlaunch API client as there is a missing or empty value for the Chainlaunch API URL. "+
				"Set the url value in the configuration or use the CHAINLAUNCH_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	// Check that either API key or username/password is provided
	hasAPIKey := apiKey != ""
	hasUsernamePassword := username != "" && password != ""

	if !hasAPIKey && !hasUsernamePassword {
		resp.Diagnostics.AddError(
			"Missing Authentication Credentials",
			"The provider requires authentication credentials. "+
				"Provide either an API key (api_key) or username and password (username and password). "+
				"These can be set in the configuration or use environment variables: "+
				"CHAINLAUNCH_API_KEY or CHAINLAUNCH_USERNAME and CHAINLAUNCH_PASSWORD.",
		)
	}

	if username != "" && password == "" {
		resp.Diagnostics.AddError(
			"Missing Password",
			"Password is required when username is provided.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Chainlaunch client using the configuration values
	client := NewClient(url, apiKey, username, password)

	// Make the Chainlaunch client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ChainlaunchProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewOrganizationResource,
		NewNodeResource,
		NewNetworkResource,
		NewKeyProviderResource,
		NewKeyResource,
		NewFabricPeerResource,
		NewFabricOrdererResource,
		NewFabricNetworkResource,
		NewFabricIdentityResource,
		NewFabricAddNodeResource,
		NewFabricJoinNodeResource,
		NewFabricAnchorPeersResource,
		NewBesuNetworkResource,
		NewBesuNodeResource,
		NewFabricChaincodeResource,
		NewFabricChaincodeDefinitionResource,
		NewFabricChaincodeInstallResource,
		NewFabricChaincodeApproveResource,
		NewFabricChaincodeCommitResource,
		NewFabricChaincodeDeployResource,
		NewBackupTargetResource,
		NewBackupScheduleResource,
		NewNodeInvitationResource,
		NewNodeAcceptInvitationResource,
		NewExternalNodesSyncResource,
		NewSyncAllExternalNodesResource,
		NewMetricsPrometheusResource,
		NewMetricsJobResource,
		NewNotificationProviderResource,
		NewPluginResource,
		NewPluginDeploymentResource,
		NewChainlaunchInstallSSHResource,
	}
}

func (p *ChainlaunchProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewOrganizationDataSource,
		NewNodeDataSource,
		NewNetworkDataSource,
		NewKeyProviderDataSource,
		NewKeyProvidersDataSource,
		NewFabricPeerDataSource,
		NewFabricOrdererDataSource,
		NewFabricNetworkDataSource,
		NewBesuNetworkDataSource,
		NewBesuNodeDataSource,
		NewFabricChaincodeDataSource,
		NewExternalFabricOrganizationsDataSource,
		NewExternalFabricPeersDataSource,
		NewExternalFabricOrderersDataSource,
		NewExternalBesuNodesDataSource,
		NewPluginDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ChainlaunchProvider{
			version: version,
		}
	}
}
