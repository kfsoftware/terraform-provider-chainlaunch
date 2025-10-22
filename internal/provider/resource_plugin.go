package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gopkg.in/yaml.v3"
)

var (
	_ resource.Resource              = &PluginResource{}
	_ resource.ResourceWithConfigure = &PluginResource{}
)

func NewPluginResource() resource.Resource {
	return &PluginResource{}
}

type PluginResource struct {
	client *Client
}

type PluginResourceModel struct {
	Name            types.String `tfsdk:"name"`
	YAMLFilePath    types.String `tfsdk:"yaml_file_path"`
	YAMLContent     types.String `tfsdk:"yaml_content"`
	APIVersion      types.String `tfsdk:"api_version"`
	Kind            types.String `tfsdk:"kind"`
	MetadataVersion types.String `tfsdk:"metadata_version"`
	Description     types.String `tfsdk:"description"`
	Author          types.String `tfsdk:"author"`
	Repository      types.String `tfsdk:"repository"`
	License         types.String `tfsdk:"license"`
}

func (r *PluginResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plugin"
}

func (r *PluginResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Chainlaunch plugin definition. Plugins extend the platform with custom functionality using Docker Compose configurations.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "Unique name of the plugin (from metadata.name in YAML)",
				Computed:    true,
			},
			"yaml_file_path": schema.StringAttribute{
				Description: "Path to the plugin YAML file. Either yaml_file_path or yaml_content must be provided.",
				Optional:    true,
			},
			"yaml_content": schema.StringAttribute{
				Description: "Raw YAML content of the plugin. Either yaml_file_path or yaml_content must be provided.",
				Optional:    true,
			},
			"api_version": schema.StringAttribute{
				Description: "API version of the plugin (e.g., 'dev.chainlaunch/v1')",
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
				Description: "Description of the plugin from metadata",
				Computed:    true,
			},
			"author": schema.StringAttribute{
				Description: "Author of the plugin from metadata",
				Computed:    true,
			},
			"repository": schema.StringAttribute{
				Description: "Repository URL of the plugin from metadata",
				Computed:    true,
			},
			"license": schema.StringAttribute{
				Description: "License of the plugin from metadata",
				Computed:    true,
			},
		},
	}
}

func (r *PluginResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ChainlaunchClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *PluginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PluginResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read plugin YAML
	var yamlContent string
	if !data.YAMLFilePath.IsNull() && data.YAMLFilePath.ValueString() != "" {
		fileContent, err := os.ReadFile(data.YAMLFilePath.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read plugin YAML file", err.Error())
			return
		}
		yamlContent = string(fileContent)
	} else if !data.YAMLContent.IsNull() && data.YAMLContent.ValueString() != "" {
		yamlContent = data.YAMLContent.ValueString()
	} else {
		resp.Diagnostics.AddError(
			"Missing plugin YAML",
			"Either yaml_file_path or yaml_content must be provided",
		)
		return
	}

	// Parse YAML to extract plugin structure
	var pluginData map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &pluginData); err != nil {
		resp.Diagnostics.AddError("Failed to parse plugin YAML", err.Error())
		return
	}

	// Create plugin via API
	pluginResp, err := r.client.DoRequest("POST", "/plugins", pluginData)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create plugin", err.Error())
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
	if name, ok := metadata["name"].(string); ok {
		data.Name = types.StringValue(name)
	}
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PluginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PluginResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plugin from API
	pluginResp, err := r.client.DoRequest("GET", fmt.Sprintf("/plugins/%s", data.Name.ValueString()), nil)
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

	// Update computed fields
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PluginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PluginResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read plugin YAML
	var yamlContent string
	if !data.YAMLFilePath.IsNull() && data.YAMLFilePath.ValueString() != "" {
		fileContent, err := os.ReadFile(data.YAMLFilePath.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read plugin YAML file", err.Error())
			return
		}
		yamlContent = string(fileContent)
	} else if !data.YAMLContent.IsNull() && data.YAMLContent.ValueString() != "" {
		yamlContent = data.YAMLContent.ValueString()
	} else {
		resp.Diagnostics.AddError(
			"Missing plugin YAML",
			"Either yaml_file_path or yaml_content must be provided",
		)
		return
	}

	// Parse YAML to extract plugin structure
	var pluginData map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &pluginData); err != nil {
		resp.Diagnostics.AddError("Failed to parse plugin YAML", err.Error())
		return
	}

	// Update plugin via API
	pluginResp, err := r.client.DoRequest("PUT", fmt.Sprintf("/plugins/%s", data.Name.ValueString()), pluginData)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update plugin", err.Error())
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

	// Update computed fields
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PluginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PluginResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete plugin via API
	_, err := r.client.DoRequest("DELETE", fmt.Sprintf("/plugins/%s", data.Name.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete plugin", err.Error())
		return
	}
}

func (r *PluginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
