package provider

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/golang-quality-of-life-pack/collections"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
)

var (
	_ resource.Resource                = &DeploymentResource{}
	_ resource.ResourceWithConfigure   = &DeploymentResource{}
	_ resource.ResourceWithImportState = &DeploymentResource{}
	_ DeploymentPermissionsReceiver    = &DeploymentResource{}
	_ ConfigurableReceiver             = &DeploymentResource{}
)

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

func (receiver *DeploymentResource) getClient() *bamboo.Client {
	return receiver.client
}

type DeploymentResource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *DeploymentResource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *DeploymentResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_deployment"
}

func (receiver *DeploymentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: `This resource define deployment.

In order for the execution to be successful, the user must have user access to all the specified repositories.`,
		Attributes: map[string]schema.Attribute{
			"retain_on_delete": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Default value is `true`, and if the value set to `false` when the resource destroyed, the deployment will be removed.",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Numeric id of the deployment.",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIfConfigured(),
				},
				MarkdownDescription: "Name of the deployment.",
			},
			"plan_key": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Plan key that will be the source of the deployment.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description the deployment.",
			},
			"repository_specs_managed": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Computer value that defines the repository is managed by spec.",
			},
			"repositories": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.RegexMatches(
						regexp.MustCompile(`^\d+$`),
						"value must be a numeric",
					)),
				},
				MarkdownDescription: "This deployment will add this list of linked repositories into its permission.",
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
				"CREATE",
				"CREATEREPOSITORY",
				"ADMINISTRATION",
				"CLONE",
				"WRITE",
				"BUILD",
				"VIEWCONFIGURATION"),
		},
	}
}

