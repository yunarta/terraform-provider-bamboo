package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
)

var (
	_ resource.Resource                 = &ProjectResource{}
	_ resource.ResourceWithConfigure    = &ProjectResource{}
	_ resource.ResourceWithImportState  = &ProjectResource{}
	_ resource.ResourceWithUpgradeState = &ProjectResource{}
	_ ProjectPermissionsReceiver        = &ProjectResource{}
	_ ConfigurableReceiver              = &ProjectResource{}
)

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

type ProjectResource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *ProjectResource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *ProjectResource) getClient() *bamboo.Client {
	return receiver.client
}

func (receiver *ProjectResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_project"
}

func (receiver *ProjectResource) schemaV0() schema.Schema {
	return schema.Schema{
		MarkdownDescription: `This resource define project.

The priority block has a priority that defines the final assigned permissions of the user or group.`,
		Attributes: map[string]schema.Attribute{
			"retain_on_delete": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Default value is `true`, and if the value set to `false` when the resource destroyed, the project will be removed.",
			},
			"key": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
				MarkdownDescription: "Project key.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project name.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Project description.",
			},
			"assignment_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Assignment version, used to force update the permission.",
			},
			"computed_users":  ComputedAssignmentSchema,
			"computed_groups": ComputedAssignmentSchema,
		},
		Blocks: map[string]schema.Block{
			"assignments": AssignmentSchema(
				"READ",
				"VIEWCONFIGURATION",
				"WRITE",
				"BUILD",
				"CLONE",
				"CREATE",
				"CREATEREPOSITORY",
				"ADMINISTRATION",
			),
		},
	}
}

func (receiver *ProjectResource) schemaV1() schema.Schema {
	return schema.Schema{
		Version: 1,
		MarkdownDescription: `This resource define project.

The priority block has a priority that defines the final assigned permissions of the user or group.`,
		Attributes: map[string]schema.Attribute{
			"retain_on_delete": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Default value is `true`, and if the value set to `false` when the resource destroyed, the project will be removed.",
			},
			"project": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
				MarkdownDescription: "Project key.",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Project name.",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				Optional:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Project description.",
			},
			"assignment_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Assignment version, used to force update the permission.",
			},
			"computed_users":  ComputedAssignmentSchema,
			"computed_groups": ComputedAssignmentSchema,
		},
		Blocks: map[string]schema.Block{
			"assignments": AssignmentSchema(
				"READ",
				"VIEWCONFIGURATION",
				"WRITE",
				"BUILD",
				"CLONE",
				"CREATE",
				"CREATEREPOSITORY",
				"ADMINISTRATION",
			),
		},
	}
}

func (receiver *ProjectResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = receiver.schemaV1()
}

func (receiver *ProjectResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	v0 := receiver.schemaV0()
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &v0,
			StateUpgrader: receiver.upgradeExampleResourceStateV0toV1,
		},
	}
}

func (receiver *ProjectResource) upgradeExampleResourceStateV0toV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var old ProjectModel0
	req.State.Get(ctx, &old)

	diags := resp.State.Set(ctx, FromProjectModel0(old))
	if util.TestDiagnostic(&resp.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *ProjectResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics

		plan ProjectModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	project, err := receiver.client.ProjectService().Create(bamboo.CreateProject{
		Key:         plan.Project.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	})
	if util.TestError(&response.Diagnostics, err, "Failed to create project") {
		return
	}

	computation, diags := CreateProjectAssignments(ctx, receiver, plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewProjectModel(plan, project, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics

		state ProjectModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	project, err := receiver.client.ProjectService().Read(state.Project.ValueString())
	if util.TestError(&response.Diagnostics, err, "Failed to create project") {
		return
	}

	computation, diags := ComputeProjectAssignments(ctx, receiver, state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewProjectModel(state, project, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags diag.Diagnostics

		plan, state ProjectModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	project, err := receiver.client.ProjectService().Read(plan.Project.ValueString())
	if util.TestError(&response.Diagnostics, err, "Failed to read project") {
		return
	}

	forceUpdate := !plan.AssignmentVersion.Equal(state.AssignmentVersion)
	computation, diags := UpdateProjectAssignments(ctx, receiver, plan, state, forceUpdate)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewProjectModel(plan, project, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		state ProjectModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	if !state.RetainOnDelete.ValueBool() {
		err := receiver.client.ProjectService().Delete(state.Project.ValueString())
		if util.TestError(&response.Diagnostics, err, "Failed to delete project") {
			return
		}
	}

	response.State.RemoveResource(ctx)
}

func (receiver *ProjectResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("project"), request, response)
}
