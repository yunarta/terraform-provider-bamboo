package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
)

type ProjectPermissionsReceiver interface {
	getClient() *bamboo.Client
}

type ProjectPermissionInterface interface {
	getAssignment(ctx context.Context) (Assignments, diag.Diagnostics)
	getProjectKey(ctx context.Context) string
}

func CreateProjectAssignments(ctx context.Context, receiver ProjectPermissionsReceiver, plan ProjectPermissionInterface) (*AssignmentResult, diag.Diagnostics) {
	assignments, diags := plan.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	projectKey := plan.getProjectKey(ctx)

	_ = receiver.getClient().ProjectService().UpdateRolePermissions(projectKey, "LOGGED_IN", make([]string, 0))
	_ = receiver.getClient().ProjectService().UpdateRolePermissions(projectKey, "ANONYMOUS", make([]string, 0))

	return ApplyNewAssignmentSet(ctx, receiver.getClient().UserService(),
		*assignmentOrder,
		func(user string) (*bamboo.UserPermission, error) {
			return receiver.getClient().ProjectService().FindAvailableUser(projectKey, user)
		},
		func(group string) (*bamboo.GroupPermission, error) {
			return receiver.getClient().ProjectService().FindAvailableGroup(projectKey, group)
		},
		func(user string, requestedPermissions []string) error {
			return receiver.getClient().ProjectService().UpdateUserPermissions(projectKey, user, requestedPermissions)
		},
		func(group string, requestedPermissions []string) error {
			return receiver.getClient().ProjectService().UpdateGroupPermissions(projectKey, group, requestedPermissions)
		},
	)
}

func ComputeProjectAssignments(ctx context.Context, receiver ProjectPermissionsReceiver, state ProjectPermissionInterface) (*AssignmentResult, diag.Diagnostics) {
	assignments, diags := state.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	projectKey := state.getProjectKey(ctx)
	assignedPermissions, err := receiver.getClient().ProjectService().ReadPermissions(projectKey)
	if err != nil {
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Failed to read Project permissions", err.Error())}
	}

	return ComputeAssignment(ctx, assignedPermissions, *assignmentOrder)
}

func UpdateProjectAssignments(ctx context.Context, receiver ProjectPermissionsReceiver,
	plan ProjectPermissionInterface,
	state ProjectPermissionInterface,
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

	// the plan does not have computed value Project ID
	projectKey := state.getProjectKey(ctx)

	return UpdateAssignment(ctx, receiver.getClient().UserService(),
		*inStateAssignmentOrder,
		*plannedAssignmentOrder,
		forceUpdate,
		func(user string) (*bamboo.UserPermission, error) {
			return receiver.getClient().ProjectService().FindAvailableUser(projectKey, user)
		},
		func(group string) (*bamboo.GroupPermission, error) {
			return receiver.getClient().ProjectService().FindAvailableGroup(projectKey, group)
		},
		func(user string, requestedPermissions []string) error {
			return receiver.getClient().ProjectService().UpdateUserPermissions(projectKey, user, requestedPermissions)
		},
		func(group string, requestedPermissions []string) error {
			return receiver.getClient().ProjectService().UpdateGroupPermissions(projectKey, group, requestedPermissions)
		},
	)
}

func DeleteProjectAssignments(ctx context.Context, receiver ProjectPermissionsReceiver, state ProjectPermissionInterface) diag.Diagnostics {
	assignments, diags := state.getAssignment(ctx)
	if diags != nil {
		return diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return diags
	}

	projectKey := state.getProjectKey(ctx)

	assignedPermissions, err := receiver.getClient().ProjectService().ReadPermissions(projectKey)
	if err != nil {
		return []diag.Diagnostic{diag.NewErrorDiagnostic("Failed to read Project permissions", err.Error())}
	}

	return RemoveAssignment(ctx, assignedPermissions, assignmentOrder,
		func(user string) (*bamboo.UserPermission, error) {
			return receiver.getClient().ProjectService().FindAvailableUser(projectKey, user)
		},
		func(group string) (*bamboo.GroupPermission, error) {
			return receiver.getClient().ProjectService().FindAvailableGroup(projectKey, group)
		},
		func(user string, requestedPermissions []string) error {
			return receiver.getClient().ProjectService().UpdateUserPermissions(projectKey, user, requestedPermissions)
		},
		func(group string, requestedPermissions []string) error {
			return receiver.getClient().ProjectService().UpdateGroupPermissions(projectKey, group, requestedPermissions)
		})
}
