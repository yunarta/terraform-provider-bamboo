package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
)

type ProjectLinkedRepositoryPermissionReceiver interface {
	getClient() *bamboo.Client
}

type ProjectLinkedRepositoryPermissionInterface interface {
	getAssignment(ctx context.Context) (Assignments, diag.Diagnostics)
	getLinkedRepositoryId(ctx context.Context) int
}

func CreateProjectLinkedRepositoryAssignments(ctx context.Context, receiver ProjectLinkedRepositoryPermissionReceiver, plan ProjectLinkedRepositoryPermissionInterface) (*AssignmentResult, diag.Diagnostics) {
	assignments, diags := plan.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	repositoryId := plan.getLinkedRepositoryId(ctx)

	_ = receiver.getClient().RepositoryService().UpdateRolePermissions(repositoryId, "LOGGED_IN", make([]string, 0))

	return ApplyNewAssignmentSet(ctx, receiver.getClient().UserService(),
		*assignmentOrder,
		func(user string) (*bamboo.UserPermission, error) {
			return receiver.getClient().RepositoryService().FindAvailableUser(repositoryId, user)
		},
		func(group string) (*bamboo.GroupPermission, error) {
			return receiver.getClient().RepositoryService().FindAvailableGroup(repositoryId, group)
		},
		func(user string, requestedPermissions []string) error {
			return receiver.getClient().RepositoryService().UpdateUserPermissions(repositoryId, user, requestedPermissions)
		},
		func(group string, requestedPermissions []string) error {
			return receiver.getClient().RepositoryService().UpdateGroupPermissions(repositoryId, group, requestedPermissions)
		},
	)
}

func ComputeProjectLinkedRepositoryAssignments(ctx context.Context, receiver ProjectLinkedRepositoryPermissionReceiver, state ProjectLinkedRepositoryPermissionInterface) (*AssignmentResult, diag.Diagnostics) {
	assignments, diags := state.getAssignment(ctx)
	if diags != nil {
		return nil, diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return nil, diags
	}

	repositoryId := state.getLinkedRepositoryId(ctx)
	assignedPermissions, err := receiver.getClient().RepositoryService().ReadPermissions(repositoryId)
	if err != nil {
		return nil, []diag.Diagnostic{diag.NewErrorDiagnostic("Failed to read deployment permissions", err.Error())}
	}

	return ComputeAssignment(ctx, assignedPermissions, *assignmentOrder)
}

func UpdateProjectLinkedRepositoryAssignments(ctx context.Context, receiver ProjectLinkedRepositoryPermissionReceiver,
	plan ProjectLinkedRepositoryPermissionInterface,
	state ProjectLinkedRepositoryPermissionInterface,
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
	repositoryId := state.getLinkedRepositoryId(ctx)

	return UpdateAssignment(ctx, receiver.getClient().UserService(),
		*inStateAssignmentOrder,
		*plannedAssignmentOrder,
		forceUpdate,
		func(user string) (*bamboo.UserPermission, error) {
			return receiver.getClient().RepositoryService().FindAvailableUser(repositoryId, user)
		},
		func(group string) (*bamboo.GroupPermission, error) {
			return receiver.getClient().RepositoryService().FindAvailableGroup(repositoryId, group)
		},
		func(user string, requestedPermissions []string) error {
			return receiver.getClient().RepositoryService().UpdateUserPermissions(repositoryId, user, requestedPermissions)
		},
		func(group string, requestedPermissions []string) error {
			return receiver.getClient().RepositoryService().UpdateGroupPermissions(repositoryId, group, requestedPermissions)
		},
	)
}

func DeleteProjectLinkedRepositoryAssignments(ctx context.Context, receiver ProjectLinkedRepositoryPermissionReceiver, state ProjectLinkedRepositoryPermissionInterface) diag.Diagnostics {
	assignments, diags := state.getAssignment(ctx)
	if diags != nil {
		return diags
	}

	assignmentOrder, diags := assignments.CreateAssignmentOrder(ctx)
	if diags != nil {
		return diags
	}

	repositoryId := state.getLinkedRepositoryId(ctx)

	assignedPermissions, err := receiver.getClient().RepositoryService().ReadPermissions(repositoryId)
	if err != nil {
		return []diag.Diagnostic{diag.NewErrorDiagnostic("Failed to read deployment permissions", err.Error())}
	}

	return RemoveAssignment(ctx, assignedPermissions, assignmentOrder,
		func(user string) (*bamboo.UserPermission, error) {
			return receiver.getClient().RepositoryService().FindAvailableUser(repositoryId, user)
		},
		func(group string) (*bamboo.GroupPermission, error) {
			return receiver.getClient().RepositoryService().FindAvailableGroup(repositoryId, group)
		},
		func(user string, requestedPermissions []string) error {
			return receiver.getClient().RepositoryService().UpdateUserPermissions(repositoryId, user, requestedPermissions)
		},
		func(group string, requestedPermissions []string) error {
			return receiver.getClient().RepositoryService().UpdateGroupPermissions(repositoryId, group, requestedPermissions)
		})
}
