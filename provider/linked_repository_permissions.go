package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
)

type LinkedRepositoryPermissionsReceiver interface {
	getClient() *bamboo.Client
}

type LinkedRepositoryPermissionInterface interface {
	getAssignment(ctx context.Context) (Assignments, diag.Diagnostics)
	getLinkedRepositoryId(ctx context.Context) int
}

func CreateLinkedRepositoryAssignments(ctx context.Context, receiver LinkedRepositoryPermissionsReceiver, plan LinkedRepositoryPermissionInterface) (*AssignmentResult, diag.Diagnostics) {
	assignments, diags := plan.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	deploymentId := plan.getLinkedRepositoryId(ctx)

	_ = receiver.getClient().RepositoryService().UpdateRolePermissions(deploymentId, "LOGGED_IN", make([]string, 0))

	return ApplyNewAssignmentSet(ctx, receiver.getClient().UserService(),
		*assignmentOrder,
		func(user string) (*bamboo.UserPermission, error) {
			return receiver.getClient().RepositoryService().FindAvailableUser(deploymentId, user)
		},
		func(group string) (*bamboo.GroupPermission, error) {
			return receiver.getClient().RepositoryService().FindAvailableGroup(deploymentId, group)
		},
		func(user string, requestedPermissions []string) error {
			return receiver.getClient().RepositoryService().UpdateUserPermissions(deploymentId, user, requestedPermissions)
		},
		func(group string, requestedPermissions []string) error {
			return receiver.getClient().RepositoryService().UpdateGroupPermissions(deploymentId, group, requestedPermissions)
		},
	)
}

func ComputeLinkedRepositoryAssignments(ctx context.Context, receiver LinkedRepositoryPermissionsReceiver, state LinkedRepositoryPermissionInterface) (*AssignmentResult, diag.Diagnostics) {
	assignments, diags := state.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	deploymentId := state.getLinkedRepositoryId(ctx)
	assignedPermissions, err := receiver.getClient().RepositoryService().ReadPermissions(deploymentId)
	if err != nil {
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Failed to read deployment permissions", err.Error())}
	}

	return ComputeAssignment(ctx, assignedPermissions, *assignmentOrder)
}

func UpdateLinkedRepositoryAssignments(ctx context.Context, receiver LinkedRepositoryPermissionsReceiver,
	plan LinkedRepositoryPermissionInterface,
	state LinkedRepositoryPermissionInterface,
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
	deploymentId := state.getLinkedRepositoryId(ctx)

	return UpdateAssignment(ctx, receiver.getClient().UserService(),
		*inStateAssignmentOrder,
		*plannedAssignmentOrder,
		forceUpdate,
		func(user string) (*bamboo.UserPermission, error) {
			return receiver.getClient().RepositoryService().FindAvailableUser(deploymentId, user)
		},
		func(group string) (*bamboo.GroupPermission, error) {
			return receiver.getClient().RepositoryService().FindAvailableGroup(deploymentId, group)
		},
		func(user string, requestedPermissions []string) error {
			return receiver.getClient().RepositoryService().UpdateUserPermissions(deploymentId, user, requestedPermissions)
		},
		func(group string, requestedPermissions []string) error {
			return receiver.getClient().RepositoryService().UpdateGroupPermissions(deploymentId, group, requestedPermissions)
		},
	)
}

func DeleteLinkedRepositoryAssignments(ctx context.Context, receiver LinkedRepositoryPermissionsReceiver, state LinkedRepositoryPermissionInterface) diag.Diagnostics {
	assignments, diags := state.getAssignment(ctx)
	if diags != nil {
		return diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return diags
	}

	deploymentId := state.getLinkedRepositoryId(ctx)

	assignedPermissions, err := receiver.getClient().RepositoryService().ReadPermissions(deploymentId)
	if err != nil {
		return []diag.Diagnostic{diag.NewErrorDiagnostic("Failed to read deployment permissions", err.Error())}
	}

	return RemoveAssignment(ctx, assignedPermissions, assignmentOrder,
		func(user string) (*bamboo.UserPermission, error) {
			return receiver.getClient().RepositoryService().FindAvailableUser(deploymentId, user)
		},
		func(group string) (*bamboo.GroupPermission, error) {
			return receiver.getClient().RepositoryService().FindAvailableGroup(deploymentId, group)
		},
		func(user string, requestedPermissions []string) error {
			return receiver.getClient().RepositoryService().UpdateUserPermissions(deploymentId, user, requestedPermissions)
		},
		func(group string, requestedPermissions []string) error {
			return receiver.getClient().RepositoryService().UpdateGroupPermissions(deploymentId, group, requestedPermissions)
		})
}