func (receiver *DeploymentResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *DeploymentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics
		err   error

		plan                    DeploymentModel
		deploymentRepositoryIDs = make([]string, 0)
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = plan.Repositories.ElementsAs(ctx, &deploymentRepositoryIDs, true)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	deployment, err := receiver.client.DeploymentService().Create(bamboo.CreateDeployment{
		Name:        plan.Name.ValueString(),
		PlanKey:     bamboo.Key{Key: plan.PlanKey.ValueString()},
		Description: plan.Description.ValueString(),
	})
	if util.TestError(&response.Diagnostics, err, "Failed to create deployment") {
		return
	}

	deploymentID := types.StringValue(fmt.Sprintf("%v", deployment.ID))
	plan.ID = deploymentID

	diags = response.State.SetAttribute(ctx, path.Root("id"), deploymentID)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	for _, repository := range deploymentRepositoryIDs {
		repositoryId, err := strconv.Atoi(repository)
		if util.TestError(&response.Diagnostics, err, errorProvidedRepositoryMustBeNumber) {
			return
		}

		_, err = receiver.client.DeploymentService().AddSpecRepositories(deployment.ID, repositoryId)
		if util.TestError(&response.Diagnostics, err, "Failed to create add deployment repository") {
			return
		}
	}

	computation, diags := CreateDeploymentAssignments(ctx, receiver, plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	deploymentModel := NewDeploymentModel(plan, deployment, computation)

	diags = response.State.Set(ctx, deploymentModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *DeploymentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics
		err   error

		state        DeploymentModel
		deployment   *bamboo.Deployment
		deploymentId int
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	if state.ID.IsNull() {
		deployment, err = receiver.client.DeploymentService().Read(state.Name.ValueString())
		if util.TestError(&response.Diagnostics, err, errorProvidedDeploymentIdMustBeNumber) {
			return
		}

		deploymentId = deployment.ID
	} else {
		deploymentId, err = strconv.Atoi(state.ID.ValueString())
		if util.TestError(&response.Diagnostics, err, errorProvidedDeploymentIdMustBeNumber) {
			return
		}

		deployment, err = receiver.client.DeploymentService().ReadWithId(deploymentId)
		if util.TestError(&response.Diagnostics, err, errorFailedToReadDeployment) {
			return
		}
	}

	state.ID = types.StringValue(strconv.Itoa(deploymentId))

	repositories, err := receiver.client.DeploymentService().GetSpecRepositories(deploymentId)
	if util.TestError(&response.Diagnostics, err, "Failed to read deployment repositories") {
		return
	}

	computation, diags := ComputeDeploymentAssignments(ctx, receiver, state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	var deploymentRepositoryIDs = make([]string, 0)
	for _, repository := range repositories {
		deploymentRepositoryIDs = append(deploymentRepositoryIDs, strconv.Itoa(repository.ID))
	}

	sort.Strings(deploymentRepositoryIDs)
	repositoryList, diags := types.ListValueFrom(ctx, types.StringType, deploymentRepositoryIDs)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	deploymentModel := NewDeploymentModel(state, deployment, computation)

	if len(deploymentRepositoryIDs) > 0 {
		deploymentModel.Repositories = repositoryList
	}

	diags = response.State.Set(ctx, deploymentModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *DeploymentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags diag.Diagnostics
		err   error

		plan, state DeploymentModel
		computation *AssignmentResult
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	deploymentId, err := strconv.Atoi(state.ID.ValueString())
	if util.TestError(&response.Diagnostics, err, errorProvidedDeploymentIdMustBeNumber) {
		return
	}

	deployment, err := receiver.client.DeploymentService().ReadWithId(deploymentId)
	if util.TestError(&response.Diagnostics, err, errorFailedToReadDeployment) {
		return
	}

	// if the deployment is managed by repository spec, no more update can be made
	if !deployment.RepositorySpecsManaged {
		deployment, err = receiver.client.DeploymentService().UpdateWithId(deploymentId, bamboo.UpdateDeployment{
			Name:        plan.Name.ValueString(),
			PlanKey:     bamboo.Key{Key: plan.PlanKey.ValueString()},
			Description: plan.Description.ValueString(),
		})

		if util.TestError(&response.Diagnostics, err, "Failed to update deployment") {
			return
		}

		forceUpdate := !plan.AssignmentVersion.Equal(state.AssignmentVersion)
		computation, diags = UpdateDeploymentAssignments(ctx, receiver, plan, state, forceUpdate)
		if util.TestDiagnostic(&response.Diagnostics, diags) {
			return
		}
	} else {
		computation, diags = ComputeDeploymentAssignments(ctx, receiver, state)
		if util.TestDiagnostic(&response.Diagnostics, diags) {
			return
		}
	}

	diags = receiver.UpdateLinkedRepositories(ctx, deploymentId, plan, state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	deploymentModel := NewDeploymentModel(plan, deployment, computation)

	diags = response.State.Set(ctx, deploymentModel)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *DeploymentResource) UpdateLinkedRepositories(ctx context.Context, deploymentId int, plan DeploymentModel, state DeploymentModel) diag.Diagnostics {
	var (
		diags diag.Diagnostics

		incomingRepositoryIDs = make([]string, 0)
		existingRepositoryIDs = make([]string, 0)
	)

	diags = plan.Repositories.ElementsAs(ctx, &incomingRepositoryIDs, true)
	if diags != nil {
		return diags
	}

	diags = state.Repositories.ElementsAs(ctx, &existingRepositoryIDs, true)
	if diags != nil {
		return diags
	}

	adding, removing := collections.Delta(existingRepositoryIDs, incomingRepositoryIDs)
	for _, repository := range adding {
		repositoryId, err := strconv.Atoi(repository)
		if err != nil {
			return []diag.Diagnostic{diag.NewErrorDiagnostic(errorProvidedRepositoryMustBeNumber, err.Error())}
		}

		_, err = receiver.client.DeploymentService().AddSpecRepositories(deploymentId, repositoryId)
		if err != nil {
			return []diag.Diagnostic{diag.NewErrorDiagnostic("Failed to add deployment repositories", err.Error())}
		}
	}

	for _, repository := range removing {
		repositoryId, err := strconv.Atoi(repository)
		if err != nil {
			return []diag.Diagnostic{diag.NewErrorDiagnostic(errorProvidedRepositoryMustBeNumber, err.Error())}
		}

		err = receiver.client.DeploymentService().RemoveSpecRepositories(deploymentId, repositoryId)
		if err != nil {
			return []diag.Diagnostic{diag.NewErrorDiagnostic("Failed to remove deployment repositories", err.Error())}
		}
	}

	return nil
}

func (receiver *DeploymentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state DeploymentModel

	diags := request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	deploymentId, err := strconv.Atoi(state.ID.ValueString())
	if util.TestError(&response.Diagnostics, err, errorProvidedDeploymentIdMustBeNumber) {
		return
	}

	if !state.RetainOnDelete.ValueBool() {
		err = receiver.client.DeploymentService().Delete(deploymentId)
		if util.TestError(&response.Diagnostics, err, "Failed to delete deployment") {
			return
		}
	}

	response.State.RemoveResource(ctx)
}

func (receiver *DeploymentResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), request, response)
}
