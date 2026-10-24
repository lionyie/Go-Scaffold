package service

import (
	"context"

	v1 "github.com/example/scaffold/services/api/identity/v1"
	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/service/assembler"
)

// RoleService adapts role usecases to the generated transport interfaces.
type RoleService struct {
	v1.UnimplementedRoleServiceServer
	uc *biz.RoleUsecase
}

// NewRoleService creates the role service.
func NewRoleService(uc *biz.RoleUsecase) *RoleService {
	return &RoleService{uc: uc}
}

func (s *RoleService) ListRoles(ctx context.Context, req *v1.ListRolesRequest) (*v1.ListRolesResponse, error) {
	items, total, err := s.uc.List(ctx, biz.ListOptions{
		PageSize:  int(req.GetPageSize()),
		PageToken: req.GetPageToken(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListRolesResponse{Items: assembler.RolesToProto(items), TotalSize: int32(total)}, nil
}

func (s *RoleService) CreateRole(ctx context.Context, req *v1.CreateRoleRequest) (*v1.CreateRoleResponse, error) {
	role, err := s.uc.Create(ctx, assembler.RoleFromProto(req.GetRole()))
	if err != nil {
		return nil, err
	}
	return &v1.CreateRoleResponse{Role: assembler.RoleToProto(role)}, nil
}

func (s *RoleService) UpdateRole(ctx context.Context, req *v1.UpdateRoleRequest) (*v1.UpdateRoleResponse, error) {
	role := assembler.RoleFromProto(req.GetRole())
	if role == nil {
		return nil, biz.ErrInvalidArgument("role is required")
	}
	role.ID = req.GetId()
	updated, err := s.uc.Update(ctx, role)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateRoleResponse{Role: assembler.RoleToProto(updated)}, nil
}

func (s *RoleService) DeleteRole(ctx context.Context, req *v1.DeleteRoleRequest) (*v1.DeleteRoleResponse, error) {
	if err := s.uc.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &v1.DeleteRoleResponse{}, nil
}

func (s *RoleService) ListPermissions(ctx context.Context, _ *v1.ListPermissionsRequest) (*v1.ListPermissionsResponse, error) {
	items, err := s.uc.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListPermissionsResponse{Items: assembler.PermissionsToProto(items)}, nil
}
