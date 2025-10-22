package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &PluginDeploymentResource{}
	_ resource.ResourceWithConfigure = &PluginDeploymentResource{}
)

func NewPluginDeploymentResource() resource.Resource {
	return &PluginDeploymentResource{}
}

type PluginDeploymentResource struct {
	client *Client
}

type PluginDeploymentResourceModel struct {
	PluginName  types.String `tfsdk:"plugin_name"`
	Parameters  types.String `tfsdk:"parameters"`
	Status      types.String `tfsdk:"status"`
	ProjectName types.String `tfsdk:"project_name"`
	StartedAt   types.String `tfsdk:"started_at"`
	StoppedAt   types.String `tfsdk:"stopped_at"`
	Error       types.String `tfsdk:"error"`
	ID          types.String `tfsdk:"id"`
}

func (r *PluginDeploymentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plugin_deployment"
}

func (r *PluginDeploymentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Deploys a Chainlaunch plugin with specific parameters. This manages the lifecycle of a plugin deployment including starting, stopping, and monitoring the deployed services. When parameters change, the plugin is automatically stopped and redeployed with the new configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Internal identifier for the deployment (uses plugin_name)",
				Computed:    true,
			},
			"plugin_name": schema.StringAttribute{
				Description: "Name of the plugin to deploy (must exist)",
				Required:    true,
			},
			"parameters": schema.StringAttribute{
				Description: "JSON-encoded deployment parameters. The structure depends on the plugin's parameter schema. Changing this value will trigger a stop and restart of the plugin deployment.",
				Required:    true,
			},
			"status": schema.StringAttribute{
				Description: "Deployment status (e.g., 'deployed', 'stopped', 'error')",
				Computed:    true,
			},
			"project_name": schema.StringAttribute{
				Description: "Docker Compose project name for this deployment",
				Computed:    true,
			},
			"started_at": schema.StringAttribute{
				Description: "Timestamp when deployment started",
				Computed:    true,
			},
			"stopped_at": schema.StringAttribute{
				Description: "Timestamp when deployment stopped (if applicable)",
				Computed:    true,
			},
			"error": schema.StringAttribute{
				Description: "Error message if deployment failed",
				Computed:    true,
			},
		},
	}
}

