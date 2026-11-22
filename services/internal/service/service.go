package service

import "github.com/google/wire"

// BuildInfo carries linker-provided build metadata.
type BuildInfo struct {
	Version   string
	GoVersion string
}

// Services aggregates every transport-facing service.
type Services struct {
	Auth   *AuthService
	User   *UserService
	Role   *RoleService
	Authz  *AuthzService
	Plugin *PluginService
	System *SystemService
}

// NewServices builds the aggregate used by the servers.
func NewServices(
	auth *AuthService,
	user *UserService,
	role *RoleService,
	authz *AuthzService,
	plugin *PluginService,
	system *SystemService,
) *Services {
	return &Services{
		Auth:   auth,
		User:   user,
		Role:   role,
		Authz:  authz,
		Plugin: plugin,
		System: system,
	}
}

// ProviderSet is the service layer provider set.
var ProviderSet = wire.NewSet(
	NewAuthService,
	NewUserService,
	NewRoleService,
	NewAuthzService,
	NewPluginService,
	NewSystemService,
	NewServices,
)
