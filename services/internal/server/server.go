package server

import (
	"github.com/google/wire"

	"github.com/example/scaffold/services/internal/webui"
)

// ProviderSet is the transport layer provider set.
var ProviderSet = wire.NewSet(
	NewHTTPServer,
	NewGRPCServer,
	webui.NewHandler,
)
