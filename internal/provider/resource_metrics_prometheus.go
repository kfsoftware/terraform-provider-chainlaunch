package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
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
	// RetentionTime sets --storage.tsdb.retention.time (e.g. "30d", "90d"). Empty
	// means the flag is not passed so Prometheus keeps its implicit 15d default.
	RetentionTime types.String `tfsdk:"retention_time"`
	// RetentionSize sets --storage.tsdb.retention.size (e.g. "50GB", "512MB").
	// Empty means no size-based retention limit is enforced.
	RetentionSize types.String `tfsdk:"retention_size"`
	// RemoteWrite configures one or more Prometheus remote_write endpoints so
	// metrics are shipped to an external long-term store. Empty means no
	// remote_write block is rendered.
	RemoteWrite types.List `tfsdk:"remote_write"`
	Status      types.String `tfsdk:"status"`
	StartedAt   types.String `tfsdk:"started_at"`
}

// RemoteWriteModel is a single remote_write endpoint as configured in Terraform.
// Credentials (basic_auth password, bearer_token) are write-only: they are sent
// to the API but never returned, so they are preserved from prior state on Read.
type RemoteWriteModel struct {
	URL         types.String          `tfsdk:"url"`
	Name        types.String          `tfsdk:"name"`
	BearerToken types.String          `tfsdk:"bearer_token"`
	BasicAuth   *RemoteWriteBasicAuth `tfsdk:"basic_auth"`
	TLSConfig   *RemoteWriteTLSConfig `tfsdk:"tls"`
}

