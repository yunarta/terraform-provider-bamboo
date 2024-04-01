package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
		MarkdownDescription: `This resource define assignment of executable (project, plan, job, deployment, environment) to a Bamboo agent.
`,
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Numeric id of the assignment.",
			},
			"agent": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
				MarkdownDescription: "Numeric id of the agent.",
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("AGENT", "IMAGE", "EPHEMERAL"),
				},
				MarkdownDescription: "Agent type (AGENT, IMAGE - elastic EC2 agent, EPHEMERAL - K8S agent).",
			},
			"executable_id": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplaceIfConfigured(),
				},
				MarkdownDescription: "Numeric id of the executable. As per current only deployment project is usable as i dont have data source for other type yet.",
			},
			"executable_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				Validators: []validator.String{
					//stringvalidator.OneOf("DEPLOYMENT_PROJECT"),
					stringvalidator.OneOf("PROJECT", "PLAN", "JOB", "DEPLOYMENT_PROJECT", "ENVIRONMENT"),
				},
				MarkdownDescription: "Executable type.",
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

	err := receiver.client.AgentAssignmentService().Create(bamboo.AgentAssignmentRequest{
		ExecutorType:   plan.Type,
		ExecutorId:     plan.AgentId,
		EntityId:       plan.ExecutableId,
		AssignmentType: plan.ExecutableType,
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
	//var (
	//	diags diag.Diagnostics
	//	err error
	//
	//	state AgentAssignmentModel
	//)
	//
	//diags = request.State.Get(ctx, &state)
	//if util.TestDiagnostic(&response.Diagnostics, diags) {
	//	return
	//}
	//
	//diags = response.State.Set(ctx, state)
	//if util.TestDiagnostic(&response.Diagnostics, diags) {
	//	return
	//}
	//
	//cacheDir := filepath.Join(".cache", "agent")
	//_ = os.MkdirAll(cacheDir, 0755)
	//filename := filepath.Join(cacheDir, fmt.Sprintf("%d.json", state.AgentId))
	//
	//var assignmentList *[]bamboo.AgentAssignment
	//if _, err := os.Stat(filename); err == nil {
	//	fileBytes, err := os.ReadFile(filename)
	//	if err != nil {
	//		err = json.Unmarshal(fileBytes, assignmentList)
	//	}
	//}
	//
	//if assignmentList == nil {
	//	assignmentList, err = receiver.client.AgentAssignmentService().Read(bamboo.AgentQuery{
	//		ExecutorType: state.Type,
	//		ExecutorId:   state.AgentId,
	//	})
	//	if util.TestError(&response.Diagnostics, err, "Failed to retrieve agent assignments") {
	//		return
	//	}
	//
	//	marshal, _ := json.Marshal(assignmentList)
	//	_ = os.WriteFile(filename, marshal, 0644)
	//}
	//
	//assignmentList
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

	err := receiver.client.AgentAssignmentService().Create(bamboo.AgentAssignmentRequest{
		ExecutorType:   plan.Type,
		ExecutorId:     plan.AgentId,
		EntityId:       plan.ExecutableId,
		AssignmentType: plan.ExecutableType,
	})

	if util.TestError(&response.Diagnostics, err, "Failed to update agent assignment") {
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

	err := receiver.client.AgentAssignmentService().Delete(bamboo.AgentAssignmentRequest{
		ExecutorType:   state.Type,
		ExecutorId:     state.AgentId,
		EntityId:       state.ExecutableId,
		AssignmentType: state.ExecutableType,
	})
	if util.TestError(&response.Diagnostics, err, "Failed to delete assignment") {
		return
	}

	response.State.RemoveResource(ctx)
}