func (r *PluginDeploymentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PluginDeploymentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PluginDeploymentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse parameters JSON
	var parameters map[string]interface{}
	if err := json.Unmarshal([]byte(data.Parameters.ValueString()), &parameters); err != nil {
		resp.Diagnostics.AddError("Failed to parse parameters JSON", err.Error())
		return
	}

	// Deploy plugin
	_, err := r.client.DoRequest("POST", fmt.Sprintf("/plugins/%s/deploy", data.PluginName.ValueString()), parameters)
	if err != nil {
		resp.Diagnostics.AddError("Failed to deploy plugin", err.Error())
		return
	}

	// Set ID
	data.ID = data.PluginName

	// Wait for deployment to start and get status
	if err := r.waitForDeploymentReady(ctx, data.PluginName.ValueString(), 60*time.Second); err != nil {
		resp.Diagnostics.AddWarning(
			"Deployment Status Check",
			fmt.Sprintf("Could not verify deployment status: %v. The deployment may still be in progress.", err),
		)
	}

	// Get deployment status
	statusResp, err := r.client.DoRequest("GET", fmt.Sprintf("/plugins/%s/deployment-status", data.PluginName.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddWarning("Failed to get deployment status", err.Error())
	} else {
		var statusResult map[string]interface{}
		if err := json.Unmarshal(statusResp, &statusResult); err == nil {
			if status, ok := statusResult["status"].(string); ok {
				data.Status = types.StringValue(status)
			}
			if projectName, ok := statusResult["projectName"].(string); ok {
				data.ProjectName = types.StringValue(projectName)
			}
			if startedAt, ok := statusResult["startedAt"].(string); ok && startedAt != "" {
				data.StartedAt = types.StringValue(startedAt)
			} else {
				data.StartedAt = types.StringValue("")
			}
			if stoppedAt, ok := statusResult["stoppedAt"].(string); ok && stoppedAt != "" {
				data.StoppedAt = types.StringValue(stoppedAt)
			} else {
				data.StoppedAt = types.StringValue("")
			}
			if errorMsg, ok := statusResult["error"].(string); ok && errorMsg != "" {
				data.Error = types.StringValue(errorMsg)
			} else {
				data.Error = types.StringValue("")
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PluginDeploymentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PluginDeploymentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get deployment status
	statusResp, err := r.client.DoRequest("GET", fmt.Sprintf("/plugins/%s/deployment-status", data.PluginName.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get deployment status", err.Error())
		return
	}

	var statusResult map[string]interface{}
	if err := json.Unmarshal(statusResp, &statusResult); err != nil {
		resp.Diagnostics.AddError("Failed to parse deployment status", err.Error())
		return
	}

	// Update status fields
	if status, ok := statusResult["status"].(string); ok {
		data.Status = types.StringValue(status)
	}
	if projectName, ok := statusResult["projectName"].(string); ok {
		data.ProjectName = types.StringValue(projectName)
	}
	if startedAt, ok := statusResult["startedAt"].(string); ok && startedAt != "" {
		data.StartedAt = types.StringValue(startedAt)
	}
	if stoppedAt, ok := statusResult["stoppedAt"].(string); ok && stoppedAt != "" {
		data.StoppedAt = types.StringValue(stoppedAt)
	}
	if errorMsg, ok := statusResult["error"].(string); ok && errorMsg != "" {
		data.Error = types.StringValue(errorMsg)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PluginDeploymentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PluginDeploymentResourceModel
	var state PluginDeploymentResourceModel

	// Get current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve ID
	data.ID = state.ID

	// Stop current deployment
	_, err := r.client.DoRequest("POST", fmt.Sprintf("/plugins/%s/stop", data.PluginName.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddWarning("Failed to stop plugin", err.Error())
	}

	// Parse parameters JSON
	var parameters map[string]interface{}
	if err := json.Unmarshal([]byte(data.Parameters.ValueString()), &parameters); err != nil {
		resp.Diagnostics.AddError("Failed to parse parameters JSON", err.Error())
		return
	}

	// Redeploy with new parameters
	_, err = r.client.DoRequest("POST", fmt.Sprintf("/plugins/%s/deploy", data.PluginName.ValueString()), parameters)
	if err != nil {
		resp.Diagnostics.AddError("Failed to redeploy plugin", err.Error())
		return
	}

	// Wait for deployment to start
	if err := r.waitForDeploymentReady(ctx, data.PluginName.ValueString(), 60*time.Second); err != nil {
		resp.Diagnostics.AddWarning(
			"Deployment Status Check",
			fmt.Sprintf("Could not verify deployment status: %v", err),
		)
	}

	// Get updated deployment status
	statusResp, err := r.client.DoRequest("GET", fmt.Sprintf("/plugins/%s/deployment-status", data.PluginName.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddWarning("Failed to get deployment status", err.Error())
	} else {
		var statusResult map[string]interface{}
		if err := json.Unmarshal(statusResp, &statusResult); err == nil {
			if status, ok := statusResult["status"].(string); ok {
				data.Status = types.StringValue(status)
			}
			if projectName, ok := statusResult["projectName"].(string); ok {
				data.ProjectName = types.StringValue(projectName)
			}
			if startedAt, ok := statusResult["startedAt"].(string); ok && startedAt != "" {
				data.StartedAt = types.StringValue(startedAt)
			} else {
				data.StartedAt = types.StringValue("")
			}
			if stoppedAt, ok := statusResult["stoppedAt"].(string); ok && stoppedAt != "" {
				data.StoppedAt = types.StringValue(stoppedAt)
			} else {
				data.StoppedAt = types.StringValue("")
			}
			if errorMsg, ok := statusResult["error"].(string); ok && errorMsg != "" {
				data.Error = types.StringValue(errorMsg)
			} else {
				data.Error = types.StringValue("")
			}
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PluginDeploymentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PluginDeploymentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Stop plugin deployment
	_, err := r.client.DoRequest("POST", fmt.Sprintf("/plugins/%s/stop", data.PluginName.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to stop plugin deployment", err.Error())
		return
	}
}

func (r *PluginDeploymentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("plugin_name"), req, resp)
}

// waitForDeploymentReady waits for the plugin deployment to reach a ready state
func (r *PluginDeploymentResource) waitForDeploymentReady(ctx context.Context, pluginName string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	checkInterval := 2 * time.Second

	for time.Now().Before(deadline) {
		statusResp, err := r.client.DoRequest("GET", fmt.Sprintf("/plugins/%s/deployment-status", pluginName), nil)
		if err == nil {
			var statusResult map[string]interface{}
			if err := json.Unmarshal(statusResp, &statusResult); err == nil {
				status, ok := statusResult["status"].(string)
				if ok && (status == "deployed" || status == "error") {
					return nil
				}
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(checkInterval):
			// Continue waiting
		}
	}

	return fmt.Errorf("timeout waiting for deployment to be ready")
}
