package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/golang-quality-of-life-pack/collections"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
	"strconv"
)

type ProjectRepositoriesModel struct {
	Key          types.String `tfsdk:"key"`
	Repositories types.List   `tfsdk:"repositories"`
}

var (
	_ resource.Resource                = &ProjectRepositoriesResource{}
	_ resource.ResourceWithConfigure   = &ProjectRepositoriesResource{}
	_ resource.ResourceWithImportState = &ProjectRepositoriesResource{}
	_ ConfigurableReceiver             = &ProjectRepositoriesResource{}
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

func (receiver *ProjectRepositoriesResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
			},
			"repositories": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
		},
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

		_, err = receiver.client.ProjectService().AddSpecRepositories(plan.Key.ValueString(), repositoryId)
		if util.TestError(&response.Diagnostics, err, errorFailedToAddProjectRepositories) {
			return
		}
	}

	diags = response.State.Set(ctx, &ProjectRepositoriesModel{
		Key:          types.StringValue(plan.Key.ValueString()),
		Repositories: plan.Repositories,
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
		existingRepositoryIDs = make([]string, 0)
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositories, err := receiver.client.ProjectService().GetSpecRepositories(state.Key.ValueString())
	if err != nil {
		response.Diagnostics.AddError(errorFailedToReadRepository, err.Error())
		return
	}

	for _, repository := range repositories {
		projectRepositoryIDs = append(projectRepositoryIDs, strconv.Itoa(repository.ID))
	}

	diags = state.Repositories.ElementsAs(ctx, &projectRepositoryIDs, true)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	for _, project := range projectRepositoryIDs {
		if collections.Contains(projectRepositoryIDs, project) {
			existingRepositoryIDs = append(existingRepositoryIDs, project)
		}
	}

	listValue, diags := types.ListValueFrom(ctx, types.StringType, existingRepositoryIDs)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = response.State.Set(ctx, &ProjectRepositoriesModel{
		Key:          types.StringValue(state.Key.ValueString()),
		Repositories: listValue,
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

		_, err = receiver.client.ProjectService().AddSpecRepositories(plan.Key.ValueString(), repositoryId)
		if util.TestError(&response.Diagnostics, err, errorFailedToAddProjectRepositories) {
			return
		}
	}

	for _, repository := range removing {
		repositoryId, err := strconv.Atoi(repository)
		if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
			return
		}

		err = receiver.client.ProjectService().RemoveSpecRepositories(plan.Key.ValueString(), repositoryId)
		if util.TestError(&response.Diagnostics, err, errorFailedToRemoveProjectRepositories) {
			return
		}
	}

	diags = response.State.Set(ctx, &ProjectRepositoriesModel{
		Key:          types.StringValue(plan.Key.ValueString()),
		Repositories: plan.Repositories,
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

	diags = state.Repositories.ElementsAs(ctx, &existingRepositoryIDs, true)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	for _, repository := range existingRepositoryIDs {
		repositoryId, err := strconv.Atoi(repository)
		if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
			return
		}

		err = receiver.client.ProjectService().RemoveSpecRepositories(state.Key.ValueString(), repositoryId)
		if util.TestError(&response.Diagnostics, err, errorFailedToRemoveProjectRepositories) {
			return
		}
	}
	response.State.RemoveResource(ctx)
}

func (receiver *ProjectRepositoriesResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), request, response)
}
