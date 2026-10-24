package plugin

import (
	"github.com/google/wire"

	"github.com/example/scaffold/services/internal/biz"
)

// ProviderSet exposes the registry both concretely and as biz.PluginLister.
var ProviderSet = wire.NewSet(
	NewRegistry,
	wire.Bind(new(biz.PluginLister), new(*Registry)),
)
