package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
)

var (
	_ resource.Resource                = &ProjectPermissionsResource{}
	_ resource.ResourceWithConfigure   = &ProjectPermissionsResource{}
	_ resource.ResourceWithImportState = &ProjectPermissionsResource{}
	_ ProjectPermissionsReceiver       = &ProjectPermissionsResource{}
	_ ConfigurableReceiver             = &ProjectPermissionsResource{}
)

func NewProjectPermissionsResource() resource.Resource {
	return &ProjectPermissionsResource{}
}

type ProjectPermissionsResource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *ProjectPermissionsResource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *ProjectPermissionsResource) getClient() *bamboo.Client {
	return receiver.client
}

func (receiver *ProjectPermissionsResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_project_permissions"
}

func (receiver *ProjectPermissionsResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: `This resource define project user and groups project and plan permissions.

The priority block has a priority that defines the final assigned permissions of the user or group.`,
		Attributes: map[string]schema.Attribute{
			"retain_on_delete": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Default value is `true`, and if the value set to `false` when the resource destroyed, the permission will be removed.",
			},
			"key": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				MarkdownDescription: "Project key where the permissions will be added.",
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

func (receiver *ProjectPermissionsResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *ProjectPermissionsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics

		plan ProjectPermissionsModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	computation, diags := CreateProjectAssignments(ctx, receiver, plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewProjectPermissionsModel(plan, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectPermissionsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics

		state ProjectPermissionsModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	computation, diags := ComputeProjectAssignments(ctx, receiver, state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewProjectPermissionsModel(state, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectPermissionsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags diag.Diagnostics

		plan, state ProjectPermissionsModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	forceUpdate := !plan.AssignmentVersion.Equal(state.AssignmentVersion)
	computation, diags := UpdateProjectAssignments(ctx, receiver, plan, state, forceUpdate)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewProjectPermissionsModel(plan, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectPermissionsResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		state ProjectPermissionsModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	if !state.RetainOnDelete.ValueBool() {
		diags = DeleteProjectAssignments(ctx, receiver, state)
		if util.TestDiagnostic(&response.Diagnostics, diags) {
			return
		}
	}

	response.State.RemoveResource(ctx)
}

func (receiver *ProjectPermissionsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), request, response)
}
