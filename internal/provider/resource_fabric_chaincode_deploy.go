package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &FabricChaincodeDeployResource{}

func NewFabricChaincodeDeployResource() resource.Resource {
	return &FabricChaincodeDeployResource{}
}

type FabricChaincodeDeployResource struct {
	client *Client
}

type FabricChaincodeDeployResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	DefinitionID         types.Int64  `tfsdk:"definition_id"`
	EnvironmentVariables types.Map    `tfsdk:"environment_variables"`
	Status               types.String `tfsdk:"status"`
	Message              types.String `tfsdk:"message"`
}

func (r *FabricChaincodeDeployResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabric_chaincode_deploy"
}

func (r *FabricChaincodeDeployResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Deploys a Fabric chaincode by starting the Docker container based on the definition. This step starts the chaincode container using the Docker image specified in the definition. The chaincode must be installed and committed before deployment.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for this deployment (format: definition_id).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"definition_id": schema.Int64Attribute{
				Required:    true,
				Description: "The ID of the chaincode definition to deploy.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"environment_variables": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Environment variables to pass to the chaincode container.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The status of the deployment operation.",
			},
			"message": schema.StringAttribute{
				Computed:    true,
				Description: "Message from the deployment operation.",
			},
		},
	}
}

func (r *FabricChaincodeDeployResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FabricChaincodeDeployResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FabricChaincodeDeployResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Prepare environment variables
	var envVars map[string]string
	if !data.EnvironmentVariables.IsNull() {
		resp.Diagnostics.Append(data.EnvironmentVariables.ElementsAs(ctx, &envVars, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Deploy chaincode request
	deployReq := struct {
		EnvironmentVariables map[string]string `json:"environment_variables,omitempty"`
	}{
		EnvironmentVariables: envVars,
	}

	endpoint := fmt.Sprintf("/sc/fabric/definitions/%d/deploy", data.DefinitionID.ValueInt64())
	body, err := r.client.DoRequest("POST", endpoint, deployReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to deploy chaincode: %s", err))
		return
	}

	// Parse response (it returns a map[string]string)
	var deployResp map[string]string
	if err := json.Unmarshal(body, &deployResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse response: %s", err))
		return
	}

	// Set state
	data.ID = types.StringValue(fmt.Sprintf("%d", data.DefinitionID.ValueInt64()))

	// Extract status and message from response
	if status, ok := deployResp["status"]; ok {
		data.Status = types.StringValue(status)
	} else {
		data.Status = types.StringValue("success")
	}

	if message, ok := deployResp["message"]; ok {
		data.Message = types.StringValue(message)
	} else {
		data.Message = types.StringValue("Chaincode deployed successfully")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeDeployResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FabricChaincodeDeployResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API doesn't provide a GET endpoint to verify chaincode deployment status
	// We keep the state as-is

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeDeployResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data FabricChaincodeDeployResourceModel

	// Get plan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// All fields are marked as RequiresReplace, so this shouldn't be called
	// Including for consistency

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FabricChaincodeDeployResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FabricChaincodeDeployResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Undeploy the chaincode (stop the Docker container)
	endpoint := fmt.Sprintf("/sc/fabric/definitions/%d/undeploy", data.DefinitionID.ValueInt64())
	_, err := r.client.DoRequest("POST", endpoint, nil)
	if err != nil {
		resp.Diagnostics.AddWarning("Undeploy Warning", fmt.Sprintf("Unable to undeploy chaincode: %s. The container may still be running.", err))
		// Continue with state removal even if undeploy fails
	}
}
