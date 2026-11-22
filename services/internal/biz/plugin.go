package biz

import (
	"context"

	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
)

// PluginUsecase exposes the plugin catalog to the service layer.
type PluginUsecase struct {
	lister PluginLister
}

// NewPluginUsecase wires the plugin registry into the usecase.
func NewPluginUsecase(lister PluginLister) *PluginUsecase {
	return &PluginUsecase{lister: lister}
}

func (uc *PluginUsecase) List(ctx context.Context, principal *authpkg.Principal) ([]*PluginInfo, error) {
	return uc.lister.ListPlugins(ctx, principal)
}

func (uc *PluginUsecase) Get(ctx context.Context, id string) (*PluginInfo, error) {
	return uc.lister.GetPlugin(ctx, id)
}

func (uc *PluginUsecase) InstallFromGit(ctx context.Context, rawURL, ref string, replace bool) (*PluginInfo, error) {
	return uc.lister.InstallFromGit(ctx, rawURL, ref, replace)
}

func (uc *PluginUsecase) InstallFromArchive(ctx context.Context, data []byte, replace bool) (*PluginInfo, error) {
	return uc.lister.InstallFromArchive(ctx, data, replace)
}

func (uc *PluginUsecase) Uninstall(ctx context.Context, id string) error {
	return uc.lister.Uninstall(ctx, id)
}
