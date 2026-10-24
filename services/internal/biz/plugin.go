package biz

import "context"

// PluginUsecase exposes the plugin catalog to the service layer.
type PluginUsecase struct {
	lister PluginLister
}

// NewPluginUsecase wires the plugin registry into the usecase.
func NewPluginUsecase(lister PluginLister) *PluginUsecase {
	return &PluginUsecase{lister: lister}
}

func (uc *PluginUsecase) List(ctx context.Context, userID string, admin bool) ([]*PluginInfo, error) {
	return uc.lister.ListPlugins(ctx, userID, admin)
}

func (uc *PluginUsecase) Get(ctx context.Context, id string) (*PluginInfo, error) {
	return uc.lister.GetPlugin(ctx, id)
}
