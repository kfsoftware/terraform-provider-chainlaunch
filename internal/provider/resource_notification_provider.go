package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &NotificationProviderResource{}
var _ resource.ResourceWithImportState = &NotificationProviderResource{}

func NewNotificationProviderResource() resource.Resource {
	return &NotificationProviderResource{}
}

type NotificationProviderResource struct {
	client *Client
}

type NotificationProviderResourceModel struct {
	ID                  types.String     `tfsdk:"id"`
	Name                types.String     `tfsdk:"name"`
	Type                types.String     `tfsdk:"type"`
	IsDefault           types.Bool       `tfsdk:"is_default"`
	NotifyBackupSuccess types.Bool       `tfsdk:"notify_backup_success"`
	NotifyBackupFailure types.Bool       `tfsdk:"notify_backup_failure"`
	NotifyNodeDowntime  types.Bool       `tfsdk:"notify_node_downtime"`
	NotifyS3ConnIssue   types.Bool       `tfsdk:"notify_s3_conn_issue"`
	SMTPConfig          *SMTPConfigModel `tfsdk:"smtp_config"`
	CreatedAt           types.String     `tfsdk:"created_at"`
	LastTestAt          types.String     `tfsdk:"last_test_at"`
	LastTestStatus      types.String     `tfsdk:"last_test_status"`
	LastTestMessage     types.String     `tfsdk:"last_test_message"`
}

