package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/golang-quality-of-life-pack/collections"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
	"regexp"
	"sort"
	"strconv"
)

type LinkedRepositoryDependencyModel struct {
	RetainOnDelete types.Bool   `tfsdk:"retain_on_delete"`
	ID             types.String `tfsdk:"id"`
	Repositories   types.List   `tfsdk:"requires"`
}

var (
	_ resource.Resource              = &LinkedRepositoryDependencyResource{}
	_ resource.ResourceWithConfigure = &LinkedRepositoryDependencyResource{}
	_ ConfigurableReceiver           = &LinkedRepositoryDependencyResource{}
)

func NewLinkedRepositoryDependencyResource() resource.Resource {
	return &LinkedRepositoryDependencyResource{}
}

type LinkedRepositoryDependencyResource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *LinkedRepositoryDependencyResource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *LinkedRepositoryDependencyResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_linked_repository_dependency"
}

func (receiver *LinkedRepositoryDependencyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: `This resource define relationship where repository specified by id will requires access to list of specified required repositories.

In order for the execution to be successful, the user must have admin access to all the required repositories. 
`,
		Attributes: map[string]schema.Attribute{
			"retain_on_delete": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Default value is `true`, and if the value set to `false` when the resource destroyed, the permission will be removed.",
			},
			"id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Numeric id of the linked repository.",
			},
			"requires": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.RegexMatches(
						regexp.MustCompile(`^\d+$`),
						"value must be a numeric",
					)),
				},
				MarkdownDescription: "This repository will be added into to this list of linked repositories permissions.",
			},
		},
	}
}

func (receiver *LinkedRepositoryDependencyResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *LinkedRepositoryDependencyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics
		err   error

		plan         LinkedRepositoryDependencyModel
		dependencies = make([]string, 0)

		repositoryId int
	)

	if util.TestDiagnostics(&response.Diagnostics,
		request.Plan.Get(ctx, &plan),
		plan.Repositories.ElementsAs(ctx, &dependencies, true),
	) {
		return
	}

	repositoryId, err = strconv.Atoi(plan.ID.ValueString())
	if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
		return
	}

	if collections.Contains(dependencies, plan.ID.ValueString()) {
		response.Diagnostics.AddError("Cannot add self as accessor", fmt.Sprintf("Repository %s", plan.ID.ValueString()))
		return
	}

	for _, repository := range dependencies {
		var (
			dependency   int
			repositories []bamboo.Repository
		)

		dependency, err = strconv.Atoi(repository)
		if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
			return
		}

		repositories, err = receiver.client.RepositoryService().ReadAccessor(dependency)
		if util.TestError(&response.Diagnostics, err, errorFailedToReadDeployment) {
			return
		}

		if !collections.ContainsFunc(repositories, func(e bamboo.Repository) bool { return e.ID == repositoryId }) {
			_, err = receiver.client.RepositoryService().AddAccessor(dependency, repositoryId)
			if util.TestError(&response.Diagnostics, err, errorFailedToAddRepositoryAccessor) {
				return
			}
		}
	}

	diags = response.State.Set(ctx, &LinkedRepositoryDependencyModel{
		RetainOnDelete: plan.RetainOnDelete,
		ID:             types.StringValue(strconv.Itoa(repositoryId)),
		Repositories:   plan.Repositories,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *LinkedRepositoryDependencyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state LinkedRepositoryDependencyModel
	var err error
	var dependencies = make([]string, 0)
	var newDependencies []string

	if util.TestDiagnostics(&response.Diagnostics,
		request.State.Get(ctx, &state),
		state.Repositories.ElementsAs(ctx, &dependencies, true),
	) {
		return
	}

	repositoryId, err := strconv.Atoi(state.ID.ValueString())
	if util.TestError(&response.Diagnostics, err, errorProvidedDeploymentIdMustBeNumber) {
		return
	}

	for _, repository := range dependencies {
		var (
			dependency   int
			repositories []bamboo.Repository
		)

		dependency, err = strconv.Atoi(repository)
		if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
			return
		}

		repositories, err = receiver.client.RepositoryService().ReadAccessor(dependency)
		if err != nil {
			response.Diagnostics.AddError(errorFailedToReadDeployment, err.Error())
			return
		}

		for _, accessor := range repositories {
			if repositoryId == accessor.ID {
				newDependencies = append(newDependencies, strconv.Itoa(dependency))
				break
			}
		}
	}

	sort.Strings(newDependencies)
	from, diags := types.ListValueFrom(ctx, types.StringType, newDependencies)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = response.State.Set(ctx, &LinkedRepositoryDependencyModel{
		RetainOnDelete: state.RetainOnDelete,
		ID:             types.StringValue(fmt.Sprintf("%v", repositoryId)),
		Repositories:   from,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *LinkedRepositoryDependencyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state LinkedRepositoryDependencyModel

	var diags diag.Diagnostics
	var err error
	var repositoryId int

	var plannedDependencies = make([]string, 0)
	var inStateDependencies = make([]string, 0)

	if util.TestDiagnostics(&response.Diagnostics,
		request.Plan.Get(ctx, &plan),
		request.State.Get(ctx, &state),
		plan.Repositories.ElementsAs(ctx, &plannedDependencies, true),
		state.Repositories.ElementsAs(ctx, &inStateDependencies, true),
	) {
		return
	}

	repositoryId, err = strconv.Atoi(plan.ID.ValueString())
	if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
		return
	}

	adding, removing := collections.Delta(inStateDependencies, plannedDependencies)

	if collections.Contains(adding, plan.ID.ValueString()) {
		response.Diagnostics.AddError("Cannot add self as accessor", fmt.Sprintf("Repository %s", plan.ID.ValueString()))
		return
	}

	for _, repository := range adding {
		dependency, _ := strconv.Atoi(repository)
		_, err = receiver.client.RepositoryService().AddAccessor(dependency, repositoryId)
		if util.TestError(&response.Diagnostics, err, errorFailedToAddRepository) {
			return
		}
	}

	for _, repository := range removing {
		dependency, _ := strconv.Atoi(repository)
		err = receiver.client.RepositoryService().RemoveAccessor(dependency, repositoryId)
		if util.TestError(&response.Diagnostics, err, errorFailedToRemoveRepository) {
			return
		}
	}

	diags = response.State.Set(ctx, &LinkedRepositoryDependencyModel{
		RetainOnDelete: plan.RetainOnDelete,
		ID:             types.StringValue(strconv.Itoa(repositoryId)),
		Repositories:   plan.Repositories,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *LinkedRepositoryDependencyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state LinkedRepositoryDependencyModel
	var existingDependencies = make([]string, 0)

	var err error

	if util.TestDiagnostics(&response.Diagnostics,
		request.State.Get(ctx, &state),
		state.Repositories.ElementsAs(ctx, &existingDependencies, true),
	) {
		return
	}

	if !state.RetainOnDelete.ValueBool() {
		var repositoryId int
		repositoryId, err = strconv.Atoi(state.ID.ValueString())
		if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
			return
		}

		for _, repository := range existingDependencies {
			dependency, _ := strconv.Atoi(repository)
			err = receiver.client.RepositoryService().RemoveAccessor(dependency, repositoryId)
			if util.TestError(&response.Diagnostics, err, errorFailedToRemoveRepository) {
				return
			}
		}
	}

	response.State.RemoveResource(ctx)
}
