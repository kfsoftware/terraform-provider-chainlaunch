package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &MetricsJobResource{}

func NewMetricsJobResource() resource.Resource {
	return &MetricsJobResource{}
}

type MetricsJobResource struct {
	client *Client
}

type MetricsJobResourceModel struct {
	ID             types.String `tfsdk:"id"`
	JobName        types.String `tfsdk:"job_name"`
	Targets        types.List   `tfsdk:"targets"`
	MetricsPath    types.String `tfsdk:"metrics_path"`
	ScrapeInterval types.String `tfsdk:"scrape_interval"`
}

func (r *MetricsJobResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_metrics_job"
}

func (r *MetricsJobResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Prometheus scrape job for monitoring Chainlaunch nodes. This resource automatically synchronizes node endpoints with Prometheus configuration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier (same as job_name)",
			},
			"job_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the Prometheus job (e.g., 'fabric-peers', 'fabric-orderers')",
			},
			"targets": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "List of target endpoints in 'host:port' format",
			},
			"metrics_path": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("/metrics"),
				MarkdownDescription: "Path to metrics endpoint (default: /metrics)",
			},
			"scrape_interval": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("15s"),
				MarkdownDescription: "Scrape interval for this job (default: 15s)",
			},
		},
	}
}

func (r *MetricsJobResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MetricsJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MetricsJobResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert targets list to string slice
	var targets []string
	resp.Diagnostics.Append(data.Targets.ElementsAs(ctx, &targets, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobReq := map[string]interface{}{
		"job_name":        data.JobName.ValueString(),
		"targets":         targets,
		"metrics_path":    data.MetricsPath.ValueString(),
		"scrape_interval": data.ScrapeInterval.ValueString(),
	}

	body, err := r.client.DoRequest("POST", "/metrics/job/add", jobReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create metrics job, got error: %s", err))
		return
	}

	var jobResp map[string]interface{}
	if err := json.Unmarshal(body, &jobResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse job response: %s", err))
		return
	}

	data.ID = types.StringValue(data.JobName.ValueString())

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetricsJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MetricsJobResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// List all jobs and find this one
	body, err := r.client.DoRequest("GET", "/metrics/jobs", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read metrics jobs, got error: %s", err))
		return
	}

	var jobs map[string]interface{}
	if err := json.Unmarshal(body, &jobs); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse jobs response: %s", err))
		return
	}

	jobName := data.JobName.ValueString()
	if _, exists := jobs[jobName]; !exists {
		// Job doesn't exist, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Job exists, keep current state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetricsJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MetricsJobResourceModel
	var state MetricsJobResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve ID from state
	data.ID = state.ID

	// Job name cannot change (it's the ID)
	if !data.JobName.Equal(state.JobName) {
		resp.Diagnostics.AddError(
			"Job Name Change Not Supported",
			"Job name cannot be changed. Delete and recreate the job instead.",
		)
		return
	}

	// Delete old job
	jobName := state.JobName.ValueString()
	_, err := r.client.DoRequest("DELETE", fmt.Sprintf("/metrics/job/%s", jobName), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete old metrics job, got error: %s", err))
		return
	}

	// Create new job with updated config
	var targets []string
	resp.Diagnostics.Append(data.Targets.ElementsAs(ctx, &targets, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobReq := map[string]interface{}{
		"job_name":        data.JobName.ValueString(),
		"targets":         targets,
		"metrics_path":    data.MetricsPath.ValueString(),
		"scrape_interval": data.ScrapeInterval.ValueString(),
	}

	_, err = r.client.DoRequest("POST", "/metrics/job/add", jobReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update metrics job, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MetricsJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MetricsJobResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	jobName := data.JobName.ValueString()
	_, err := r.client.DoRequest("DELETE", fmt.Sprintf("/metrics/job/%s", jobName), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete metrics job, got error: %s", err))
		return
	}
}
