package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
)

type PlanModel struct {
	RetainOnDelete types.Bool   `tfsdk:"retain_on_delete"`
	Id             types.Int64  `tfsdk:"id"`
	Key            types.String `tfsdk:"key"`
	PlanKey        types.String `tfsdk:"plan_key"`
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
		Key:            types.StringValue(bambooPlan.ProjectKey),
		PlanKey:        types.StringValue(bambooPlan.ShortKey),
		Name:           plan.Name,
	}
}
