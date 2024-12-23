package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

type LinkedRepositoryModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	RssEnabled types.Bool   `tfsdk:"rss_enabled"`

	Project types.String `tfsdk:"project"`
	Slug    types.String `tfsdk:"slug"`
	Branch  types.String `tfsdk:"branch"`

	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       types.List   `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ LinkedRepositoryPermissionInterface = &LinkedRepositoryModel{}

func (d LinkedRepositoryModel) getAssignment(ctx context.Context) (Assignments, diag.Diagnostics) {
	var assignments Assignments = make([]Assignment, 0)

	diags := d.Assignments.ElementsAs(ctx, &assignments, true)
	return assignments, diags
}

func (d LinkedRepositoryModel) getLinkedRepositoryId(ctx context.Context) int {
	deploymentId, _ := strconv.Atoi(d.ID.ValueString())
	return deploymentId
}

func NewLinkedRepositoryModel(plan LinkedRepositoryModel, repositoryId int, assignmentResult *AssignmentResult) *LinkedRepositoryModel {
	return &LinkedRepositoryModel{
		ID:                types.StringValue(fmt.Sprintf("%v", repositoryId)),
		Name:              plan.Name,
		RssEnabled:        plan.RssEnabled,
		Project:           plan.Project,
		Slug:              plan.Slug,
		Branch:            plan.Branch,
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     assignmentResult.ComputedUsers,
		ComputedGroups:    assignmentResult.ComputedGroups,
	}
}
