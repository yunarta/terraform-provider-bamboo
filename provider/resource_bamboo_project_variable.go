package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
	"strings"
)

var (
	_ resource.Resource                = &ProjectVariableResource{}
	_ resource.ResourceWithConfigure   = &ProjectVariableResource{}
	_ resource.ResourceWithImportState = &ProjectVariableResource{}
	_ ProjectPermissionsReceiver       = &ProjectVariableResource{}
	_ ConfigurableReceiver             = &ProjectVariableResource{}
)

func NewProjectVariableResource() resource.Resource {
	return &ProjectVariableResource{}
}

type ProjectVariableResource struct {
	config BambooProviderConfig
	client *bamboo.Client
}

func (receiver *ProjectVariableResource) setConfig(config BambooProviderConfig, client *bamboo.Client) {
	receiver.config = config
	receiver.client = client
}

func (receiver *ProjectVariableResource) getClient() *bamboo.Client {
	return receiver.client
}

func (receiver *ProjectVariableResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_project_variable"
}

func (receiver *ProjectVariableResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: `This resource define project variables.
`,
		Attributes: map[string]schema.Attribute{
			"key": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
				MarkdownDescription: "Project key where the variable will be added",
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					util.ReplaceIfStringDiff(),
				},
				MarkdownDescription: "Name of the variable",
			},
			"value": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Value of the variable",
			},
			"secret": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Sensitive value of the variable. It will be masked during operation",
			},
		},
	}
}

func (receiver *ProjectVariableResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	ConfigureResource(receiver, ctx, request, response)
}

func (receiver *ProjectVariableResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var (
		diags diag.Diagnostics

		plan ProjectVariableModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	var value string
	if !plan.Secret.IsNull() {
		value = plan.Secret.ValueString()
	} else {
		value = plan.Value.ValueString()
	}

	err := receiver.client.ProjectService().PutVariables(
		plan.Key.ValueString(),
		plan.Name.ValueString(),
		value,
	)

	if util.TestError(&response.Diagnostics, err, "Failed to create project variable") {
		return
	}

	diags = response.State.Set(ctx, plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectVariableResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var (
		diags diag.Diagnostics

		state ProjectVariableModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	value, err := receiver.client.ProjectService().GetVariables(state.Key.ValueString(), state.Name.ValueString())
	if util.TestError(&response.Diagnostics, err, "Failed to create project") {
		return
	}

	if value == "********" && state.Secret.IsNull() {
		response.Diagnostics.AddError("Cannot import secret", fmt.Sprintf("%s is secret", state.Name.ValueString()))
		return
	}

	if !state.Secret.IsNull() {
		value = ""
	}

	diags = response.State.Set(ctx, ProjectVariableModel{
		Key:    state.Key,
		Name:   state.Name,
		Value:  util.NullString(value),
		Secret: state.Secret,
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectVariableResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		diags diag.Diagnostics

		plan, state ProjectVariableModel
	)

	diags = request.Plan.Get(ctx, &plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	var value string
	if !plan.Secret.IsNull() {
		value = plan.Secret.ValueString()
	} else {
		value = plan.Value.ValueString()
	}

	err := receiver.client.ProjectService().PutVariables(
		plan.Key.ValueString(),
		plan.Name.ValueString(),
		value,
	)
	if util.TestError(&response.Diagnostics, err, "Failed to update project variable") {
		return
	}

	diags = response.State.Set(ctx, plan)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}

func (receiver *ProjectVariableResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var (
		diags diag.Diagnostics
		state ProjectVariableModel
	)

	diags = request.State.Get(ctx, &state)
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}

	err := receiver.client.ProjectService().DeleteVariables(
		state.Key.ValueString(),
		state.Name.ValueString(),
	)
	if util.TestError(&response.Diagnostics, err, "Failed to delete project variable") {
		return
	}

	response.State.RemoveResource(ctx)
}

func (receiver *ProjectVariableResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	tokens := strings.Split(request.ID, "/")
	diags := response.State.Set(ctx, &ProjectVariableModel{
		Key:  types.StringValue(tokens[0]),
		Name: types.StringValue(tokens[1]),
	})
	if util.TestDiagnostic(&response.Diagnostics, diags) {
		return
	}
}
