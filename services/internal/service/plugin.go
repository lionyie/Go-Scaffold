package service

import (
	"context"

	v1 "github.com/example/scaffold/services/api/plugin/v1"
	"github.com/example/scaffold/services/internal/biz"
	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
	"github.com/example/scaffold/services/internal/service/assembler"
)

// PluginService adapts the plugin registry to the generated transport interfaces.
type PluginService struct {
	v1.UnimplementedPluginServiceServer
	uc *biz.PluginUsecase
}

// NewPluginService creates the plugin service.
func NewPluginService(uc *biz.PluginUsecase) *PluginService {
	return &PluginService{uc: uc}
}

func (s *PluginService) ListPlugins(ctx context.Context, _ *v1.ListPluginsRequest) (*v1.ListPluginsResponse, error) {
	userID, admin := "", false
	if principal, ok := authpkg.PrincipalFromContext(ctx); ok {
		userID, admin = principal.UserID, principal.IsAdmin
	}
	items, err := s.uc.List(ctx, userID, admin)
	if err != nil {
		return nil, err
	}
	out := make([]*v1.Plugin, 0, len(items))
	for _, item := range items {
		out = append(out, assembler.PluginToProto(item))
	}
	return &v1.ListPluginsResponse{Items: out}, nil
}

func (s *PluginService) GetPlugin(ctx context.Context, req *v1.GetPluginRequest) (*v1.GetPluginResponse, error) {
	item, err := s.uc.Get(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &v1.GetPluginResponse{Plugin: assembler.PluginToProto(item)}, nil
}
