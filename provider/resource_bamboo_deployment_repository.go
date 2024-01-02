package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/golang-quality-of-life-pack/collections"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
	"strconv"
)

type DeploymentRepositoriesModel struct {
	ID           types.String `tfsdk:"id"`
	Repositories types.List   `tfsdk:"repositories"`
}

var (
	_ resource.Resource                = &DeploymentRepositoriesResource{}
	_ resource.ResourceWithConfigure   = &DeploymentRepositoriesResource{}
	_ resource.ResourceWithImportState = &DeploymentRepositoriesResource{}
	_ ConfigurableReceiver             = &DeploymentRepositoriesResource{}
)

func NewDeploymentRepositoryResource() resource.Resource {
	return &DeploymentRepositoriesResource{}
}

type DeploymentRepositoriesResource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *DeploymentRepositoriesResource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *DeploymentRepositoriesResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_deployment_repositories"
}

func (receiver *DeploymentRepositoriesResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"repositories": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (receiver *DeploymentRepositoriesResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *DeploymentRepositoriesResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		plan                    DeploymentRepositoriesModel
		diags                   diag.Diagnostics
		err                     error
		deploymentRepositoryIDs = make([]string, 0)
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	deploymentId, err := strconv.Atoi(plan.ID.ValueString())
	if util.TestError(&response.Diagnostics, err, errorProvidedDeploymentIdMustBeNumber) {
		return
	}

	diags = plan.Repositories.ElementsAs(ctx, &deploymentRepositoryIDs, true)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	for _, repository := range deploymentRepositoryIDs {
		repositoryId, err := strconv.Atoi(repository)
		if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
			return
		}

		_, err = receiver.client.DeploymentService().AddSpecRepositories(deploymentId, repositoryId)
		if util.TestError(&response.Diagnostics, err, "Failed to add deployment repository") {
			return
		}
	}

	diags = response.State.Set(ctx, &DeploymentRepositoriesModel{
		ID:           types.StringValue(strconv.Itoa(deploymentId)),
		Repositories: plan.Repositories,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *DeploymentRepositoriesResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		state                   DeploymentRepositoriesModel
		diags                   diag.Diagnostics
		err                     error
		deploymentRepositoryIDs = make([]string, 0)
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	deploymentId, err := strconv.Atoi(state.ID.ValueString())
	if util.TestError(&response.Diagnostics, err, errorProvidedDeploymentIdMustBeNumber) {
		return
	}

	repositories, err := receiver.client.DeploymentService().GetSpecRepositories(deploymentId)
	if err != nil {
		response.Diagnostics.AddError("Failed to read repositories", err.Error())
		return
	}

	for _, repository := range repositories {
		deploymentRepositoryIDs = append(deploymentRepositoryIDs, strconv.Itoa(repository.ID))
	}

	listValue, diags := types.ListValueFrom(ctx, types.StringType, deploymentRepositoryIDs)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = response.State.Set(ctx, &DeploymentRepositoriesModel{
		ID:           types.StringValue(fmt.Sprintf("%v", deploymentId)),
		Repositories: listValue,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *DeploymentRepositoriesResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		plan, state           DeploymentRepositoriesModel
		incomingRepositoryIDs = make([]string, 0)
		existingRepositoryIDs = make([]string, 0)
		diags                 diag.Diagnostics
		err                   error
	)

	if util.TestDiagnostics(&response.Diagnostics,
		request.Plan.Get(ctx, &plan),
		request.State.Get(ctx, &state),
		plan.Repositories.ElementsAs(ctx, &incomingRepositoryIDs, true),
		state.Repositories.ElementsAs(ctx, &existingRepositoryIDs, true),
	) {
		return
	}

	deploymentId, err := strconv.Atoi(plan.ID.ValueString())
	if util.TestError(&response.Diagnostics, err, errorProvidedDeploymentIdMustBeNumber) {
		return
	}

	adding, removing := collections.Delta(existingRepositoryIDs, incomingRepositoryIDs)
	for _, repository := range adding {
		repositoryId, err := strconv.Atoi(repository)
		if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
			return
		}

		_, err = receiver.client.DeploymentService().AddSpecRepositories(deploymentId, repositoryId)
		if util.TestError(&response.Diagnostics, err, "Failed to add deployment repositories") {
			return
		}
	}

	for _, repository := range removing {
		repositoryId, err := strconv.Atoi(repository)
		if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
			return
		}

		err = receiver.client.DeploymentService().RemoveSpecRepositories(deploymentId, repositoryId)
		if util.TestError(&response.Diagnostics, err, "Failed to remove deployment repositories") {
			return
		}
	}

	diags = response.State.Set(ctx, &DeploymentRepositoriesModel{
		ID:           types.StringValue(strconv.Itoa(deploymentId)),
		Repositories: plan.Repositories,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *DeploymentRepositoriesResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		state                 DeploymentRepositoriesModel
		existingRepositoryIDs []string
		diags                 diag.Diagnostics
		err                   error
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	deploymentId, err := strconv.Atoi(state.ID.ValueString())
	if util.TestError(&response.Diagnostics, err, errorProvidedDeploymentIdMustBeNumber) {
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

		err = receiver.client.DeploymentService().RemoveSpecRepositories(deploymentId, repositoryId)
		if util.TestError(&response.Diagnostics, err, "Failed to remove deployment repositories") {
			return
		}
	}
	response.State.RemoveResource(ctx)
}

func (receiver *DeploymentRepositoriesResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
