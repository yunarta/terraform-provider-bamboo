package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
	"strconv"
)

type DeploymentModel struct {
	RetainOnDelete         types.Bool   `tfsdk:"retain_on_delete"`
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	PlanKey                types.String `tfsdk:"plan_key"`
	Description            types.String `tfsdk:"description"`
	RepositorySpecsManaged types.Bool   `tfsdk:"repository_specs_managed"`
	Repositories           types.List   `tfsdk:"repositories"`

	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       types.List   `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ DeploymentPermissionInterface = &DeploymentModel{}

func (d DeploymentModel) getAssignment(ctx context.Context) (Assignments, diag.Diagnostics) {
	var assignments Assignments = make([]Assignment, 0)

	diags := d.Assignments.ElementsAs(ctx, &assignments, true)
	return assignments, diags
}

func (d DeploymentModel) getDeploymentId(ctx context.Context) int {
	deploymentId, _ := strconv.Atoi(d.ID.ValueString())
	return deploymentId
}

func NewDeploymentModel(plan DeploymentModel, deployment *bamboo.Deployment, assignmentResult *AssignmentResult) *DeploymentModel {
	return &DeploymentModel{
		RetainOnDelete:         plan.RetainOnDelete,
		ID:                     types.StringValue(fmt.Sprintf("%v", deployment.ID)),
		Name:                   types.StringValue(deployment.Name),
		PlanKey:                types.StringValue(deployment.PlanKey.Key),
		Description:            util.NullString(deployment.Description),
		RepositorySpecsManaged: types.BoolValue(deployment.RepositorySpecsManaged),
		Repositories:           plan.Repositories,
		AssignmentVersion:      plan.AssignmentVersion,
		Assignments:            plan.Assignments,
		ComputedUsers:          assignmentResult.ComputedUsers,
		ComputedGroups:         assignmentResult.ComputedGroups,
	}
}
