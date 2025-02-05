package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/golang-quality-of-life-pack/collections"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
	"sort"
	"strconv"
)

type ProjectRepositoriesModel0 struct {
	RetainOnDelete types.Bool   `tfsdk:"retain_on_delete"`
	Key            types.String `tfsdk:"key"`
	Repositories   types.List   `tfsdk:"repositories"`
}

type ProjectRepositoriesModel struct {
	RetainOnDelete types.Bool   `tfsdk:"retain_on_delete"`
	Project        types.String `tfsdk:"project"`
	Repositories   types.List   `tfsdk:"repositories"`
}

var (
	_ resource.Resource                 = &ProjectRepositoriesResource{}
	_ resource.ResourceWithConfigure    = &ProjectRepositoriesResource{}
	_ resource.ResourceWithImportState  = &ProjectRepositoriesResource{}
	_ resource.ResourceWithUpgradeState = &ProjectRepositoriesResource{}
	_ ConfigurableReceiver              = &ProjectRepositoriesResource{}
)

func NewProjectRepositoriesResource() resource.Resource {
	return &ProjectRepositoriesResource{}
}

type ProjectRepositoriesResource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *ProjectRepositoriesResource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *ProjectRepositoriesResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_project_repositories"
}

func (receiver *ProjectRepositoriesResource) schemaV0() schema.Schema {
	return schema.Schema{
		MarkdownDescription: `This resource define project repository spec permissions.

In order for the execution to be successful, the user must have user access to all the specified repositories.`,
		Attributes: map[string]schema.Attribute{
			"retain_on_delete": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default value is `true`, and if the value set to `false` when the resource destroyed, the permission will be removed.",
			},
			"key": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
				MarkdownDescription: "Project key where the variable will be added",
			},
			"repositories": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "This project will add this list of linked repositories into its permission.",
			},
		},
	}
}

func (receiver *ProjectRepositoriesResource) schemaV1() schema.Schema {
	return schema.Schema{
		Version: 1,
		MarkdownDescription: `This resource define project repository spec permissions.

In order for the execution to be successful, the user must have user access to all the specified repositories.`,
		Attributes: map[string]schema.Attribute{
			"retain_on_delete": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default value is `true`, and if the value set to `false` when the resource destroyed, the permission will be removed.",
			},
			"project": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
				MarkdownDescription: "Project key where the variable will be added",
			},
			"repositories": schema.ListAttribute{
				Required:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "This project will add this list of linked repositories into its permission.",
			},
		},
	}
}

func (receiver *ProjectRepositoriesResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = receiver.schemaV1()
}

func (receiver *ProjectRepositoriesResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	v0 := receiver.schemaV0()
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &v0,
			StateUpgrader: receiver.upgradeExampleResourceStateV0toV1,
		},
	}
}