// RemoteWriteBasicAuth holds HTTP basic-auth credentials for a remote_write endpoint.
type RemoteWriteBasicAuth struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// RemoteWriteTLSConfig mirrors the supported subset of the Prometheus tls_config block.
type RemoteWriteTLSConfig struct {
	CAFile             types.String `tfsdk:"ca_file"`
	CertFile           types.String `tfsdk:"cert_file"`
	KeyFile            types.String `tfsdk:"key_file"`
	ServerName         types.String `tfsdk:"server_name"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
}

// remoteWriteObjectType describes the object schema of a single remote_write
// entry. It is used to build the empty-list default and to convert the model
// list to/from the framework value.
func remoteWriteObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"url":          types.StringType,
			"name":         types.StringType,
			"bearer_token": types.StringType,
			"basic_auth": types.ObjectType{AttrTypes: map[string]attr.Type{
				"username": types.StringType,
				"password": types.StringType,
			}},
			"tls": types.ObjectType{AttrTypes: map[string]attr.Type{
				"ca_file":              types.StringType,
				"cert_file":            types.StringType,
				"key_file":             types.StringType,
				"server_name":          types.StringType,
				"insecure_skip_verify": types.BoolType,
			}},
		},
	}
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
			"retention_time": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "TSDB time-based retention passed to --storage.tsdb.retention.time (e.g. \"30d\", \"90d\"). Empty (default) leaves Prometheus at its implicit 15d retention. Applied on the next start/restart.",
			},
			"retention_size": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "TSDB size-based retention passed to --storage.tsdb.retention.size (e.g. \"50GB\", \"512MB\"). Empty (default) enforces no size-based limit. Applied on the next start/restart.",
			},
			"remote_write": schema.ListNestedAttribute{
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(remoteWriteObjectType(), []attr.Value{})),
				MarkdownDescription: "Prometheus remote_write endpoints used to ship metrics to an external long-term store (e.g. Grafana Cloud, Thanos, Mimir). Empty (default) renders no remote_write block. Credentials (basic_auth password, bearer_token) are write-only and never read back from the API.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"url": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The remote_write endpoint URL (e.g. https://prometheus-prod.grafana.net/api/prom/push)",
						},
						"name": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(""),
							MarkdownDescription: "Optional name for the remote_write endpoint",
						},
						"bearer_token": schema.StringAttribute{
							Optional:            true,
							Computed:            true,
							Sensitive:           true,
							Default:             stringdefault.StaticString(""),
							MarkdownDescription: "Bearer token for authenticating to the remote_write endpoint (write-only, never returned by the API)",
						},
						"basic_auth": schema.SingleNestedAttribute{
							Optional:            true,
							MarkdownDescription: "HTTP basic-auth credentials for the remote_write endpoint",
							Attributes: map[string]schema.Attribute{
								"username": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString(""),
									MarkdownDescription: "Basic-auth username",
								},
								"password": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									Sensitive:           true,
									Default:             stringdefault.StaticString(""),
									MarkdownDescription: "Basic-auth password (write-only, never returned by the API)",
								},
							},
						},
						"tls": schema.SingleNestedAttribute{
							Optional:            true,
							MarkdownDescription: "TLS configuration for the remote_write endpoint",
							Attributes: map[string]schema.Attribute{
								"ca_file": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString(""),
									MarkdownDescription: "Path to the CA certificate file (inside the Prometheus container/host)",
								},
								"cert_file": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString(""),
									MarkdownDescription: "Path to the client certificate file",
								},
								"key_file": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString(""),
									MarkdownDescription: "Path to the client key file",
								},
								"server_name": schema.StringAttribute{
									Optional:            true,
									Computed:            true,
									Default:             stringdefault.StaticString(""),
									MarkdownDescription: "ServerName used to verify the hostname on the returned certificate",
								},
								"insecure_skip_verify": schema.BoolAttribute{
									Optional:            true,
									Computed:            true,
									Default:             booldefault.StaticBool(false),
									MarkdownDescription: "Disable validation of the remote_write server certificate",
								},
							},
						},
					},
				},
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

	deployReq, diags := r.buildDeployRequest(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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

	// Preserve ID from state
	data.ID = state.ID

	// Stop the current Prometheus instance
	_, err := r.client.DoRequest("POST", "/metrics/stop", nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to stop Prometheus for update, got error: %s", err))
		return
	}

	// Redeploy with new configuration
	deployReq, diags := r.buildDeployRequest(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err = r.client.DoRequest("POST", "/metrics/deploy", deployReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to redeploy Prometheus, got error: %s", err))
		return
	}

	// Read the updated status
	if err := r.readStatus(ctx, &data); err != nil {
		resp.Diagnostics.AddWarning("Status Check", fmt.Sprintf("Prometheus redeployed but status check failed: %s", err))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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

// buildDeployRequest assembles the POST /metrics/deploy body from the planned
// model, including the remote_write endpoints and TSDB retention launch flags.
// Empty retention strings and an empty remote_write list are omitted so the API
// behaves exactly as it did before these fields existed.
func (r *MetricsPrometheusResource) buildDeployRequest(ctx context.Context, data *MetricsPrometheusResourceModel) (map[string]interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

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

	// TSDB retention launch flags (omitted when empty so default behavior is kept)
	if v := data.RetentionTime.ValueString(); v != "" {
		deployReq["retention_time"] = v
	}
	if v := data.RetentionSize.ValueString(); v != "" {
		deployReq["retention_size"] = v
	}

	// remote_write endpoints
	if !data.RemoteWrite.IsNull() && !data.RemoteWrite.IsUnknown() {
		var rws []RemoteWriteModel
		diags.Append(data.RemoteWrite.ElementsAs(ctx, &rws, false)...)
		if diags.HasError() {
			return deployReq, diags
		}
		if len(rws) > 0 {
			remoteWrites := make([]map[string]interface{}, 0, len(rws))
			for _, rw := range rws {
				entry := map[string]interface{}{
					"url": rw.URL.ValueString(),
				}
				if v := rw.Name.ValueString(); v != "" {
					entry["name"] = v
				}
				if v := rw.BearerToken.ValueString(); v != "" {
					entry["bearer_token"] = v
				}
				if rw.BasicAuth != nil {
					basicAuth := map[string]interface{}{}
					if v := rw.BasicAuth.Username.ValueString(); v != "" {
						basicAuth["username"] = v
					}
					if v := rw.BasicAuth.Password.ValueString(); v != "" {
						basicAuth["password"] = v
					}
					if len(basicAuth) > 0 {
						entry["basic_auth"] = basicAuth
					}
				}
				if rw.TLSConfig != nil {
					tlsConfig := map[string]interface{}{}
					if v := rw.TLSConfig.CAFile.ValueString(); v != "" {
						tlsConfig["ca_file"] = v
					}
					if v := rw.TLSConfig.CertFile.ValueString(); v != "" {
						tlsConfig["cert_file"] = v
					}
					if v := rw.TLSConfig.KeyFile.ValueString(); v != "" {
						tlsConfig["key_file"] = v
					}
					if v := rw.TLSConfig.ServerName.ValueString(); v != "" {
						tlsConfig["server_name"] = v
					}
					if rw.TLSConfig.InsecureSkipVerify.ValueBool() {
						tlsConfig["insecure_skip_verify"] = true
					}
					if len(tlsConfig) > 0 {
						entry["tls_config"] = tlsConfig
					}
				}
				remoteWrites = append(remoteWrites, entry)
			}
			deployReq["remote_write"] = remoteWrites
		}
	}

	return deployReq, diags
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
	} else {
		// Set to empty string if not provided
		data.Status = types.StringValue("")
	}

	if startedAt, ok := status["started_at"].(string); ok {
		data.StartedAt = types.StringValue(startedAt)
	} else {
		// Set to empty string if not provided (e.g., when Prometheus is starting)
		data.StartedAt = types.StringValue("")
	}

	if deploymentMode, ok := status["deployment_mode"].(string); ok {
		data.DeploymentMode = types.StringValue(deploymentMode)
	}

	if networkMode, ok := status["network_mode"].(string); ok {
		data.NetworkMode = types.StringValue(networkMode)
	}

	// remote_write, retention_time and retention_size are write-only/launch-flag
	// settings that the /metrics/status endpoint does not echo back (credentials
	// are never returned at all). They are therefore preserved as-is from the
	// existing state to avoid a perpetual diff or "inconsistent result" errors.
	// data.RetentionTime, data.RetentionSize and data.RemoteWrite are left
	// untouched here.

	return nil
}
