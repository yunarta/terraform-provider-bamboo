package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
)

type ProjectModel struct {
	RetainOnDelete    types.Bool   `tfsdk:"retain_on_delete"`
	Key               types.String `tfsdk:"key"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	AssignmentVersion types.String `tfsdk:"assignment_version"`
	Assignments       types.List   `tfsdk:"assignments"`
	ComputedUsers     types.List   `tfsdk:"computed_users"`
	ComputedGroups    types.List   `tfsdk:"computed_groups"`
}

var _ ProjectPermissionInterface = &ProjectModel{}

func (d ProjectModel) getAssignment(ctx context.Context) (Assignments, diag.Diagnostics) {
	var assignments Assignments = make([]Assignment, 0)

	diags := d.Assignments.ElementsAs(ctx, &assignments, true)
	return assignments, diags
}

func (d ProjectModel) getProjectKey(ctx context.Context) string {
	return d.Key.ValueString()
}

func NewProjectModel(plan ProjectModel, project *bamboo.Project, assignmentResult *AssignmentResult) *ProjectModel {
	return &ProjectModel{
		RetainOnDelete:    plan.RetainOnDelete,
		Key:               types.StringValue(project.Key),
		Name:              types.StringValue(project.Name),
		Description:       util.NullString(project.Description),
		AssignmentVersion: plan.AssignmentVersion,
		Assignments:       plan.Assignments,
		ComputedUsers:     assignmentResult.ComputedUsers,
		ComputedGroups:    assignmentResult.ComputedGroups,
	}
}
