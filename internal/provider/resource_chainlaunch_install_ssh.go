package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"golang.org/x/crypto/ssh"
)

var _ resource.Resource = &ChainlaunchInstallSSHResource{}
var _ resource.ResourceWithImportState = &ChainlaunchInstallSSHResource{}

func NewChainlaunchInstallSSHResource() resource.Resource {
	return &ChainlaunchInstallSSHResource{}
}

type ChainlaunchInstallSSHResource struct{}

type ChainlaunchInstallSSHResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Host             types.String `tfsdk:"host"`
	Port             types.Int64  `tfsdk:"port"`
	User             types.String `tfsdk:"user"`
	Password         types.String `tfsdk:"password"`
	PrivateKey       types.String `tfsdk:"private_key"`
	Version          types.String `tfsdk:"version"`
	InstallPath      types.String `tfsdk:"install_path"`
	DataPath         types.String `tfsdk:"data_path"`
	Port8100         types.Int64  `tfsdk:"port_8100"`
	Environment      types.Map    `tfsdk:"environment"`
	AutoStart        types.Bool   `tfsdk:"auto_start"`
	ServiceStatus    types.String `tfsdk:"service_status"`
	ChainlaunchURL   types.String `tfsdk:"chainlaunch_url"`
	InstalledVersion types.String `tfsdk:"installed_version"`
}

func (r *ChainlaunchInstallSSHResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_install_ssh"
}

func (r *ChainlaunchInstallSSHResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages Chainlaunch installation as a systemd service on a remote Linux machine via SSH",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier (host:port)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host": schema.StringAttribute{
				MarkdownDescription: "SSH host address (IP or hostname)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "SSH port (default: 22)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"user": schema.StringAttribute{
				MarkdownDescription: "SSH username",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "SSH password (use private_key instead for better security)",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "SSH private key content (PEM format)",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Chainlaunch version to install (default: latest)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("latest"),
			},
			"install_path": schema.StringAttribute{
				MarkdownDescription: "Installation directory (default: /opt/chainlaunch)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("/opt/chainlaunch"),
			},
			"data_path": schema.StringAttribute{
				MarkdownDescription: "Data directory for Chainlaunch (default: /var/lib/chainlaunch)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("/var/lib/chainlaunch"),
			},
			"port_8100": schema.Int64Attribute{
				MarkdownDescription: "Chainlaunch API port (default: 8100)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"environment": schema.MapAttribute{
				MarkdownDescription: "Environment variables for Chainlaunch service",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"auto_start": schema.BoolAttribute{
				MarkdownDescription: "Enable service auto-start on boot (default: true)",
				Optional:            true,
				Computed:            true,
			},
			"service_status": schema.StringAttribute{
				MarkdownDescription: "Current systemd service status",
				Computed:            true,
			},
			"chainlaunch_url": schema.StringAttribute{
				MarkdownDescription: "URL to access Chainlaunch (http://host:port_8100)",
				Computed:            true,
			},
			"installed_version": schema.StringAttribute{
				MarkdownDescription: "Currently installed Chainlaunch version",
				Computed:            true,
			},
		},
	}
}

func (r *ChainlaunchInstallSSHResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ChainlaunchInstallSSHResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults
	if data.Port.IsNull() {
		data.Port = types.Int64Value(22)
	}
	if data.Port8100.IsNull() {
		data.Port8100 = types.Int64Value(8100)
	}
	if data.AutoStart.IsNull() {
		data.AutoStart = types.BoolValue(true)
	}

	// Establish SSH connection
	client, err := r.connectSSH(&data)
	if err != nil {
		resp.Diagnostics.AddError("SSH Connection Failed", fmt.Sprintf("Could not connect to %s: %s", data.Host.ValueString(), err.Error()))
		return
	}
	defer client.Close()

	// Install Chainlaunch
	if err := r.installChainlaunch(client, &data); err != nil {
		resp.Diagnostics.AddError("Installation Failed", fmt.Sprintf("Could not install Chainlaunch: %s", err.Error()))
		return
	}

	// Create systemd service
	if err := r.createSystemdService(client, &data); err != nil {
		resp.Diagnostics.AddError("Service Creation Failed", fmt.Sprintf("Could not create systemd service: %s", err.Error()))
		return
	}

	// Start service
	if err := r.startService(client); err != nil {
		resp.Diagnostics.AddError("Service Start Failed", fmt.Sprintf("Could not start service: %s", err.Error()))
		return
	}

	// Get service status
	status, err := r.getServiceStatus(client)
	if err != nil {
		resp.Diagnostics.AddWarning("Status Check Failed", fmt.Sprintf("Could not get service status: %s", err.Error()))
		status = "unknown"
	}
	data.ServiceStatus = types.StringValue(status)

	// Get installed version
	version, err := r.getInstalledVersion(client, &data)
	if err != nil {
		resp.Diagnostics.AddWarning("Version Check Failed", fmt.Sprintf("Could not get installed version: %s", err.Error()))
		version = data.Version.ValueString()
	}
	data.InstalledVersion = types.StringValue(version)

	// Set computed fields
	data.ID = types.StringValue(fmt.Sprintf("%s:%d", data.Host.ValueString(), data.Port.ValueInt64()))
	data.ChainlaunchURL = types.StringValue(fmt.Sprintf("http://%s:%d", data.Host.ValueString(), data.Port8100.ValueInt64()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ChainlaunchInstallSSHResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ChainlaunchInstallSSHResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Establish SSH connection
	client, err := r.connectSSH(&data)
	if err != nil {
		resp.Diagnostics.AddError("SSH Connection Failed", fmt.Sprintf("Could not connect to %s: %s", data.Host.ValueString(), err.Error()))
		return
	}
	defer client.Close()

	// Get service status
	status, err := r.getServiceStatus(client)
	if err != nil {
		// If service doesn't exist, remove from state
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "could not be found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddWarning("Status Check Failed", fmt.Sprintf("Could not get service status: %s", err.Error()))
		status = "unknown"
	}
	data.ServiceStatus = types.StringValue(status)

	// Get installed version
	version, err := r.getInstalledVersion(client, &data)
	if err != nil {
		resp.Diagnostics.AddWarning("Version Check Failed", fmt.Sprintf("Could not get installed version: %s", err.Error()))
	} else {
		data.InstalledVersion = types.StringValue(version)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ChainlaunchInstallSSHResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ChainlaunchInstallSSHResourceModel
	var state ChainlaunchInstallSSHResourceModel

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

	// Establish SSH connection
	client, err := r.connectSSH(&data)
	if err != nil {
		resp.Diagnostics.AddError("SSH Connection Failed", fmt.Sprintf("Could not connect to %s: %s", data.Host.ValueString(), err.Error()))
		return
	}
	defer client.Close()

	// Check if version changed
	if !data.Version.Equal(state.Version) {
		// Stop service
		if err := r.stopService(client); err != nil {
			resp.Diagnostics.AddWarning("Service Stop Failed", fmt.Sprintf("Could not stop service: %s", err.Error()))
		}

		// Reinstall with new version
		if err := r.installChainlaunch(client, &data); err != nil {
			resp.Diagnostics.AddError("Installation Failed", fmt.Sprintf("Could not install Chainlaunch: %s", err.Error()))
			return
		}
	}

	// Update systemd service if needed
	if err := r.createSystemdService(client, &data); err != nil {
		resp.Diagnostics.AddError("Service Update Failed", fmt.Sprintf("Could not update systemd service: %s", err.Error()))
		return
	}

	// Restart service
	if err := r.restartService(client); err != nil {
		resp.Diagnostics.AddError("Service Restart Failed", fmt.Sprintf("Could not restart service: %s", err.Error()))
		return
	}

	// Get service status
	status, err := r.getServiceStatus(client)
	if err != nil {
		resp.Diagnostics.AddWarning("Status Check Failed", fmt.Sprintf("Could not get service status: %s", err.Error()))
		status = "unknown"
	}
	data.ServiceStatus = types.StringValue(status)

	// Get installed version
	version, err := r.getInstalledVersion(client, &data)
	if err != nil {
		resp.Diagnostics.AddWarning("Version Check Failed", fmt.Sprintf("Could not get installed version: %s", err.Error()))
		version = data.Version.ValueString()
	}
	data.InstalledVersion = types.StringValue(version)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ChainlaunchInstallSSHResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ChainlaunchInstallSSHResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Establish SSH connection
	client, err := r.connectSSH(&data)
	if err != nil {
		resp.Diagnostics.AddError("SSH Connection Failed", fmt.Sprintf("Could not connect to %s: %s", data.Host.ValueString(), err.Error()))
		return
	}
	defer client.Close()

	// Stop and disable service
	if err := r.stopService(client); err != nil {
		resp.Diagnostics.AddWarning("Service Stop Failed", fmt.Sprintf("Could not stop service: %s", err.Error()))
	}

	if err := r.disableService(client); err != nil {
		resp.Diagnostics.AddWarning("Service Disable Failed", fmt.Sprintf("Could not disable service: %s", err.Error()))
	}

	// Remove systemd service file
	if err := r.removeSystemdService(client); err != nil {
		resp.Diagnostics.AddWarning("Service Removal Failed", fmt.Sprintf("Could not remove service file: %s", err.Error()))
	}

	// Optionally remove installation (commented out for safety)
	// if err := r.runCommand(client, fmt.Sprintf("sudo rm -rf %s", data.InstallPath.ValueString())); err != nil {
	// 	resp.Diagnostics.AddWarning("Cleanup Failed", fmt.Sprintf("Could not remove installation: %s", err.Error()))
	// }
}

func (r *ChainlaunchInstallSSHResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// SSH helper functions

func (r *ChainlaunchInstallSSHResource) connectSSH(data *ChainlaunchInstallSSHResourceModel) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		User:            data.User.ValueString(),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Make this configurable for production
		Timeout:         30 * time.Second,
	}

	// Use private key or password
	if !data.PrivateKey.IsNull() && data.PrivateKey.ValueString() != "" {
		signer, err := ssh.ParsePrivateKey([]byte(data.PrivateKey.ValueString()))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else if !data.Password.IsNull() && data.Password.ValueString() != "" {
		config.Auth = []ssh.AuthMethod{ssh.Password(data.Password.ValueString())}
	} else {
		return nil, fmt.Errorf("either password or private_key must be provided")
	}

	address := fmt.Sprintf("%s:%d", data.Host.ValueString(), data.Port.ValueInt64())
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return client, nil
}

func (r *ChainlaunchInstallSSHResource) runCommand(client *ssh.Client, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return fmt.Errorf("command failed: %s\nOutput: %s", err.Error(), string(output))
	}

	return nil
}

func (r *ChainlaunchInstallSSHResource) runCommandOutput(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("command failed: %s\nOutput: %s", err.Error(), string(output))
	}

	return strings.TrimSpace(string(output)), nil
}

func (r *ChainlaunchInstallSSHResource) installChainlaunch(client *ssh.Client, data *ChainlaunchInstallSSHResourceModel) error {
	installPath := data.InstallPath.ValueString()
	dataPath := data.DataPath.ValueString()
	version := data.Version.ValueString()

	// Create directories
	if err := r.runCommand(client, fmt.Sprintf("sudo mkdir -p %s", installPath)); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}
	if err := r.runCommand(client, fmt.Sprintf("sudo mkdir -p %s", dataPath)); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Download Chainlaunch binary
	downloadURL := fmt.Sprintf("https://github.com/kfsoftware/chainlaunch/releases/download/%s/chainlaunch-linux-amd64", version)
	if version == "latest" {
		downloadURL = "https://github.com/kfsoftware/chainlaunch/releases/latest/download/chainlaunch-linux-amd64"
	}

	downloadCmd := fmt.Sprintf(`
		sudo curl -L -o %s/chainlaunch %s && \
		sudo chmod +x %s/chainlaunch
	`, installPath, downloadURL, installPath)

	if err := r.runCommand(client, downloadCmd); err != nil {
		return fmt.Errorf("failed to download Chainlaunch: %w", err)
	}

	return nil
}

