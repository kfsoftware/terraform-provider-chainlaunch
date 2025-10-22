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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &BackupScheduleResource{}
var _ resource.ResourceWithImportState = &BackupScheduleResource{}

func NewBackupScheduleResource() resource.Resource {
	return &BackupScheduleResource{}
}

// BackupScheduleResource defines the resource implementation.
type BackupScheduleResource struct {
	client *Client
}

// BackupScheduleResourceModel describes the resource data model.
type BackupScheduleResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	TargetID       types.Int64  `tfsdk:"target_id"`
	CronExpression types.String `tfsdk:"cron_expression"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	RetentionDays  types.Int64  `tfsdk:"retention_days"`
	LastRunAt      types.String `tfsdk:"last_run_at"`
	NextRunAt      types.String `tfsdk:"next_run_at"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func (r *BackupScheduleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup_schedule"
}

func (r *BackupScheduleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an automated backup schedule for Chainlaunch backups.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the backup schedule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the backup schedule.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "A description of the backup schedule.",
			},
			"target_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the backup target where backups will be stored.",
			},
			"cron_expression": schema.StringAttribute{
				Required:    true,
				Description: "Cron expression defining when backups should run (e.g., '0 0 * * *' for daily at midnight).",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the backup schedule is enabled.",
			},
			"retention_days": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(30),
				Description: "Number of days to retain backups before automatic deletion.",
			},
			"last_run_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the schedule last executed a backup.",
			},
			"next_run_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the schedule will next execute a backup.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the backup schedule was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the backup schedule was last updated.",
			},
		},
	}
}

func (r *BackupScheduleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *BackupScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BackupScheduleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := CreateBackupScheduleRequest{
		Name:           data.Name.ValueString(),
		Description:    data.Description.ValueString(),
		TargetID:       int(data.TargetID.ValueInt64()),
		CronExpression: data.CronExpression.ValueString(),
		Enabled:        data.Enabled.ValueBool(),
		RetentionDays:  int(data.RetentionDays.ValueInt64()),
	}

	body, err := r.client.DoRequest("POST", "/backups/schedules", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create backup schedule, got error: %s", err))
		return
	}

	var schedule BackupSchedule
	if err := json.Unmarshal(body, &schedule); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse backup schedule response, got error: %s\nResponse body: %s", err, string(body)))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", schedule.ID))
	data.Name = types.StringValue(schedule.Name)
	if schedule.Description != "" {
		data.Description = types.StringValue(schedule.Description)
	}
	data.TargetID = types.Int64Value(int64(schedule.TargetID))
	data.CronExpression = types.StringValue(schedule.CronExpression)
	data.Enabled = types.BoolValue(schedule.Enabled)
	data.RetentionDays = types.Int64Value(int64(schedule.RetentionDays))
	if schedule.LastRunAt != "" {
		data.LastRunAt = types.StringValue(schedule.LastRunAt)
	}
	if schedule.NextRunAt != "" {
		data.NextRunAt = types.StringValue(schedule.NextRunAt)
	}
	if schedule.CreatedAt != "" {
		data.CreatedAt = types.StringValue(schedule.CreatedAt)
	}
	if schedule.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(schedule.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BackupScheduleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.DoRequest("GET", fmt.Sprintf("/backups/schedules/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read backup schedule, got error: %s", err))
		return
	}

	var schedule BackupSchedule
	if err := json.Unmarshal(body, &schedule); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse backup schedule response, got error: %s", err))
		return
	}

	data.Name = types.StringValue(schedule.Name)
	if schedule.Description != "" {
		data.Description = types.StringValue(schedule.Description)
	}
	data.TargetID = types.Int64Value(int64(schedule.TargetID))
	data.CronExpression = types.StringValue(schedule.CronExpression)
	data.Enabled = types.BoolValue(schedule.Enabled)
	data.RetentionDays = types.Int64Value(int64(schedule.RetentionDays))
	if schedule.LastRunAt != "" {
		data.LastRunAt = types.StringValue(schedule.LastRunAt)
	}
	if schedule.NextRunAt != "" {
		data.NextRunAt = types.StringValue(schedule.NextRunAt)
	}
	if schedule.CreatedAt != "" {
		data.CreatedAt = types.StringValue(schedule.CreatedAt)
	}
	if schedule.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(schedule.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BackupScheduleResourceModel
	var state BackupScheduleResourceModel

	// Get current state to preserve computed fields
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve created_at from state
	data.CreatedAt = state.CreatedAt

	updateReq := CreateBackupScheduleRequest{
		Name:           data.Name.ValueString(),
		Description:    data.Description.ValueString(),
		TargetID:       int(data.TargetID.ValueInt64()),
		CronExpression: data.CronExpression.ValueString(),
		Enabled:        data.Enabled.ValueBool(),
		RetentionDays:  int(data.RetentionDays.ValueInt64()),
	}

	body, err := r.client.DoRequest("PUT", fmt.Sprintf("/backups/schedules/%s", data.ID.ValueString()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update backup schedule, got error: %s", err))
		return
	}

	var schedule BackupSchedule
	if err := json.Unmarshal(body, &schedule); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse backup schedule response, got error: %s", err))
		return
	}

	data.Name = types.StringValue(schedule.Name)
	if schedule.Description != "" {
		data.Description = types.StringValue(schedule.Description)
	}
	data.TargetID = types.Int64Value(int64(schedule.TargetID))
	data.CronExpression = types.StringValue(schedule.CronExpression)
	data.Enabled = types.BoolValue(schedule.Enabled)
	data.RetentionDays = types.Int64Value(int64(schedule.RetentionDays))
	if schedule.LastRunAt != "" {
		data.LastRunAt = types.StringValue(schedule.LastRunAt)
	}
	if schedule.NextRunAt != "" {
		data.NextRunAt = types.StringValue(schedule.NextRunAt)
	}
	if schedule.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(schedule.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BackupScheduleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DoRequest("DELETE", fmt.Sprintf("/backups/schedules/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete backup schedule, got error: %s", err))
		return
	}
}

func (r *BackupScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
