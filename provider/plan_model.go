package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
)

type PlanModel struct {
	RetainOnDelete types.Bool   `tfsdk:"retain_on_delete"`
	Id             types.Int64  `tfsdk:"id"`
	ProjectKey     types.String `tfsdk:"project"`
	Key            types.String `tfsdk:"key"`
	Name           types.String `tfsdk:"name"`
}

//var _ ProjectPermissionInterface = &ProjectModel{}
//
//func (d ProjectModel) getAssignment(ctx context.Context) (Assignments, diag.Diagnostics) {
//	var assignments Assignments = make([]Assignment, 0)
//
//	diags := d.Assignments.ElementsAs(ctx, &assignments, true)
//	return assignments, diags
//}
//
//func (d ProjectModel) getProjectKey(ctx context.Context) string {
//	return d.Key.ValueString()
//}

func NewPlanModel(plan PlanModel, bambooPlan *bamboo.Plan) *PlanModel {
	//id, _ := strconv.Atoi(bambooPlan.Id)
	return &PlanModel{
		RetainOnDelete: plan.RetainOnDelete,
		Id:             types.Int64Value(bambooPlan.Id),
		ProjectKey:     types.StringValue(bambooPlan.ProjectKey),
		Key:            types.StringValue(bambooPlan.ShortKey),
		Name:           plan.Name,
	}
}
