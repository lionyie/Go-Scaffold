package biz

import (
	"context"
	"regexp"
	"strings"

	"github.com/example/scaffold/services/internal/pkg/id"
	"github.com/example/scaffold/services/internal/pkg/pagination"
)

var roleNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,38}$`)

// permissionCodePattern keeps custom codes composable as resource:action and
// safe to place in a URL path segment.
var permissionCodePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(:[a-z0-9][a-z0-9_-]*)+$`)

// RoleUsecase contains role and permission catalog rules.
type RoleUsecase struct {
	repo RoleRepo
}

// NewRoleUsecase wires the role repository into the usecase.
func NewRoleUsecase(repo RoleRepo) *RoleUsecase {
	return &RoleUsecase{repo: repo}
}

func (uc *RoleUsecase) List(ctx context.Context, opts ListOptions) ([]*Role, int, error) {
	opts.PageSize = pagination.Normalize(opts.PageSize)
	return uc.repo.List(ctx, opts)
}

func (uc *RoleUsecase) Create(ctx context.Context, role *Role) (*Role, error) {
	role.Name = strings.TrimSpace(role.Name)
	if !roleNamePattern.MatchString(role.Name) {
		return nil, ErrInvalidArgument("role name must match ^[a-z][a-z0-9-]{1,38}$")
	}
	if role.ID == "" {
		role.ID = id.New()
	}
	return uc.repo.Create(ctx, role)
}

func (uc *RoleUsecase) Update(ctx context.Context, role *Role) (*Role, error) {
	return uc.repo.Update(ctx, role)
}

func (uc *RoleUsecase) Delete(ctx context.Context, roleID string) error {
	return uc.repo.Delete(ctx, roleID)
}

func (uc *RoleUsecase) ListPermissions(ctx context.Context) ([]*Permission, error) {
	return uc.repo.ListPermissions(ctx)
}

func (uc *RoleUsecase) CreatePermission(ctx context.Context, permission *Permission) (*Permission, error) {
	if permission == nil {
		return nil, ErrInvalidArgument("permission is required")
	}
	permission.Code = strings.TrimSpace(permission.Code)
	if !permissionCodePattern.MatchString(permission.Code) {
		return nil, ErrInvalidArgument("permission code must match ^[a-z][a-z0-9]*(:[a-z0-9][a-z0-9_-]*)+$")
	}
	if err := normalizePermission(permission); err != nil {
		return nil, err
	}
	permission.Source = PermissionSourceCustom
	return uc.repo.CreatePermission(ctx, permission)
}

func (uc *RoleUsecase) UpdatePermission(ctx context.Context, permission *Permission) (*Permission, error) {
	if permission == nil || strings.TrimSpace(permission.Code) == "" {
		return nil, ErrInvalidArgument("permission code is required")
	}
	permission.Code = strings.TrimSpace(permission.Code)
	if err := normalizePermission(permission); err != nil {
		return nil, err
	}
	return uc.repo.UpdatePermission(ctx, permission)
}

func (uc *RoleUsecase) DeletePermission(ctx context.Context, code string) error {
	code = strings.TrimSpace(code)
	if code == "" {
		return ErrInvalidArgument("permission code is required")
	}
	return uc.repo.DeletePermission(ctx, code)
}

func normalizePermission(permission *Permission) error {
	permission.Name = strings.TrimSpace(permission.Name)
	if permission.Name == "" {
		return ErrInvalidArgument("permission name is required")
	}
	permission.Resource = strings.TrimSpace(permission.Resource)
	permission.Action = strings.TrimSpace(permission.Action)
	if permission.Resource == "" {
		resource, _, _ := strings.Cut(permission.Code, ":")
		permission.Resource = resource
	}
	if permission.Action == "" {
		_, action, _ := strings.Cut(permission.Code, ":")
		permission.Action = action
	}
	return nil
}