type SMTPConfigModel struct {
	Host          types.String `tfsdk:"host"`
	Port          types.Int64  `tfsdk:"port"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	FromEmail     types.String `tfsdk:"from_email"`
	FromName      types.String `tfsdk:"from_name"`
	ToEmail       types.String `tfsdk:"to_email"`
	UseTLS        types.Bool   `tfsdk:"use_tls"`
	SkipTLSVerify types.Bool   `tfsdk:"skip_tls_verify"`
}

func (r *NotificationProviderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_provider"
}

func (r *NotificationProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages notification providers for alerting and monitoring.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Notification provider ID",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the notification provider",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of notification provider (currently only 'SMTP' is supported)",
			},
			"is_default": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether this is the default notification provider",
			},
			"notify_backup_success": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Send notifications on successful backups",
			},
			"notify_backup_failure": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Send notifications on backup failures",
			},
			"notify_node_downtime": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Send notifications on node downtime",
			},
			"notify_s3_conn_issue": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Send notifications on S3 connection issues",
			},
			"smtp_config": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "SMTP configuration (required when type is 'SMTP')",
				Attributes: map[string]schema.Attribute{
					"host": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "SMTP server hostname or IP address",
					},
					"port": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						Default:             int64default.StaticInt64(587),
						MarkdownDescription: "SMTP server port (default: 587)",
					},
					"username": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "SMTP authentication username",
					},
					"password": schema.StringAttribute{
						Optional:            true,
						Sensitive:           true,
						MarkdownDescription: "SMTP authentication password",
					},
					"from_email": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Sender email address",
					},
					"from_name": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("Chainlaunch Notifications"),
						MarkdownDescription: "Sender display name",
					},
					"to_email": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "Recipient email address",
					},
					"use_tls": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
						MarkdownDescription: "Use TLS for SMTP connection",
					},
					"skip_tls_verify": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Skip TLS certificate verification (use for self-signed certificates)",
					},
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Creation timestamp",
			},
			"last_test_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last test timestamp",
			},
			"last_test_status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last test status (success/failure)",
			},
			"last_test_message": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Last test message",
			},
		},
	}
}

func (r *NotificationProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NotificationProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NotificationProviderResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build config based on type
	var config map[string]interface{}
	if data.Type.ValueString() == "SMTP" && data.SMTPConfig != nil {
		config = map[string]interface{}{
			"host":            data.SMTPConfig.Host.ValueString(),
			"port":            data.SMTPConfig.Port.ValueInt64(),
			"from_email":      data.SMTPConfig.FromEmail.ValueString(),
			"from_name":       data.SMTPConfig.FromName.ValueString(),
			"to_email":        data.SMTPConfig.ToEmail.ValueString(),
			"use_tls":         data.SMTPConfig.UseTLS.ValueBool(),
			"skip_tls_verify": data.SMTPConfig.SkipTLSVerify.ValueBool(),
		}

		if !data.SMTPConfig.Username.IsNull() {
			config["username"] = data.SMTPConfig.Username.ValueString()
		}
		if !data.SMTPConfig.Password.IsNull() {
			config["password"] = data.SMTPConfig.Password.ValueString()
		}
	} else {
		resp.Diagnostics.AddError("Invalid Configuration", "SMTP config is required when type is 'SMTP'")
		return
	}

	createReq := map[string]interface{}{
		"name":                  data.Name.ValueString(),
		"type":                  data.Type.ValueString(),
		"is_default":            data.IsDefault.ValueBool(),
		"notify_backup_success": data.NotifyBackupSuccess.ValueBool(),
		"notify_backup_failure": data.NotifyBackupFailure.ValueBool(),
		"notify_node_downtime":  data.NotifyNodeDowntime.ValueBool(),
		"notify_s3_conn_issue":  data.NotifyS3ConnIssue.ValueBool(),
		"config":                config,
	}

	body, err := r.client.DoRequest("POST", "/notifications/providers", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create notification provider, got error: %s", err))
		return
	}

	var providerResp map[string]interface{}
	if err := json.Unmarshal(body, &providerResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse provider response: %s", err))
		return
	}

	// Set ID and other computed fields
	if id, ok := providerResp["id"].(float64); ok {
		data.ID = types.StringValue(fmt.Sprintf("%d", int64(id)))
	}

	if createdAt, ok := providerResp["createdAt"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}

	// Set last_test fields to empty if not present (they may be null on creation)
	if lastTestAt, ok := providerResp["lastTestAt"].(string); ok && lastTestAt != "" {
		data.LastTestAt = types.StringValue(lastTestAt)
	} else {
		data.LastTestAt = types.StringValue("")
	}

	if lastTestStatus, ok := providerResp["lastTestStatus"].(string); ok && lastTestStatus != "" {
		data.LastTestStatus = types.StringValue(lastTestStatus)
	} else {
		data.LastTestStatus = types.StringValue("")
	}

	if lastTestMessage, ok := providerResp["lastTestMessage"].(string); ok && lastTestMessage != "" {
		data.LastTestMessage = types.StringValue(lastTestMessage)
	} else {
		data.LastTestMessage = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationProviderResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.DoRequest("GET", fmt.Sprintf("/notifications/providers/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read notification provider, got error: %s", err))
		return
	}

	var providerResp map[string]interface{}
	if err := json.Unmarshal(body, &providerResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse provider response: %s", err))
		return
	}

	// Update fields from response
	if name, ok := providerResp["name"].(string); ok {
		data.Name = types.StringValue(name)
	}

	if providerType, ok := providerResp["type"].(string); ok {
		data.Type = types.StringValue(providerType)
	}

	if isDefault, ok := providerResp["isDefault"].(bool); ok {
		data.IsDefault = types.BoolValue(isDefault)
	}

	if notifyBackupSuccess, ok := providerResp["notifyBackupSuccess"].(bool); ok {
		data.NotifyBackupSuccess = types.BoolValue(notifyBackupSuccess)
	}

	if notifyBackupFailure, ok := providerResp["notifyBackupFailure"].(bool); ok {
		data.NotifyBackupFailure = types.BoolValue(notifyBackupFailure)
	}

	if notifyNodeDowntime, ok := providerResp["notifyNodeDowntime"].(bool); ok {
		data.NotifyNodeDowntime = types.BoolValue(notifyNodeDowntime)
	}

	if notifyS3ConnIssue, ok := providerResp["notifyS3ConnIssue"].(bool); ok {
		data.NotifyS3ConnIssue = types.BoolValue(notifyS3ConnIssue)
	}

	if createdAt, ok := providerResp["createdAt"].(string); ok {
		data.CreatedAt = types.StringValue(createdAt)
	}

	if lastTestAt, ok := providerResp["lastTestAt"].(string); ok && lastTestAt != "" {
		data.LastTestAt = types.StringValue(lastTestAt)
	}

	if lastTestStatus, ok := providerResp["lastTestStatus"].(string); ok && lastTestStatus != "" {
		data.LastTestStatus = types.StringValue(lastTestStatus)
	}

	if lastTestMessage, ok := providerResp["lastTestMessage"].(string); ok && lastTestMessage != "" {
		data.LastTestMessage = types.StringValue(lastTestMessage)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NotificationProviderResourceModel
	var state NotificationProviderResourceModel

	// Get current state to preserve computed fields
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve created_at from state
	data.CreatedAt = state.CreatedAt

	// Build config based on type
	var config map[string]interface{}
	if data.Type.ValueString() == "SMTP" && data.SMTPConfig != nil {
		config = map[string]interface{}{
			"host":            data.SMTPConfig.Host.ValueString(),
			"port":            data.SMTPConfig.Port.ValueInt64(),
			"from_email":      data.SMTPConfig.FromEmail.ValueString(),
			"from_name":       data.SMTPConfig.FromName.ValueString(),
			"to_email":        data.SMTPConfig.ToEmail.ValueString(),
			"use_tls":         data.SMTPConfig.UseTLS.ValueBool(),
			"skip_tls_verify": data.SMTPConfig.SkipTLSVerify.ValueBool(),
		}

		if !data.SMTPConfig.Username.IsNull() {
			config["username"] = data.SMTPConfig.Username.ValueString()
		}
		if !data.SMTPConfig.Password.IsNull() {
			config["password"] = data.SMTPConfig.Password.ValueString()
		}
	} else {
		resp.Diagnostics.AddError("Invalid Configuration", "SMTP config is required when type is 'SMTP'")
		return
	}

	updateReq := map[string]interface{}{
		"name":                  data.Name.ValueString(),
		"type":                  data.Type.ValueString(),
		"is_default":            data.IsDefault.ValueBool(),
		"notify_backup_success": data.NotifyBackupSuccess.ValueBool(),
		"notify_backup_failure": data.NotifyBackupFailure.ValueBool(),
		"notify_node_downtime":  data.NotifyNodeDowntime.ValueBool(),
		"notify_s3_conn_issue":  data.NotifyS3ConnIssue.ValueBool(),
		"config":                config,
	}

	body, err := r.client.DoRequest("PUT", fmt.Sprintf("/notifications/providers/%s", data.ID.ValueString()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update notification provider, got error: %s", err))
		return
	}

	var providerResp map[string]interface{}
	if err := json.Unmarshal(body, &providerResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse provider response: %s", err))
		return
	}

	// Update computed fields
	if lastTestAt, ok := providerResp["lastTestAt"].(string); ok && lastTestAt != "" {
		data.LastTestAt = types.StringValue(lastTestAt)
	}

	if lastTestStatus, ok := providerResp["lastTestStatus"].(string); ok && lastTestStatus != "" {
		data.LastTestStatus = types.StringValue(lastTestStatus)
	}

	if lastTestMessage, ok := providerResp["lastTestMessage"].(string); ok && lastTestMessage != "" {
		data.LastTestMessage = types.StringValue(lastTestMessage)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NotificationProviderResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DoRequest("DELETE", fmt.Sprintf("/notifications/providers/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete notification provider, got error: %s", err))
		return
	}
}

func (r *NotificationProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
