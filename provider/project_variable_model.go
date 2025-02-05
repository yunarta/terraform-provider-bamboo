package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProjectVariableModel0 struct {
	Key    types.String `tfsdk:"key"`
	Name   types.String `tfsdk:"name"`
	Value  types.String `tfsdk:"value"`
	Secret types.String `tfsdk:"secret"`
}

type ProjectVariableModel struct {
	Project types.String `tfsdk:"project"`
	Name    types.String `tfsdk:"name"`
	Value   types.String `tfsdk:"value"`
	Secret  types.String `tfsdk:"secret"`
}

func FromProjectVariableModel0(plan ProjectVariableModel0) *ProjectVariableModel {
	return &ProjectVariableModel{
		Project: plan.Key,
		Name:    plan.Name,
		Value:   plan.Value,
		Secret:  plan.Secret,
	}
}