func (r *ChainlaunchInstallSSHResource) createSystemdService(client *ssh.Client, data *ChainlaunchInstallSSHResourceModel) error {
	installPath := data.InstallPath.ValueString()
	dataPath := data.DataPath.ValueString()
	port := data.Port8100.ValueInt64()

	// Build environment variables
	envVars := ""
	if !data.Environment.IsNull() {
		envMap := make(map[string]string)
		data.Environment.ElementsAs(context.Background(), &envMap, false)
		for key, value := range envMap {
			envVars += fmt.Sprintf("Environment=\"%s=%s\"\n", key, value)
		}
	}

	serviceContent := fmt.Sprintf(`[Unit]
Description=Chainlaunch Blockchain Platform
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=root
WorkingDirectory=%s
ExecStart=%s/chainlaunch server --port %d --data-dir %s
Restart=always
RestartSec=10
%s
[Install]
WantedBy=multi-user.target
`, dataPath, installPath, port, dataPath, envVars)

	// Write service file
	writeCmd := fmt.Sprintf(`
		echo '%s' | sudo tee /etc/systemd/system/chainlaunch.service > /dev/null && \
		sudo systemctl daemon-reload
	`, serviceContent)

	if err := r.runCommand(client, writeCmd); err != nil {
		return fmt.Errorf("failed to create service file: %w", err)
	}

	// Enable service if auto_start is true
	if data.AutoStart.ValueBool() {
		if err := r.runCommand(client, "sudo systemctl enable chainlaunch"); err != nil {
			return fmt.Errorf("failed to enable service: %w", err)
		}
	}

	return nil
}

