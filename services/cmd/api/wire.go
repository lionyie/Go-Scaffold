//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"log/slog"

	"github.com/google/wire"

	"github.com/go-kratos/kratos/v3"

	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/conf"
	"github.com/example/scaffold/services/internal/data"
	"github.com/example/scaffold/services/internal/plugin"
	"github.com/example/scaffold/services/internal/server"
	"github.com/example/scaffold/services/internal/service"
)

// wireApp initializes the Kratos application.
func wireApp(*conf.Bootstrap, *slog.Logger, *service.BuildInfo) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		plugin.ProviderSet,
		newApp,
	))
}
