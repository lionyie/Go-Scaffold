package service

import (
	"context"

	v1 "github.com/example/scaffold/services/api/identity/v1"
	"github.com/example/scaffold/services/internal/server/authz"
)

// AuthzService exposes the operation catalog and runtime permission bindings.
type AuthzService struct {
	v1.UnimplementedAuthzServiceServer
	manager *authz.Manager
}

// NewAuthzService creates the authorization service.
func NewAuthzService(manager *authz.Manager) *AuthzService {
	return &AuthzService{manager: manager}
}

func (s *AuthzService) ListOperations(ctx context.Context, _ *v1.ListOperationsRequest) (*v1.ListOperationsResponse, error) {
	items, err := s.manager.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.OperationInfo, 0, len(items))
	for _, item := range items {
		out = append(out, operationToProto(item))
	}
	return &v1.ListOperationsResponse{Items: out}, nil
}

func (s *AuthzService) SetOperationPermission(ctx context.Context, req *v1.SetOperationPermissionRequest) (*v1.SetOperationPermissionResponse, error) {
	view, err := s.manager.Set(ctx, req.GetOperation(), req.GetPermission())
	if err != nil {
		return nil, err
	}
	return &v1.SetOperationPermissionResponse{Operation: operationToProto(view)}, nil
}

func (s *AuthzService) ResetOperationPermission(ctx context.Context, req *v1.ResetOperationPermissionRequest) (*v1.ResetOperationPermissionResponse, error) {
	view, err := s.manager.Reset(ctx, req.GetOperation())
	if err != nil {
		return nil, err
	}
	return &v1.ResetOperationPermissionResponse{Operation: operationToProto(view)}, nil
}

func operationToProto(view authz.OperationView) *v1.OperationInfo {
	return &v1.OperationInfo{
		Operation:        view.Operation,
		Service:          view.Service,
		Method:           view.Method,
		Path:             view.Path,
		Permission:       view.Permission,
		StaticPermission: view.StaticPermission,
		Source:           view.Source,
		Locked:           view.Locked,
	}
}
