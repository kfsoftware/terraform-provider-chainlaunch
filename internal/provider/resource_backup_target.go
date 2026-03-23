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
	UseInstanceRole types.Bool   `tfsdk:"use_instance_role"`
	RoleARN         types.String `tfsdk:"role_arn"`
	ExternalID      types.String `tfsdk:"external_id"`
	Profile         types.String `tfsdk:"profile"`
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
				Optional:    true,
				Sensitive:   true,
				Description: "S3 access key ID. Required when using static credentials.",
			},
			"secret_access_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "S3 secret access key. Required when using static credentials.",
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
			"use_instance_role": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Use EC2 instance role or EKS IRSA for authentication instead of static credentials.",
			},
			"role_arn": schema.StringAttribute{
				Optional:    true,
				Description: "AWS IAM role ARN to assume for S3 access. Can be combined with static credentials or instance role as base credentials.",
			},
			"external_id": schema.StringAttribute{
				Optional:    true,
				Description: "External ID for STS AssumeRole (used with role_arn for cross-account access).",
			},
			"profile": schema.StringAttribute{
				Optional:    true,
				Description: "AWS named profile from ~/.aws/credentials or ~/.aws/config.",
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

	createReq := backupTargetConfigRequest(data)

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
	populateBackupTargetFromResponse(&data, &target)

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

	populateBackupTargetFromResponse(&data, &target)

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

	updateReq := backupTargetConfigRequest(data)

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

	populateBackupTargetFromResponse(&data, &target)

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

// populateBackupTargetFromResponse sets model fields from the API response.
// It reads from both legacy top-level fields and the config JSON object.
func populateBackupTargetFromResponse(data *BackupTargetResourceModel, target *BackupTarget) {
	data.Name = types.StringValue(target.Name)
	data.Type = types.StringValue(target.Type)

	// Prefer config JSON values over legacy top-level fields
	cfg := target.Config

	region := target.Region
	if v, ok := cfg["region"].(string); ok && v != "" {
		region = v
	}
	data.Region = types.StringValue(region)

	endpoint := target.Endpoint
	if v, ok := cfg["endpoint"].(string); ok && v != "" {
		endpoint = v
	}
	if endpoint != "" {
		data.Endpoint = types.StringValue(endpoint)
	}

	bucketName := target.BucketName
	if v, ok := cfg["bucketName"].(string); ok && v != "" {
		bucketName = v
	}
	data.BucketName = types.StringValue(bucketName)

	bucketPath := target.BucketPath
	if v, ok := cfg["bucketPath"].(string); ok && v != "" {
		bucketPath = v
	}
	if bucketPath != "" {
		data.BucketPath = types.StringValue(bucketPath)
	}

	forcePathStyle := target.ForcePathStyle
	if v, ok := cfg["forcePathStyle"].(bool); ok {
		forcePathStyle = v
	}
	data.ForcePathStyle = types.BoolValue(forcePathStyle)

	// Auth fields from config — don't read sensitive values back from API
	// (access_key_id and secret_access_key are preserved from state by Terraform)
	if v, ok := cfg["useInstanceRole"].(bool); ok && v {
		data.UseInstanceRole = types.BoolValue(true)
	}
	if v, ok := cfg["roleArn"].(string); ok && v != "" {
		data.RoleARN = types.StringValue(v)
	}
	if v, ok := cfg["externalId"].(string); ok && v != "" {
		data.ExternalID = types.StringValue(v)
	}
	if v, ok := cfg["profile"].(string); ok && v != "" {
		data.Profile = types.StringValue(v)
	}

	if target.CreatedAt != "" {
		data.CreatedAt = types.StringValue(target.CreatedAt)
	}
	if target.UpdatedAt != "" {
		data.UpdatedAt = types.StringValue(target.UpdatedAt)
	}
}

// backupTargetConfigRequest builds a request body that uses the JSON config approach,
// supporting static credentials, instance role, role assumption, and named profile.
func backupTargetConfigRequest(data BackupTargetResourceModel) map[string]interface{} {
	config := map[string]interface{}{
		"region":         data.Region.ValueString(),
		"bucketName":     data.BucketName.ValueString(),
		"forcePathStyle": data.ForcePathStyle.ValueBool(),
		"resticPassword": data.ResticPassword.ValueString(),
	}
	if !data.Endpoint.IsNull() && data.Endpoint.ValueString() != "" {
		config["endpoint"] = data.Endpoint.ValueString()
	}
	if !data.BucketPath.IsNull() && data.BucketPath.ValueString() != "" {
		config["bucketPath"] = data.BucketPath.ValueString()
	}

	// Auth: static credentials
	if !data.AccessKeyID.IsNull() && data.AccessKeyID.ValueString() != "" {
		config["accessKeyId"] = data.AccessKeyID.ValueString()
	}
	if !data.SecretAccessKey.IsNull() && data.SecretAccessKey.ValueString() != "" {
		config["secretKey"] = data.SecretAccessKey.ValueString()
	}

	// Auth: instance role
	if !data.UseInstanceRole.IsNull() && data.UseInstanceRole.ValueBool() {
		config["useInstanceRole"] = true
	}

	// Auth: role assumption
	if !data.RoleARN.IsNull() && data.RoleARN.ValueString() != "" {
		config["roleArn"] = data.RoleARN.ValueString()
	}
	if !data.ExternalID.IsNull() && data.ExternalID.ValueString() != "" {
		config["externalId"] = data.ExternalID.ValueString()
	}

	// Auth: named profile
	if !data.Profile.IsNull() && data.Profile.ValueString() != "" {
		config["profile"] = data.Profile.ValueString()
	}

	return map[string]interface{}{
		"name":   data.Name.ValueString(),
		"type":   data.Type.ValueString(),
		"config": config,
	}
}
