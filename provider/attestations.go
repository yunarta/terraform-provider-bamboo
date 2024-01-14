package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yunarta/terraform-atlassian-api-client/bamboo"
	"github.com/yunarta/terraform-provider-commons/util"
)

func CreateAttestation(ctx context.Context, permissions *bamboo.ObjectPermission, diagnostics *diag.Diagnostics) (basetypes.MapValue, basetypes.MapValue, diag.Diagnostics) {
	var userPermissionsMap = make(map[string][]string)
	var groupPermissionsMap = make(map[string][]string)
	for _, user := range permissions.Users {
		for _, permission := range user.Permissions {
			userInPermission, ok := userPermissionsMap[permission]
			if !ok {
				userInPermission = make([]string, 0)
				userPermissionsMap[permission] = userInPermission
			}

			userInPermission = append(userInPermission, user.Name)
			userPermissionsMap[permission] = userInPermission
		}
	}

	for _, group := range permissions.Groups {
		for _, permission := range group.Permissions {
			groupInPermission, ok := groupPermissionsMap[permission]
			if !ok {
				groupInPermission = make([]string, 0)
			}

			groupInPermission = append(groupInPermission, group.Name)
			groupPermissionsMap[permission] = groupInPermission
		}
	}

	users, diags := types.MapValueFrom(ctx, types.ListType{
		ElemType: types.StringType,
	}, userPermissionsMap)
	if util.TestDiagnostic(diagnostics, diags) {
		return basetypes.MapValue{}, basetypes.MapValue{}, diags
	}

	groups, diags := types.MapValueFrom(ctx, types.ListType{
		ElemType: types.StringType,
	}, groupPermissionsMap)
	if util.TestDiagnostic(diagnostics, diags) {
		return basetypes.MapValue{}, basetypes.MapValue{}, diags
	}

	return users, groups, nil
}