func (r *ChainlaunchInstallSSHResource) startService(client *ssh.Client) error {
	return r.runCommand(client, "sudo systemctl start chainlaunch")
}

func (r *ChainlaunchInstallSSHResource) stopService(client *ssh.Client) error {
	return r.runCommand(client, "sudo systemctl stop chainlaunch")
}

func (r *ChainlaunchInstallSSHResource) restartService(client *ssh.Client) error {
	return r.runCommand(client, "sudo systemctl restart chainlaunch")
}

func (r *ChainlaunchInstallSSHResource) disableService(client *ssh.Client) error {
	return r.runCommand(client, "sudo systemctl disable chainlaunch")
}

func (r *ChainlaunchInstallSSHResource) removeSystemdService(client *ssh.Client) error {
	return r.runCommand(client, `
		sudo systemctl stop chainlaunch 2>/dev/null || true && \
		sudo systemctl disable chainlaunch 2>/dev/null || true && \
		sudo rm -f /etc/systemd/system/chainlaunch.service && \
		sudo systemctl daemon-reload
	`)
}

func (r *ChainlaunchInstallSSHResource) getServiceStatus(client *ssh.Client) (string, error) {
	output, err := r.runCommandOutput(client, "sudo systemctl is-active chainlaunch")
	if err != nil {
		// Check if service exists
		_, checkErr := r.runCommandOutput(client, "sudo systemctl status chainlaunch")
		if checkErr != nil && strings.Contains(checkErr.Error(), "not found") {
			return "", fmt.Errorf("service not found")
		}
		return "inactive", nil
	}
	return output, nil
}

func (r *ChainlaunchInstallSSHResource) getInstalledVersion(client *ssh.Client, data *ChainlaunchInstallSSHResourceModel) (string, error) {
	installPath := data.InstallPath.ValueString()
	output, err := r.runCommandOutput(client, fmt.Sprintf("%s/chainlaunch version 2>/dev/null || echo 'unknown'", installPath))
	if err != nil {
		return "unknown", err
	}
	return output, nil
}
