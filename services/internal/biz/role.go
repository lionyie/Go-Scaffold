package biz

import (
	"context"
	"regexp"
	"strings"

	"github.com/example/scaffold/services/internal/pkg/id"
	"github.com/example/scaffold/services/internal/pkg/pagination"
)

var roleNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,38}$`)

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
