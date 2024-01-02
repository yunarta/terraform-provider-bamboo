package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

type LinkedRepositoryAccessorModel struct {
	ID           types.String `tfsdk:"id"`
	Repositories types.List   `tfsdk:"repositories"`
}

var (
	_ resource.Resource                = &LinkedRepositoryAccessorResource{}
	_ resource.ResourceWithConfigure   = &LinkedRepositoryAccessorResource{}
	_ resource.ResourceWithImportState = &LinkedRepositoryAccessorResource{}
	_ ConfigurableReceiver             = &LinkedRepositoryAccessorResource{}
)

func NewLinkedRepositoryAccessorResource() resource.Resource {
	return &LinkedRepositoryAccessorResource{}
}

type LinkedRepositoryAccessorResource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *LinkedRepositoryAccessorResource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *LinkedRepositoryAccessorResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_linked_repository_accessor"
}

func (receiver *LinkedRepositoryAccessorResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: `Repository accessor define relationship that allow other repositories to access this repository.
					  It only add and remove repositories listed in its properties`,
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

func (receiver *LinkedRepositoryAccessorResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *LinkedRepositoryAccessorResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan LinkedRepositoryAccessorModel

	var diags diag.Diagnostics
	var err error

	diags = request.Plan.Get(ctx, &plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	var repositoryId int

	repositoryId, err = strconv.Atoi(plan.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(errorProvidedRepositoryMustBeNumber, err.Error())
		return
	}

	var repositories []bamboo.Repository
	var existingRepositories = make([]string, 0)
	var incomingRepositories = make([]string, 0)

	repositories, err = receiver.client.RepositoryService().ReadAccessor(repositoryId)
	if err != nil {
		response.Diagnostics.AddError(errorFailedToReadDeployment, err.Error())
		return
	}

	for _, repository := range repositories {
		existingRepositories = append(existingRepositories, strconv.Itoa(repository.ID))
	}

	diags = plan.Repositories.ElementsAs(ctx, &incomingRepositories, true)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	adding, _ := collections.Delta(existingRepositories, incomingRepositories)

	if collections.Contains(adding, plan.ID.ValueString()) {
		response.Diagnostics.AddError("Cannot add self as accessor", fmt.Sprintf("Repository %s", plan.ID.ValueString()))
		return
	}

	for _, repository := range adding {
		accessorId, _ := strconv.Atoi(repository)
		_, err = receiver.client.RepositoryService().AddAccessor(repositoryId, accessorId)
		if err != nil {
			response.Diagnostics.AddError(errorFailedToAddRepositoryAccessor, err.Error())
			return
		}
	}

	diags = response.State.Set(ctx, &LinkedRepositoryAccessorModel{
		ID:           types.StringValue(strconv.Itoa(repositoryId)),
		Repositories: plan.Repositories,
	})

	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
}

func (receiver *LinkedRepositoryAccessorResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state LinkedRepositoryAccessorModel
	var existingRepositories = make([]string, 0)

	diags := request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = state.Repositories.ElementsAs(ctx, &existingRepositories, true)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryId, err := strconv.Atoi(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(errorProvidedRepositoryMustBeNumber, err.Error())
		return
	}

	repositories, err := receiver.client.RepositoryService().ReadAccessor(repositoryId)
	if err != nil {
		response.Diagnostics.AddError(errorFailedToReadRepositoryAccessor, err.Error())
		return
	}

	var repositoryIds []attr.Value
	for _, repository := range repositories {
		accessorId := fmt.Sprintf("%v", repository.ID)
		if collections.Contains(existingRepositories, accessorId) {
			repositoryIds = append(repositoryIds, types.StringValue(accessorId))
		}
	}

	diags = response.State.Set(ctx, &LinkedRepositoryAccessorModel{
		ID:           types.StringValue(fmt.Sprintf("%v", repositoryId)),
		Repositories: types.ListValueMust(types.StringType, repositoryIds),
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *LinkedRepositoryAccessorResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state LinkedRepositoryAccessorModel

	var diags diag.Diagnostics
	var err error

	var inStateRepositories = make([]string, 0)
	var plannedRepository = make([]string, 0)

	if util.TestDiagnostics(&response.Diagnostics,
		request.Plan.Get(ctx, &plan),
		request.State.Get(ctx, &state),
		plan.Repositories.ElementsAs(ctx, &plannedRepository, true),
		state.Repositories.ElementsAs(ctx, &inStateRepositories, true),
	) {
		return
	}

	var repositoryId int

	repositoryId, err = strconv.Atoi(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(errorProvidedRepositoryMustBeNumber, err.Error())
		return
	}

	var repositories []bamboo.Repository
	var existingRepositories = make([]string, 0)

	repositories, err = receiver.client.RepositoryService().ReadAccessor(repositoryId)
	if err != nil {
		response.Diagnostics.AddError(errorFailedToReadRepositoryAccessor, err.Error())
		return
	}

	for _, repository := range repositories {
		existingRepositories = append(existingRepositories, strconv.Itoa(repository.ID))
	}

	adding, removing := collections.Delta(inStateRepositories, plannedRepository)
	if collections.Contains(adding, plan.ID.ValueString()) {
		response.Diagnostics.AddError("Cannot add self as accessor", fmt.Sprintf("Repository %s", plan.ID.ValueString()))
		return
	}

	for _, repository := range adding {
		if !collections.Contains(existingRepositories, repository) {
			accessorId, _ := strconv.Atoi(repository)
			_, err = receiver.client.RepositoryService().AddAccessor(repositoryId, accessorId)
			if err != nil {
				response.Diagnostics.AddError(errorFailedToAddRepositoryAccessor, err.Error())
				return
			}
		}
	}

	for _, repository := range removing {
		accessorId, _ := strconv.Atoi(repository)
		err = receiver.client.RepositoryService().RemoveAccessor(repositoryId, accessorId)
		if err != nil {
			response.Diagnostics.AddError(errorFailedToRemoveRepositoryAccessor, err.Error())
			return
		}
	}

	diags = response.State.Set(ctx, &LinkedRepositoryAccessorModel{
		ID:           types.StringValue(strconv.Itoa(repositoryId)),
		Repositories: plan.Repositories,
	})

	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *LinkedRepositoryAccessorResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state LinkedRepositoryAccessorModel

	var err error
	var repositoryId int
	var inStateRepositories = make([]string, 0)

	if util.TestDiagnostics(&response.Diagnostics,
		request.State.Get(ctx, &state),
		state.Repositories.ElementsAs(ctx, &inStateRepositories, true)) {
		return
	}

	repositoryId, err = strconv.Atoi(state.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError(errorProvidedRepositoryMustBeNumber, err.Error())
		return
	}

	var repositories []bamboo.Repository
	var existingRepositories = make([]string, 0)

	repositories, err = receiver.client.RepositoryService().ReadAccessor(repositoryId)
	if err != nil {
		response.Diagnostics.AddError(errorFailedToReadRepositoryAccessor, err.Error())
		return
	}

	for _, repository := range repositories {
		existingRepositories = append(existingRepositories, strconv.Itoa(repository.ID))
	}

	for _, repository := range inStateRepositories {

		if collections.Contains(existingRepositories, repository) {
			accessorId, _ := strconv.Atoi(repository)
			err = receiver.client.RepositoryService().RemoveAccessor(repositoryId, accessorId)
			if err != nil {
				response.Diagnostics.AddError(errorFailedToRemoveRepositoryAccessor, err.Error())
				return
			}
		}
	}

	response.State.RemoveResource(ctx)
}

func (receiver *LinkedRepositoryAccessorResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
