package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type AgentAssignmentModel struct {
	Type         types.String `tfsdk:"type"`
	AgentId      types.Int64  `tfsdk:"agent"`
	DeploymentId types.Int64  `tfsdk:"deployment"`
}

//func NewProjectVariableModel(plan ProjectVariableModel, project *bamboo.Project, assignmentResult *AssignmentResult) *ProjectVariableModel {
//	return &ProjectVariableModel{
//		Key:               types.StringValue(project.Key),
//		Name:              types.StringValue(project.Name),
//		Value:       types.StringValue(project.Value),
//		Secret: types.StringValue(project.Value),
//		Assignments:       plan.Assignments,
//		ComputedUsers:     assignmentResult.ComputedUsers,
//		ComputedGroups:    assignmentResult.ComputedGroups,
//	}
//}
