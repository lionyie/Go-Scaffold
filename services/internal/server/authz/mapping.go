package authz

import (
	identityv1 "github.com/example/scaffold/services/api/identity/v1"
	pluginv1 "github.com/example/scaffold/services/api/plugin/v1"
	systemv1 "github.com/example/scaffold/services/api/system/v1"
	"github.com/example/scaffold/services/internal/pkg/rbac"
)

// Requirement describes how an operation is protected.
type Requirement struct {
	// Public operations skip authentication entirely.
	Public bool
	// Code is the permission code required. Empty means "authenticated only".
	Code string
}

// Requirements maps generated operation names to authorization rules. Adding an
// RPC without adding an entry makes the endpoint fail closed.
var Requirements = map[string]Requirement{
	identityv1.OperationAuthServiceLogin:          {Public: true},
	identityv1.OperationAuthServiceRefresh:        {Public: true},
	identityv1.OperationAuthServiceLogout:         {},
	identityv1.OperationAuthServiceGetCurrentUser: {},
	identityv1.OperationAuthServiceChangePassword: {},

	systemv1.OperationSystemServiceGetHealth:  {Public: true},
	systemv1.OperationSystemServiceGetReady:   {Public: true},
	systemv1.OperationSystemServiceGetVersion: {Public: true},

	identityv1.OperationUserServiceListUsers:  {Code: rbac.PermUserList},
	identityv1.OperationUserServiceCreateUser: {Code: rbac.PermUserCreate},
	identityv1.OperationUserServiceUpdateUser: {Code: rbac.PermUserUpdate},
	identityv1.OperationUserServiceDeleteUser: {Code: rbac.PermUserDelete},

	identityv1.OperationRoleServiceListRoles:       {Code: rbac.PermRoleList},
	identityv1.OperationRoleServiceCreateRole:      {Code: rbac.PermRoleCreate},
	identityv1.OperationRoleServiceUpdateRole:      {Code: rbac.PermRoleUpdate},
	identityv1.OperationRoleServiceDeleteRole:      {Code: rbac.PermRoleDelete},
	identityv1.OperationRoleServiceListPermissions: {Code: rbac.PermRolePermissionList},

	pluginv1.OperationPluginServiceListPlugins: {Public: true},
	pluginv1.OperationPluginServiceGetPlugin:   {Public: true},
}