func (receiver *ProjectRepositoriesResource) upgradeExampleResourceStateV0toV1(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var old ProjectRepositoriesModel0
	req.State.Get(ctx, &old)

	diags := resp.State.Set(ctx, &ProjectRepositoriesModel{
		RetainOnDelete: types.BoolValue(old.RetainOnDelete.ValueBool()),
		Project:        old.Key,
		Repositories:   old.Repositories,
	})
	if util.TestDiagnostic(&resp.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectRepositoriesResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *ProjectRepositoriesResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		plan                 ProjectRepositoriesModel
		diags                diag.Diagnostics
		projectRepositoryIDs = make([]string, 0)
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = plan.Repositories.ElementsAs(ctx, &projectRepositoryIDs, true)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	for _, repository := range projectRepositoryIDs {
		repositoryId, err := strconv.Atoi(repository)
		if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
			return
		}

		_, err = receiver.client.ProjectService().AddSpecRepositories(plan.Project.ValueString(), repositoryId)
		if util.TestError(&response.Diagnostics, err, errorFailedToAddProjectRepositories) {
			return
		}
	}

	diags = response.State.Set(ctx, &ProjectRepositoriesModel{
		RetainOnDelete: types.BoolValue(plan.RetainOnDelete.ValueBool()),
		Project:        types.StringValue(plan.Project.ValueString()),
		Repositories:   plan.Repositories,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectRepositoriesResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		state                 ProjectRepositoriesModel
		diags                 diag.Diagnostics
		err                   error
		projectRepositoryIDs  = make([]string, 0)
		plannedRepositoryIDs  = make([]string, 0)
		existingRepositoryIDs = make([]string, 0)
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositories, err := receiver.client.ProjectService().GetSpecRepositories(state.Project.ValueString())
	if err != nil {
		response.Diagnostics.AddError(errorFailedToReadRepository, err.Error())
		return
	}

	for _, repository := range repositories {
		projectRepositoryIDs = append(projectRepositoryIDs, strconv.Itoa(repository.ID))
	}

	diags = state.Repositories.ElementsAs(ctx, &plannedRepositoryIDs, true)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	for _, project := range plannedRepositoryIDs {
		if collections.Contains(projectRepositoryIDs, project) {
			existingRepositoryIDs = append(existingRepositoryIDs, project)
		}
	}

	sort.Strings(existingRepositoryIDs)
	listValue, diags := types.ListValueFrom(ctx, types.StringType, existingRepositoryIDs)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = response.State.Set(ctx, &ProjectRepositoriesModel{
		RetainOnDelete: types.BoolValue(state.RetainOnDelete.ValueBool()),
		Project:        types.StringValue(state.Project.ValueString()),
		Repositories:   listValue,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectRepositoriesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		plan, state           ProjectRepositoriesModel
		incomingRepositoryIDs = make([]string, 0)
		existingRepositoryIDs = make([]string, 0)
		diags                 diag.Diagnostics
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
	diags = plan.Repositories.ElementsAs(ctx, &incomingRepositoryIDs, true)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = state.Repositories.ElementsAs(ctx, &existingRepositoryIDs, true)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	adding, removing := collections.Delta(existingRepositoryIDs, incomingRepositoryIDs)
	for _, repository := range adding {
		repositoryId, err := strconv.Atoi(repository)
		if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
			return
		}

		_, err = receiver.client.ProjectService().AddSpecRepositories(plan.Project.ValueString(), repositoryId)
		if util.TestError(&response.Diagnostics, err, errorFailedToAddProjectRepositories) {
			return
		}
	}

	for _, repository := range removing {
		repositoryId, err := strconv.Atoi(repository)
		if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
			return
		}

		err = receiver.client.ProjectService().RemoveSpecRepositories(plan.Project.ValueString(), repositoryId)
		if util.TestError(&response.Diagnostics, err, errorFailedToRemoveProjectRepositories) {
			return
		}
	}

	diags = response.State.Set(ctx, &ProjectRepositoriesModel{
		RetainOnDelete: types.BoolValue(plan.RetainOnDelete.ValueBool()),
		Project:        types.StringValue(plan.Project.ValueString()),
		Repositories:   plan.Repositories,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectRepositoriesResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		state                 ProjectRepositoriesModel
		existingRepositoryIDs = make([]string, 0)
		diags                 diag.Diagnostics
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	if !state.RetainOnDelete.ValueBool() {
		diags = state.Repositories.ElementsAs(ctx, &existingRepositoryIDs, true)
		if util.TestDiagnostic(&response.Diagnostics, diags) {
			return
		}

		for _, repository := range existingRepositoryIDs {
			repositoryId, err := strconv.Atoi(repository)
			if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
				return
			}

			err = receiver.client.ProjectService().RemoveSpecRepositories(state.Project.ValueString(), repositoryId)
			if util.TestError(&response.Diagnostics, err, errorFailedToRemoveProjectRepositories) {
				return
			}
		}
	}

	response.State.RemoveResource(ctx)
}

func (receiver *ProjectRepositoriesResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("project"), request, response)
}
