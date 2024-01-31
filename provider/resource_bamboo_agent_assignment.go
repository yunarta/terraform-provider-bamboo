package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
)

var (
	_ resource.Resource              = &AgentAssignmentResource{}
	_ resource.ResourceWithConfigure = &AgentAssignmentResource{}
	_ ProjectPermissionsReceiver     = &AgentAssignmentResource{}
	_ ConfigurableReceiver           = &AgentAssignmentResource{}
)

func NewAgentAssignmentResource() resource.Resource {
	return &AgentAssignmentResource{}
}

type AgentAssignmentResource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *AgentAssignmentResource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *AgentAssignmentResource) getClient() *bamboo.Client {
	return receiver.client
}

func (receiver *AgentAssignmentResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_agent_assignment"
}

func (receiver *AgentAssignmentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"agent": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"deployment": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
			},
		},
	}
}

func (receiver *AgentAssignmentResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *AgentAssignmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics

		plan AgentAssignmentModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	err := receiver.client.AgentAssignmentService().Create(bamboo.AgentAssignment{
		ExecutorType:   plan.Type.ValueString(),
		ExecutorId:     plan.AgentId.ValueInt64(),
		EntityId:       plan.DeploymentId.ValueInt64(),
		AssignmentType: "DEPLOYMENT_PROJECT",
	})

	if util.TestError(&response.Diagnostics, err, "Failed to create assignment") {
		return
	}

	diags = response.State.Set(ctx, plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *AgentAssignmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics

		state AgentAssignmentModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = response.State.Set(ctx, state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *AgentAssignmentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags diag.Diagnostics

		plan, state AgentAssignmentModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	err := receiver.client.AgentAssignmentService().Create(bamboo.AgentAssignment{
		ExecutorType:   plan.Type.ValueString(),
		ExecutorId:     plan.AgentId.ValueInt64(),
		EntityId:       plan.DeploymentId.ValueInt64(),
		AssignmentType: "DEPLOYMENT_PROJECT",
	})

	if util.TestError(&response.Diagnostics, err, "Failed to update project variable") {
		return
	}

	diags = response.State.Set(ctx, plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *AgentAssignmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		state AgentAssignmentModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	err := receiver.client.AgentAssignmentService().Delete(bamboo.AgentAssignment{
		ExecutorType:   state.Type.ValueString(),
		ExecutorId:     state.AgentId.ValueInt64(),
		EntityId:       state.DeploymentId.ValueInt64(),
		AssignmentType: "DEPLOYMENT_PROJECT",
	})
	if util.TestError(&response.Diagnostics, err, "Failed to delete assignment") {
		return
	}

	response.State.RemoveResource(ctx)
}
