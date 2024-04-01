package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProjectVariableModel struct {
	Key    types.String `tfsdk:"key"`
	Name   types.String `tfsdk:"name"`
	Value  types.String `tfsdk:"value"`
	Secret types.String `tfsdk:"secret"`
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
