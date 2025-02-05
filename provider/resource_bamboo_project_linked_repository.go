package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
)

var (
	_ resource.Resource                   = &ProjectLinkedRepositoryResource{}
	_ resource.ResourceWithConfigure      = &ProjectLinkedRepositoryResource{}
	_ resource.ResourceWithImportState    = &ProjectLinkedRepositoryResource{}
	_ resource.ResourceWithUpgradeState   = &ProjectLinkedRepositoryResource{}
	_ LinkedRepositoryPermissionsReceiver = &ProjectLinkedRepositoryResource{}
	_ ConfigurableReceiver                = &ProjectLinkedRepositoryResource{}
)

func NewProjectLinkedRepositoryResource() resource.Resource {
	return &ProjectLinkedRepositoryResource{}
}

type ProjectLinkedRepositoryResource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *ProjectLinkedRepositoryResource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *ProjectLinkedRepositoryResource) getClient() *bamboo.Client {
	return receiver.client
}

func (receiver *ProjectLinkedRepositoryResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_project_linked_repository"
}

func (receiver *ProjectLinkedRepositoryResource) schemaV0() schema.Schema {
	return schema.Schema{
		MarkdownDescription: `This resource define project level linked repository.

One of the main focus of this resource is the permission management, which usually overlooked when creating linked repository through GUI.

The priority block has a priority that defines the final assigned permissions of the user or group.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Numeric id of the linked repository.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(projectLinkedRepositoryNameCheck, "", ""),
				},
				MarkdownDescription: "Name of the linked repository.",
			},
			"rss_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Flag to modify Bamboo Spec flag after creation.",
			},
			"key": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(projectLinkedRepositoryOwnerCheck, "", ""),
				},
				MarkdownDescription: "Bamboo project key that owns this linked repository.",
			},
			"project": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Bitbucket project key that owns the Git repository.",
			},
			"slug": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Bitbucket repository slug.",
			},
			"branch": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Bitbucket repository branch.",
				Default:             stringdefault.StaticString("master"),
			},
			"assignment_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Assignment version, used to force update the permission.",
			},
			"computed_users":  ComputedAssignmentSchema,
			"computed_groups": ComputedAssignmentSchema,
		},
		Blocks: map[string]schema.Block{
			"assignments": AssignmentSchema("READ", "ADMINISTRATION"),
		},
	}
}

func (receiver *ProjectLinkedRepositoryResource) schemaV1() schema.Schema {
	return schema.Schema{
		Version: 1,
		MarkdownDescription: `This resource define project level linked repository.

One of the main focus of this resource is the permission management, which usually overlooked when creating linked repository through GUI.

The priority block has a priority that defines the final assigned permissions of the user or group.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Numeric id of the linked repository.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(projectLinkedRepositoryNameCheck, "", ""),
				},
				MarkdownDescription: "Name of the linked repository.",
			},
			"rss_enabled": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Flag to modify Bamboo Spec flag after creation.",
			},
			"project": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(projectLinkedRepositoryOwnerCheck, "", ""),
				},
				MarkdownDescription: "Bamboo project key that owns this linked repository.",
			},
			"bitbucket_project": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Bitbucket project key that owns the Git repository.",
			},
			"slug": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Bitbucket repository slug.",
			},
			"branch": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Bitbucket repository branch.",
				Default:             stringdefault.StaticString("master"),
			},
			"assignment_version": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Assignment version, used to force update the permission.",
			},
			"computed_users":  ComputedAssignmentSchema,
			"computed_groups": ComputedAssignmentSchema,
		},
		Blocks: map[string]schema.Block{
			"assignments": AssignmentSchema("READ", "ADMINISTRATION"),
		},
	}
}

func (receiver *ProjectLinkedRepositoryResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = receiver.schemaV1()
}

func (receiver *ProjectLinkedRepositoryResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	v0 := receiver.schemaV0()
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &v0,
			StateUpgrader: receiver.upgradeExampleResourceStateV0toV1,
		},
	}
}

func (receiver *ProjectLinkedRepositoryResource) upgradeExampleResourceStateV0toV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var old ProjectLinkedRepositoryModel0
	req.State.Get(ctx, &old)

	diags := resp.State.Set(ctx, FromProjectLinkedRepositoryModel0(old))
	if util.TestDiagnostic(&resp.Diagnostics, diags) {
		return
	}
}

func projectLinkedRepositoryNameCheck(ctx context.Context, request planmodifier.StringRequest, response *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	var plan, state ProjectLinkedRepositoryModel

	diags := request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	response.RequiresReplace = !plan.Name.Equal(state.Name) && !state.Name.IsNull()
}

func projectLinkedRepositoryOwnerCheck(ctx context.Context, request planmodifier.StringRequest, response *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	var plan, state ProjectLinkedRepositoryModel

	diags := request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	response.RequiresReplace = !plan.Project.Equal(state.Project) && !state.Project.IsNull()
}

func (receiver *ProjectLinkedRepositoryResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *ProjectLinkedRepositoryResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics

		plan ProjectLinkedRepositoryModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repository, err := receiver.client.RepositoryService().ReadProject(plan.Project.ValueString(), plan.Name.ValueString())
	if err == nil && repository != nil {
		response.Diagnostics.AddError("linked repository already exists", "Unable to create as the requested repository already exists, manual deletion of project linked repository may be required")
		return
	}

	repositoryId, err := receiver.client.RepositoryService().CreateProject(bamboo.CreateProjectRepository{
		Project:        plan.Project.ValueString(),
		Name:           plan.Name.ValueString(),
		ProjectKey:     strings.ToLower(plan.BitbucketProject.ValueString()),
		RepositorySlug: strings.ToLower(plan.Slug.ValueString()),
		ServerId:       receiver.config.BambooRss.Server.ValueString(),
		ServerName:     receiver.config.BambooRss.Name.ValueString(),
		CloneUrl: strings.ToLower(fmt.Sprintf(
			receiver.config.BambooRss.CloneUrl.ValueString(),
			plan.BitbucketProject.ValueString(),
			plan.Slug.ValueString(),
		)),
	})
	if util.TestError(&response.Diagnostics, err, errorFailedToReadRepository) {
		return
	}

	err = receiver.client.RepositoryService().EnableCI(repositoryId, plan.RssEnabled.ValueBool())
	if util.TestError(&response.Diagnostics, err, errorFailedToUpdateRepository) {
		return
	}

	if plan.RssEnabled.ValueBool() {
		err = receiver.client.RepositoryService().ScanCI(repositoryId)
		if util.TestError(&response.Diagnostics, err, "Failed to execute scan repository") {
			return
		}
	}

	repository = &bamboo.Repository{
		ID:         repositoryId,
		Name:       plan.Name.ValueString(),
		RssEnabled: plan.RssEnabled.ValueBool(),
	}

	plan.ID = types.StringValue(fmt.Sprintf("%v", repository.ID))
	diags = response.State.SetAttribute(ctx, path.Root("id"), plan.ID)

	computation, diags := CreateProjectLinkedRepositoryAssignments(ctx, receiver, plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = response.State.Set(ctx, NewProjectLinkedRepositoryModel(plan, repository.ID, computation))
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectLinkedRepositoryResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics

		state ProjectLinkedRepositoryModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repository, err := receiver.client.RepositoryService().ReadProject(state.Project.ValueString(), state.Name.ValueString())
	if util.TestError(&response.Diagnostics, err, errorFailedToReadRepository) {
		return
	}

	if repository == nil {
		response.Diagnostics.AddError(errorFailedToReadRepository, fmt.Sprintf("No repository with name %s", state.Name.ValueString()))
		return
	}

	computation, diags := ComputeProjectLinkedRepositoryAssignments(ctx, receiver, state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewProjectLinkedRepositoryModel(state, repository.ID, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectLinkedRepositoryResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags       diag.Diagnostics
		plan, state ProjectLinkedRepositoryModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repository, err := receiver.client.RepositoryService().ReadProject(plan.Project.ValueString(), plan.Name.ValueString())
	if util.TestError(&response.Diagnostics, err, errorFailedToReadRepository) {
		return
	}

	if repository == nil {
		// the repository is no longer exists
		response.Diagnostics.AddError("Linked Repository no longer exists", plan.Name.ValueString())
		return
	}

	err = receiver.client.RepositoryService().EnableCI(repository.ID, plan.RssEnabled.ValueBool())
	if util.TestError(&response.Diagnostics, err, errorFailedToUpdateRepository) {
		return
	}

	if plan.RssEnabled.ValueBool() {
		err = receiver.client.RepositoryService().ScanCI(repository.ID)
		if util.TestError(&response.Diagnostics, err, "Failed to execute scan repository") {
			return
		}
	}

	if !plan.BitbucketProject.Equal(state.BitbucketProject) || !plan.Slug.Equal(state.Slug) {
		err = receiver.client.RepositoryService().UpdateProject(repository.ID, bamboo.CreateProjectRepository{
			Project:        plan.Project.ValueString(),
			Name:           plan.Name.ValueString(),
			ProjectKey:     strings.ToLower(plan.BitbucketProject.ValueString()),
			RepositorySlug: strings.ToLower(plan.Slug.ValueString()),
			ServerId:       receiver.config.BambooRss.Server.ValueString(),
			ServerName:     receiver.config.BambooRss.Name.ValueString(),
			CloneUrl: strings.ToLower(fmt.Sprintf(
				receiver.config.BambooRss.CloneUrl.ValueString(),
				plan.BitbucketProject.ValueString(),
				plan.Slug.ValueString(),
			)),
		})
		if util.TestError(&response.Diagnostics, err, errorFailedToUpdateRepository) {
			return
		}
	}

	forceUpdate := !plan.AssignmentVersion.Equal(state.AssignmentVersion)
	computation, diags := UpdateProjectLinkedRepositoryAssignments(ctx, receiver, plan, state, forceUpdate)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewProjectLinkedRepositoryModel(plan, repository.ID, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectLinkedRepositoryResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		state ProjectLinkedRepositoryModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryId, err := strconv.Atoi(state.ID.ValueString())
	if util.TestError(&response.Diagnostics, err, errorFailedToUpdateRepository) {
		return
	}

	err = receiver.client.RepositoryService().DeleteProject(state.BitbucketProject.ValueString(), repositoryId)
	if util.TestError(&response.Diagnostics, err, errorFailedToUpdateRepository) {
		return
	}

	response.State.RemoveResource(ctx)
}

func (receiver *ProjectLinkedRepositoryResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	slug := strings.Split(request.ID, "/")
	diags := response.State.Set(ctx, &ProjectLinkedRepositoryModel{
		Project:           types.StringValue(slug[0]),
		Name:              types.StringValue(slug[1]),
		RssEnabled:        types.BoolNull(),
		BitbucketProject:  types.StringNull(),
		Slug:              types.StringNull(),
		AssignmentVersion: types.StringNull(),
		Assignments:       types.ListNull(assignmentType),
		ComputedUsers:     types.ListNull(computedAssignmentType),
		ComputedGroups:    types.ListNull(computedAssignmentType),
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
	//resource.ImportStatePassthroughID(ctx, path.Root("name"), request, response)
}
