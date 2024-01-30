package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
)

type ProjectLinkedRepositoryModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	RssEnabled types.Bool   `tfsdk:"rss_enabled"`

	Owner   types.String `tfsdk:"owner"`
	Project types.String `tfsdk:"project"`
	Slug    types.String `tfsdk:"slug"`

	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       types.List   `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ LinkedRepositoryPermissionInterface = &ProjectLinkedRepositoryModel{}

func (d ProjectLinkedRepositoryModel) getAssignment(ctx context.Context) (Assignments, diag.Diagnostics) {
	var assignments Assignments = make([]Assignment, 0)

	diags := d.Assignments.ElementsAs(ctx, &assignments, true)
	return assignments, diags
}

func (d ProjectLinkedRepositoryModel) getLinkedRepositoryId(ctx context.Context) int {
	deploymentId, _ := strconv.Atoi(d.ID.ValueString())
	return deploymentId
}

func NewProjectLinkedRepositoryModel(plan ProjectLinkedRepositoryModel, repositoryId int, assignmentResult *AssignmentResult) *ProjectLinkedRepositoryModel {
	return &ProjectLinkedRepositoryModel{
		ID:                types.StringValue(fmt.Sprintf("%v", repositoryId)),
		Owner:             plan.Owner,
		Name:              plan.Name,
		RssEnabled:        plan.RssEnabled,
		Project:           plan.Project,
		Slug:              plan.Slug,
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     assignmentResult.ComputedUsers,
		ComputedGroups:    assignmentResult.ComputedGroups,
	}
}
