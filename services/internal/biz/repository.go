package biz

import (
	"context"

	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
)

// ListOptions is the transport-independent query shape shared by repositories.
type ListOptions struct {
	PageSize  int
	PageToken string
	Keyword   string
}

// User is the platform account used for authentication and authorization.
type User struct {
	ID                 string
	Username           string
	DisplayName        string
	Email              string
	PasswordHash       string
	IsAdmin            bool
	MustChangePassword bool
	Status             string
	Roles              []string
	CreatedAt          int64
}

// UserRepo is the storage contract for platform accounts.
type UserRepo interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, opts ListOptions) ([]*User, int, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id string) error
}

// Role is a named collection of permission codes.
type Role struct {
	ID          string
	Name        string
	DisplayName string
	Description string
	IsSystem    bool
	Permissions []string
}

// Permission source values. Builtin permissions come from the generated
// catalog, plugin permissions from plugin manifests, and custom permissions
// from the admin API.
const (
	PermissionSourceBuiltin = "builtin"
	PermissionSourcePlugin  = "plugin"
	PermissionSourceCustom  = "custom"
)

// Permission is a single authorization capability.
type Permission struct {
	Code     string
	Name     string
	Resource string
	Action   string
	Source   string
}

// RoleRepo is the storage contract for roles and the permission catalog.
type RoleRepo interface {
	List(ctx context.Context, opts ListOptions) ([]*Role, int, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	Create(ctx context.Context, role *Role) (*Role, error)
	Update(ctx context.Context, role *Role) (*Role, error)
	Delete(ctx context.Context, id string) error
	ListPermissions(ctx context.Context) ([]*Permission, error)
	CreatePermission(ctx context.Context, permission *Permission) (*Permission, error)
	UpdatePermission(ctx context.Context, permission *Permission) (*Permission, error)
	DeletePermission(ctx context.Context, code string) error
	UpsertPermissions(ctx context.Context, permissions []*Permission) error
	AssignRole(ctx context.Context, userID, roleName string) error
}

// PolicyRepo stores runtime overrides that bind an HTTP operation to a
// permission code. The compiled mapping remains the fallback.
type PolicyRepo interface {
	ListOperationPolicies(ctx context.Context) (map[string]string, error)
	GetOperationPolicy(ctx context.Context, operation string) (string, bool, error)
	SetOperationPolicy(ctx context.Context, operation, permissionCode string) error
	DeleteOperationPolicy(ctx context.Context, operation string) error
	PermissionExists(ctx context.Context, code string) (bool, error)
}

// RefreshToken is a persisted, revocable refresh token handle. The raw token is
// never stored, only its SHA-256 hash.
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	SessionID string
	ExpiresAt int64
	Revoked   bool
}

// RefreshTokenRepo persists refresh token hashes.
type RefreshTokenRepo interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*RefreshToken, error)
	RevokeByHash(ctx context.Context, hash string) error
	RevokeBySession(ctx context.Context, userID, sessionID string) error
}

// PluginNav is a sidebar entry contributed by a plugin.
type PluginNav struct {
	Label      string
	Icon       string
	Order      int32
	Path       string
	Permission string
}

// PluginRoute is a client-side route contributed by a plugin.
type PluginRoute struct {
	Path       string
	Name       string
	Component  string
	Permission string
}

// PluginInfo is the transport-independent view of a plugin manifest.
type PluginInfo struct {
	ID          string
	Name        string
	Version     string
	Description string
	Icon        string
	EntryURL    string
	Styles      []string
	Routes      []PluginRoute
	Navigations []PluginNav
	Permissions []string
	Source      string
}

// PluginLister is implemented by the plugin registry.
type PluginLister interface {
	ListPlugins(ctx context.Context, principal *authpkg.Principal) ([]*PluginInfo, error)
	GetPlugin(ctx context.Context, id string) (*PluginInfo, error)
	InstallFromGit(ctx context.Context, rawURL, ref string, replace bool) (*PluginInfo, error)
	InstallFromArchive(ctx context.Context, data []byte, replace bool) (*PluginInfo, error)
	Uninstall(ctx context.Context, id string) error
}

// HealthChecker reports the state of the backing stores for readiness probes.
type HealthChecker interface {
	Check(ctx context.Context) (map[string]string, error)
}
