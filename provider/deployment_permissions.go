package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
)

type DeploymentPermissionsReceiver interface {
	getClient() *bamboo.Client
}

type DeploymentPermissionInterface interface {
	getAssignment(ctx context.Context) (Assignments, diag.Diagnostics)
	getDeploymentId(ctx context.Context) int
}

func CreateDeploymentAssignments(ctx context.Context, receiver DeploymentPermissionsReceiver, plan DeploymentPermissionInterface) (*AssignmentResult, diag.Diagnostics) {
	assignments, diags := plan.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	deploymentId := plan.getDeploymentId(ctx)

	_ = receiver.getClient().DeploymentService().UpdateRolePermissions(deploymentId, "LOGGED_IN", make([]string, 0))
	_ = receiver.getClient().DeploymentService().UpdateRolePermissions(deploymentId, "ANONYMOUS", make([]string, 0))

	return ApplyNewAssignmentSet(ctx, receiver.getClient().UserService(),
		*assignmentOrder,
		func(user string) (*bamboo.UserPermission, error) {
			return receiver.getClient().DeploymentService().FindAvailableUser(deploymentId, user)
		},
		func(group string) (*bamboo.GroupPermission, error) {
			return receiver.getClient().DeploymentService().FindAvailableGroup(deploymentId, group)
		},
		func(user string, requestedPermissions []string) error {
			return receiver.getClient().DeploymentService().UpdateUserPermissions(deploymentId, user, requestedPermissions)
		},
		func(group string, requestedPermissions []string) error {
			return receiver.getClient().DeploymentService().UpdateGroupPermissions(deploymentId, group, requestedPermissions)
		},
	)
}

func ComputeDeploymentAssignments(ctx context.Context, receiver DeploymentPermissionsReceiver, state DeploymentPermissionInterface) (*AssignmentResult, diag.Diagnostics) {
	assignments, diags := state.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	deploymentId := state.getDeploymentId(ctx)
	assignedPermissions, err := receiver.getClient().DeploymentService().ReadPermissions(deploymentId)
	if err != nil {
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Failed to read deployment permissions", err.Error())}
	}

	return ComputeAssignment(ctx, assignedPermissions, *assignmentOrder)
}

func UpdateDeploymentAssignments(ctx context.Context, receiver DeploymentPermissionsReceiver,
	plan DeploymentPermissionInterface,
	state DeploymentPermissionInterface,
	forceUpdate bool) (*AssignmentResult, diag.Diagnostics) {

	plannedAssignments, diags := plan.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	inStateAssignments, diags := state.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	plannedAssignmentOrder, diags := plannedAssignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	inStateAssignmentOrder, diags := inStateAssignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	// the plan does not have computed value deployment ID
	deploymentId := state.getDeploymentId(ctx)

	return UpdateAssignment(ctx, receiver.getClient().UserService(),
		*inStateAssignmentOrder,
		*plannedAssignmentOrder,
		forceUpdate,
		func(user string) (*bamboo.UserPermission, error) {
			return receiver.getClient().DeploymentService().FindAvailableUser(deploymentId, user)
		},
		func(group string) (*bamboo.GroupPermission, error) {
			return receiver.getClient().DeploymentService().FindAvailableGroup(deploymentId, group)
		},
		func(user string, requestedPermissions []string) error {
			return receiver.getClient().DeploymentService().UpdateUserPermissions(deploymentId, user, requestedPermissions)
		},
		func(group string, requestedPermissions []string) error {
			return receiver.getClient().DeploymentService().UpdateGroupPermissions(deploymentId, group, requestedPermissions)
		},
	)
}

func DeleteDeploymentAssignments(ctx context.Context, receiver DeploymentPermissionsReceiver, state DeploymentPermissionInterface) diag.Diagnostics {
	assignments, diags := state.getAssignment(ctx)
	if diags != nil {
		return diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return diags
	}

	deploymentId := state.getDeploymentId(ctx)

	assignedPermissions, err := receiver.getClient().DeploymentService().ReadPermissions(deploymentId)
	if err != nil {
		return []diag.Diagnostic{diag.NewErrorDiagnostic("Failed to read deployment permissions", err.Error())}
	}

	return RemoveAssignment(ctx, assignedPermissions, assignmentOrder,
		func(user string) (*bamboo.UserPermission, error) {
			return receiver.getClient().DeploymentService().FindAvailableUser(deploymentId, user)
		},
		func(group string) (*bamboo.GroupPermission, error) {
			return receiver.getClient().DeploymentService().FindAvailableGroup(deploymentId, group)
		},
		func(user string, requestedPermissions []string) error {
			return receiver.getClient().DeploymentService().UpdateUserPermissions(deploymentId, user, requestedPermissions)
		},
		func(group string, requestedPermissions []string) error {
			return receiver.getClient().DeploymentService().UpdateGroupPermissions(deploymentId, group, requestedPermissions)
		})
}
