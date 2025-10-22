package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &PluginDataSource{}
	_ datasource.DataSourceWithConfigure = &PluginDataSource{}
)

func NewPluginDataSource() datasource.DataSource {
	return &PluginDataSource{}
}

type PluginDataSource struct {
	client *Client
}

type PluginDataSourceModel struct {
	Name            types.String `tfsdk:"name"`
	APIVersion      types.String `tfsdk:"api_version"`
	Kind            types.String `tfsdk:"kind"`
	MetadataVersion types.String `tfsdk:"metadata_version"`
	Description     types.String `tfsdk:"description"`
	Author          types.String `tfsdk:"author"`
	Repository      types.String `tfsdk:"repository"`
	License         types.String `tfsdk:"license"`
	Tags            types.List   `tfsdk:"tags"`
	Status          types.String `tfsdk:"status"`
	ProjectName     types.String `tfsdk:"project_name"`
}

func (d *PluginDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plugin"
}

func (d *PluginDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a Chainlaunch plugin.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Unique name of the plugin",
				Required:    true,
			},
			"api_version": schema.StringAttribute{
				Description: "API version of the plugin",
				Computed:    true,
			},
			"kind": schema.StringAttribute{
				Description: "Kind of the resource (always 'Plugin')",
				Computed:    true,
			},
			"metadata_version": schema.StringAttribute{
				Description: "Version of the plugin from metadata",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the plugin",
				Computed:    true,
			},
			"author": schema.StringAttribute{
				Description: "Author of the plugin",
				Computed:    true,
			},
			"repository": schema.StringAttribute{
				Description: "Repository URL of the plugin",
				Computed:    true,
			},
			"license": schema.StringAttribute{
				Description: "License of the plugin",
				Computed:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags associated with the plugin",
				Computed:    true,
				ElementType: types.StringType,
			},
			"status": schema.StringAttribute{
				Description: "Deployment status if deployed",
				Computed:    true,
			},
			"project_name": schema.StringAttribute{
				Description: "Docker Compose project name if deployed",
				Computed:    true,
			},
		},
	}
}

func (d *PluginDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ChainlaunchClient, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *PluginDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PluginDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plugin from API
	pluginResp, err := d.client.DoRequest("GET", fmt.Sprintf("/plugins/%s", data.Name.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read plugin", err.Error())
		return
	}

	// Parse response
	var pluginResult map[string]interface{}
	if err := json.Unmarshal(pluginResp, &pluginResult); err != nil {
		resp.Diagnostics.AddError("Failed to parse plugin response", err.Error())
		return
	}

	// Extract metadata
	metadata, ok := pluginResult["metadata"].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("Invalid plugin response", "Missing metadata field")
		return
	}

	// Set computed fields
	if apiVersion, ok := pluginResult["apiVersion"].(string); ok {
		data.APIVersion = types.StringValue(apiVersion)
	}
	if kind, ok := pluginResult["kind"].(string); ok {
		data.Kind = types.StringValue(kind)
	}
	if version, ok := metadata["version"].(string); ok {
		data.MetadataVersion = types.StringValue(version)
	}
	if description, ok := metadata["description"].(string); ok {
		data.Description = types.StringValue(description)
	}
	if author, ok := metadata["author"].(string); ok {
		data.Author = types.StringValue(author)
	}
	if repository, ok := metadata["repository"].(string); ok {
		data.Repository = types.StringValue(repository)
	}
	if license, ok := metadata["license"].(string); ok {
		data.License = types.StringValue(license)
	}

	// Extract tags
	if tags, ok := metadata["tags"].([]interface{}); ok {
		var tagsList []string
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				tagsList = append(tagsList, tagStr)
			}
		}
		tagsListValue, diags := types.ListValueFrom(ctx, types.StringType, tagsList)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			data.Tags = tagsListValue
		}
	}

	// Try to get deployment status
	statusResp, err := d.client.DoRequest("GET", fmt.Sprintf("/plugins/%s/deployment-status", data.Name.ValueString()), nil)
	if err == nil {
		var statusResult map[string]interface{}
		if err := json.Unmarshal(statusResp, &statusResult); err == nil {
			if status, ok := statusResult["status"].(string); ok {
				data.Status = types.StringValue(status)
			}
			if projectName, ok := statusResult["projectName"].(string); ok {
				data.ProjectName = types.StringValue(projectName)
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
