package provider

import (
	"context"
	"fmt"
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
	_ resource.Resource                   = &LinkedRepositoryResource{}
	_ resource.ResourceWithConfigure      = &LinkedRepositoryResource{}
	_ resource.ResourceWithImportState    = &LinkedRepositoryResource{}
	_ LinkedRepositoryPermissionsReceiver = &LinkedRepositoryResource{}
	_ ConfigurableReceiver                = &LinkedRepositoryResource{}
)

func NewLinkedRepositoryResource() resource.Resource {
	return &LinkedRepositoryResource{}
}

type LinkedRepositoryResource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *LinkedRepositoryResource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *LinkedRepositoryResource) getClient() *bamboo.Client {
	return receiver.client
}

func (receiver *LinkedRepositoryResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_linked_repository"
}

func (receiver *LinkedRepositoryResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(linkedRepositoryNameCheck, "", ""),
				},
			},
			"rss_enabled": schema.BoolAttribute{
				Optional: true,
			},
			"project": schema.StringAttribute{
				Required: true,
			},
			"slug": schema.StringAttribute{
				Required: true,
			},
			"assignment_version": schema.StringAttribute{
				Optional: true,
			},
			"computed_users":  ComputedAssignmentSchema,
			"computed_groups": ComputedAssignmentSchema,
		},
		Blocks: map[string]schema.Block{
			"assignments": AssignmentSchema("READ", "ADMINISTRATION"),
		},
	}
}

func linkedRepositoryNameCheck(ctx context.Context, request planmodifier.StringRequest, response *stringplanmodifier.RequiresReplaceIfFuncResponse) {
	var plan, state LinkedRepositoryModel

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

func (receiver *LinkedRepositoryResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *LinkedRepositoryResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics

		plan LinkedRepositoryModel
	)

	repository, err := receiver.client.RepositoryService().Read(plan.Name.ValueString())
	if err == nil && repository != nil {
		response.Diagnostics.AddError("linked repository already exists", "Unable to create as the requested repository already exists")
	}

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryId, err := receiver.client.RepositoryService().Create(bamboo.CreateRepository{
		Name:           plan.Name.ValueString(),
		ProjectKey:     strings.ToLower(plan.Project.ValueString()),
		RepositorySlug: strings.ToLower(plan.Slug.ValueString()),
		ServerId:       receiver.config.BambooRss.Server.ValueString(),
		ServerName:     receiver.config.BambooRss.Name.ValueString(),
		CloneUrl: strings.ToLower(fmt.Sprintf(
			receiver.config.BambooRss.CloneUrl.ValueString(),
			plan.Project.ValueString(),
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

	repository = &bamboo.Repository{
		ID:         repositoryId,
		Name:       plan.Name.ValueString(),
		RssEnabled: plan.RssEnabled.ValueBool(),
	}

	plan.ID = types.StringValue(fmt.Sprintf("%v", repository.ID))
	diags = response.State.SetAttribute(ctx, path.Root("id"), plan.ID)

	computation, diags := CreateLinkedRepositoryAssignments(ctx, receiver, plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = response.State.Set(ctx, NewLinkedRepositoryModel(plan, repository.ID, computation))
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *LinkedRepositoryResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics

		state LinkedRepositoryModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repository, err := receiver.client.RepositoryService().Read(state.Name.ValueString())
	if util.TestError(&response.Diagnostics, err, errorFailedToReadRepository) {
		return
	}

	if repository == nil {
		response.Diagnostics.AddError(errorFailedToReadRepository, fmt.Sprintf("No repository with name %s", state.Name.ValueString()))
		return
	}

	computation, diags := CreateLinkedRepositoryAssignments(ctx, receiver, state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewLinkedRepositoryModel(state, repository.ID, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *LinkedRepositoryResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags       diag.Diagnostics
		plan, state LinkedRepositoryModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repository, err := receiver.client.RepositoryService().Read(plan.Name.ValueString())
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

	if !plan.Project.Equal(state.Project) || !plan.Slug.Equal(state.Slug) {
		err = receiver.client.RepositoryService().Update(repository.ID, bamboo.CreateRepository{
			Name:           plan.Name.ValueString(),
			ProjectKey:     strings.ToLower(plan.Project.ValueString()),
			RepositorySlug: strings.ToLower(plan.Slug.ValueString()),
			ServerId:       receiver.config.BambooRss.Server.ValueString(),
			ServerName:     receiver.config.BambooRss.Name.ValueString(),
			CloneUrl: strings.ToLower(fmt.Sprintf(
				receiver.config.BambooRss.CloneUrl.ValueString(),
				plan.Project.ValueString(),
				plan.Slug.ValueString(),
			)),
		})
		if util.TestError(&response.Diagnostics, err, errorFailedToUpdateRepository) {
			return
		}
	}

	forceUpdate := !plan.AssignmentVersion.Equal(state.AssignmentVersion)
	computation, diags := UpdateLinkedRepositoryAssignments(ctx, receiver, plan, state, forceUpdate)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	repositoryModel := NewLinkedRepositoryModel(plan, repository.ID, computation)

	diags = response.State.Set(ctx, repositoryModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *LinkedRepositoryResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	response.State.RemoveResource(ctx)
}

func (receiver *LinkedRepositoryResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), request, response)
}
