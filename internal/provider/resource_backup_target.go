package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &BackupTargetResource{}
var _ resource.ResourceWithImportState = &BackupTargetResource{}

func NewBackupTargetResource() resource.Resource {
	return &BackupTargetResource{}
}

// BackupTargetResource defines the resource implementation.
type BackupTargetResource struct {
	client *Client
}

// BackupTargetResourceModel describes the resource data model.
type BackupTargetResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	Endpoint        types.String `tfsdk:"endpoint"`
	Region          types.String `tfsdk:"region"`
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	BucketName      types.String `tfsdk:"bucket_name"`
	BucketPath      types.String `tfsdk:"bucket_path"`
	ForcePathStyle  types.Bool   `tfsdk:"force_path_style"`
	ResticPassword  types.String `tfsdk:"restic_password"`
	CreatedAt       types.String `tfsdk:"created_at"`
	UpdatedAt       types.String `tfsdk:"updated_at"`
}

func (r *BackupTargetResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup_target"
}

func (r *BackupTargetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a backup target (S3-compatible storage) for Chainlaunch backups.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the backup target.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the backup target.",
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The type of backup target. Currently only 'S3' is supported.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"endpoint": schema.StringAttribute{
				Optional:    true,
				Description: "Custom S3 endpoint URL (e.g., for MinIO or other S3-compatible storage). Leave empty for AWS S3.",
			},
			"region": schema.StringAttribute{
				Required:    true,
				Description: "AWS region (e.g., 'us-east-1') or 'us-east-1' for MinIO.",
			},
			"access_key_id": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "S3 access key ID.",
			},
			"secret_access_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "S3 secret access key.",
			},
			"bucket_name": schema.StringAttribute{
				Required:    true,
				Description: "S3 bucket name where backups will be stored.",
			},
			"bucket_path": schema.StringAttribute{
				Optional:    true,
				Description: "Path within the bucket for backups (e.g., 'backups/fabric').",
			},
			"force_path_style": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Use path-style S3 URLs. Required for MinIO and some S3-compatible services.",
			},
			"restic_password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Password for encrypting backups with Restic.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the backup target was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the backup target was last updated.",
			},
		},
	}
}

func (r *BackupTargetResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BackupTargetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BackupTargetResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createReq := CreateBackupTargetRequest{
		Name:            data.Name.ValueString(),
		Type:            data.Type.ValueString(),
		Endpoint:        data.Endpoint.ValueString(),
		Region:          data.Region.ValueString(),
		AccessKeyID:     data.AccessKeyID.ValueString(),
		SecretAccessKey: data.SecretAccessKey.ValueString(),
		BucketName:      data.BucketName.ValueString(),
		BucketPath:      data.BucketPath.ValueString(),
		ForcePathStyle:  data.ForcePathStyle.ValueBool(),
		ResticPassword:  data.ResticPassword.ValueString(),
	}

	body, err := r.client.DoRequest("POST", "/backups/targets", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create backup target, got error: %s", err))
		return
	}

	var target BackupTarget
	if err := json.Unmarshal(body, &target); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse backup target response, got error: %s\nResponse body: %s", err, string(body)))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", target.ID))
	data.Name = types.StringValue(target.Name)
	data.Type = types.StringValue(target.Type)
	if target.Endpoint != "" {
		data.Endpoint = types.StringValue(target.Endpoint)
	}
	data.Region = types.StringValue(target.Region)
	data.BucketName = types.StringValue(target.BucketName)
	if target.BucketPath != "" {
		data.BucketPath = types.StringValue(target.BucketPath)
	}
	data.ForcePathStyle = types.BoolValue(target.ForcePathStyle)
	if target.CreatedAt != "" {
		data.CreatedAt = types.StringValue(target.CreatedAt)
	}
	if target.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(target.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupTargetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BackupTargetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	body, err := r.client.DoRequest("GET", fmt.Sprintf("/backups/targets/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read backup target, got error: %s", err))
		return
	}

	var target BackupTarget
	if err := json.Unmarshal(body, &target); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse backup target response, got error: %s", err))
		return
	}

	data.Name = types.StringValue(target.Name)
	data.Type = types.StringValue(target.Type)
	if target.Endpoint != "" {
		data.Endpoint = types.StringValue(target.Endpoint)
	}
	data.Region = types.StringValue(target.Region)
	data.BucketName = types.StringValue(target.BucketName)
	if target.BucketPath != "" {
		data.BucketPath = types.StringValue(target.BucketPath)
	}
	data.ForcePathStyle = types.BoolValue(target.ForcePathStyle)
	if target.CreatedAt != "" {
		data.CreatedAt = types.StringValue(target.CreatedAt)
	}
	if target.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(target.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupTargetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BackupTargetResourceModel
	var state BackupTargetResourceModel

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

	updateReq := CreateBackupTargetRequest{
		Name:            data.Name.ValueString(),
		Type:            data.Type.ValueString(),
		Endpoint:        data.Endpoint.ValueString(),
		Region:          data.Region.ValueString(),
		AccessKeyID:     data.AccessKeyID.ValueString(),
		SecretAccessKey: data.SecretAccessKey.ValueString(),
		BucketName:      data.BucketName.ValueString(),
		BucketPath:      data.BucketPath.ValueString(),
		ForcePathStyle:  data.ForcePathStyle.ValueBool(),
		ResticPassword:  data.ResticPassword.ValueString(),
	}

	body, err := r.client.DoRequest("PUT", fmt.Sprintf("/backups/targets/%s", data.ID.ValueString()), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update backup target, got error: %s", err))
		return
	}

	var target BackupTarget
	if err := json.Unmarshal(body, &target); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse backup target response, got error: %s", err))
		return
	}

	data.Name = types.StringValue(target.Name)
	data.Type = types.StringValue(target.Type)
	if target.Endpoint != "" {
		data.Endpoint = types.StringValue(target.Endpoint)
	}
	data.Region = types.StringValue(target.Region)
	data.BucketName = types.StringValue(target.BucketName)
	if target.BucketPath != "" {
		data.BucketPath = types.StringValue(target.BucketPath)
	}
	data.ForcePathStyle = types.BoolValue(target.ForcePathStyle)
	if target.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(target.UpdatedAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *BackupTargetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data BackupTargetResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DoRequest("DELETE", fmt.Sprintf("/backups/targets/%s", data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete backup target, got error: %s", err))
		return
	}
}

func (r *BackupTargetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
