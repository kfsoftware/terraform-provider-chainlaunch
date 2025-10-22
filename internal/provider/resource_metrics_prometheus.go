package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &MetricsPrometheusResource{}

func NewMetricsPrometheusResource() resource.Resource {
	return &MetricsPrometheusResource{}
}

type MetricsPrometheusResource struct {
	client *Client
}

type MetricsPrometheusResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Version        types.String `tfsdk:"version"`
	Port           types.Int64  `tfsdk:"port"`
	ScrapeInterval types.Int64  `tfsdk:"scrape_interval"`
	DeploymentMode types.String `tfsdk:"deployment_mode"`
	NetworkMode    types.String `tfsdk:"network_mode"`
	Status         types.String `tfsdk:"status"`
	StartedAt      types.String `tfsdk:"started_at"`
}

func (r *MetricsPrometheusResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metrics_prometheus"
}

func (r *MetricsPrometheusResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Deploys and manages a Prometheus monitoring instance for Chainlaunch nodes.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier (always 'prometheus' as only one instance is supported)",
			},
			"version": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("v2.45.0"),
				MarkdownDescription: "Prometheus version to deploy (default: v2.45.0)",
			},
			"port": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(9090),
				MarkdownDescription: "Port for Prometheus server (default: 9090)",
			},
			"scrape_interval": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(15),
				MarkdownDescription: "Scrape interval in seconds (default: 15)",
			},
			"deployment_mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("docker"),
				MarkdownDescription: "Deployment mode: docker or binary (default: docker)",
			},
			"network_mode": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("bridge"),
				MarkdownDescription: "Docker network mode: bridge or host (default: bridge)",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Current status of Prometheus instance",
			},
			"started_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when Prometheus was started",
			},
		},
	}
}

func (r *MetricsPrometheusResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MetricsPrometheusResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MetricsPrometheusResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployReq := map[string]interface{}{
		"prometheus_version": data.Version.ValueString(),
		"prometheus_port":    data.Port.ValueInt64(),
		"scrape_interval":    data.ScrapeInterval.ValueInt64(),
		"deployment_mode":    data.DeploymentMode.ValueString(),
	}

	// Add docker config if using docker mode
	if data.DeploymentMode.ValueString() == "docker" {
		deployReq["docker_config"] = map[string]interface{}{
			"network_mode": data.NetworkMode.ValueString(),
		}
	}

	body, err := r.client.DoRequest("POST", "/metrics/deploy", deployReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to deploy Prometheus, got error: %s", err))
		return
	}

	var deployResp map[string]interface{}
	if err := json.Unmarshal(body, &deployResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse deployment response: %s", err))
		return
	}

	// Wait a moment for Prometheus to start
	// Then read the status
	if err := r.readStatus(ctx, &data); err != nil {
		resp.Diagnostics.AddWarning("Status Check", fmt.Sprintf("Prometheus deployed but status check failed: %s", err))
	}

	data.ID = types.StringValue("prometheus")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetricsPrometheusResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MetricsPrometheusResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readStatus(ctx, &data); err != nil {
		// If Prometheus is not running, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetricsPrometheusResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MetricsPrometheusResourceModel
	var state MetricsPrometheusResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve computed fields from state
	data.ID = state.ID

	// Prometheus requires recreation for config changes
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Prometheus configuration changes require recreation. Use terraform apply -replace to redeploy.",
	)
}

func (r *MetricsPrometheusResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MetricsPrometheusResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Stop Prometheus
	_, err := r.client.DoRequest("POST", "/metrics/stop", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to stop Prometheus, got error: %s", err))
		return
	}
}

func (r *MetricsPrometheusResource) readStatus(ctx context.Context, data *MetricsPrometheusResourceModel) error {
	body, err := r.client.DoRequest("GET", "/metrics/status", nil)
	if err != nil {
		return err
	}

	var status map[string]interface{}
	if err := json.Unmarshal(body, &status); err != nil {
		return err
	}

	if version, ok := status["version"].(string); ok {
		data.Version = types.StringValue(version)
	}

	if port, ok := status["port"].(float64); ok {
		data.Port = types.Int64Value(int64(port))
	}

	if statusStr, ok := status["status"].(string); ok {
		data.Status = types.StringValue(statusStr)
	}

	if startedAt, ok := status["started_at"].(string); ok {
		data.StartedAt = types.StringValue(startedAt)
	}

	if deploymentMode, ok := status["deployment_mode"].(string); ok {
		data.DeploymentMode = types.StringValue(deploymentMode)
	}

	if networkMode, ok := status["network_mode"].(string); ok {
		data.NetworkMode = types.StringValue(networkMode)
	}

	return nil
}
