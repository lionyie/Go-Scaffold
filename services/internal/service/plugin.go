package service

import (
	"context"
	"encoding/base64"

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
	var principal *authpkg.Principal
	if p, ok := authpkg.PrincipalFromContext(ctx); ok {
		principal = p
	}
	items, err := s.uc.List(ctx, principal)
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

func (s *PluginService) InstallPlugin(ctx context.Context, req *v1.InstallPluginRequest) (*v1.InstallPluginResponse, error) {
	var (
		item *biz.PluginInfo
		err  error
	)
	switch {
	case req.GetGitUrl() != "":
		item, err = s.uc.InstallFromGit(ctx, req.GetGitUrl(), req.GetGitRef(), req.GetReplace())
	case req.GetArchiveBase64() != "":
		data, decodeErr := base64.StdEncoding.DecodeString(req.GetArchiveBase64())
		if decodeErr != nil {
			return nil, biz.ErrInvalidArgument("invalid archive encoding")
		}
		item, err = s.uc.InstallFromArchive(ctx, data, req.GetReplace())
	default:
		return nil, biz.ErrInvalidArgument("git_url or archive_base64 is required")
	}
	if err != nil {
		return nil, err
	}
	return &v1.InstallPluginResponse{Plugin: assembler.PluginToProto(item)}, nil
}

func (s *PluginService) UninstallPlugin(ctx context.Context, req *v1.UninstallPluginRequest) (*v1.UninstallPluginResponse, error) {
	if err := s.uc.Uninstall(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &v1.UninstallPluginResponse{}, nil
}
