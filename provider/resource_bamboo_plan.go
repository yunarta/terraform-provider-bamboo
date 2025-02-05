package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
	"strings"
)

var (
	_ resource.Resource                 = &PlanResource{}
	_ resource.ResourceWithConfigure    = &PlanResource{}
	_ resource.ResourceWithImportState  = &PlanResource{}
	_ resource.ResourceWithUpgradeState = &PlanResource{}
	_ ProjectPermissionsReceiver        = &PlanResource{}
	_ ConfigurableReceiver              = &PlanResource{}
)

func NewPlanResource() resource.Resource {
	return &PlanResource{}
}

type PlanResource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *PlanResource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *PlanResource) getClient() *bamboo.Client {
	return receiver.client
}

func (receiver *PlanResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_plan"
}

func (receiver *PlanResource) schemaV0() schema.Schema {
	return schema.Schema{
		MarkdownDescription: `This resource define project plan.

The priority block has a priority that defines the final assigned permissions of the user or group.`,
		Attributes: map[string]schema.Attribute{
			"retain_on_delete": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default value is `true`, and if the value set to `false` when the resource destroyed, the project will be removed.",
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Plan id.",
			},
			"key": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
				MarkdownDescription: "Project key.",
			},
			"plan_key": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
				MarkdownDescription: "Plan key.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project name.",
			},
		},
	}
}

func (receiver *PlanResource) schemaV1() schema.Schema {
	return schema.Schema{
		Version: 1,
		MarkdownDescription: `This resource define project plan.

The priority block has a priority that defines the final assigned permissions of the user or group.`,
		Attributes: map[string]schema.Attribute{
			"retain_on_delete": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default value is `true`, and if the value set to `false` when the resource destroyed, the project will be removed.",
			},
			"id": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Plan id.",
			},
			"project": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
				MarkdownDescription: "Project key.",
			},
			"plan_key": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
				MarkdownDescription: "Plan key.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project name.",
			},
		},
	}
}

func (receiver *PlanResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = receiver.schemaV1()
}

func (receiver *PlanResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	v0 := receiver.schemaV0()
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &v0,
			StateUpgrader: receiver.upgradeExampleResourceStateV0toV1,
		},
	}
}

func (receiver *PlanResource) upgradeExampleResourceStateV0toV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var old PlanModel0
	req.State.Get(ctx, &old)

	diags := resp.State.Set(ctx, FromPlanModel0(old))
	if util.TestDiagnostic(&resp.Diagnostics, diags) {
		return
	}
}

func (receiver *PlanResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *PlanResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics

		plan PlanModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	bambooPlan, err := receiver.client.PlanService().Create(bamboo.CreatePlan{
		PlanKey:    plan.PlanKey.ValueString(),
		Name:       plan.Name.ValueString(),
		ProjectKey: plan.Project.ValueString(),
	})
	if util.TestError(&response.Diagnostics, err, "Failed to create project") {
		return
	}

	//computation, diags := CreateProjectAssignments(ctx, receiver, plan)
	//if util.TestDiagnostic(&response.Diagnostics, diags) {
	//	return
	//}

	planModel := NewPlanModel(plan, bambooPlan)

	diags = response.State.Set(ctx, planModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *PlanResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics

		state PlanModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	bambooPlan, err := receiver.client.PlanService().Read(fmt.Sprintf("%s-%s", state.Project.ValueString(), state.PlanKey.ValueString()))
	if util.TestError(&response.Diagnostics, err, "Failed to create plan") {
		return
	}

	//computation, diags := ComputeProjectAssignments(ctx, receiver, state)
	//if util.TestDiagnostic(&response.Diagnostics, diags) {
	//	return
	//}

	planModel := NewPlanModel(state, bambooPlan)

	diags = response.State.Set(ctx, planModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *PlanResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags diag.Diagnostics

		plan, state PlanModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	bambooPlan, err := receiver.client.PlanService().Read(fmt.Sprintf("%s-%s", state.Project.ValueString(), state.PlanKey.ValueString()))
	if util.TestError(&response.Diagnostics, err, "Failed to read plan") {
		return
	}

	//forceUpdate := !plan.AssignmentVersion.Equal(state.AssignmentVersion)
	//computation, diags := UpdateProjectAssignments(ctx, receiver, plan, state, forceUpdate)
	//if util.TestDiagnostic(&response.Diagnostics, diags) {
	//	return
	//}

	planModel := NewPlanModel(plan, bambooPlan)

	diags = response.State.Set(ctx, planModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *PlanResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		state PlanModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	if !state.RetainOnDelete.ValueBool() {
		err := receiver.client.PlanService().Delete(fmt.Sprintf("%s-%s", state.Project.ValueString(), state.PlanKey.ValueString()))
		if util.TestError(&response.Diagnostics, err, "Failed to delete plan") {
			return
		}
	}

	response.State.RemoveResource(ctx)
}

func (receiver *PlanResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	slug := strings.Split(request.ID, "-")
	diags := response.State.Set(ctx, &PlanModel{
		Project: types.StringValue(slug[0]),
		PlanKey: types.StringValue(slug[1]),
		Name:    types.StringNull(),
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}
