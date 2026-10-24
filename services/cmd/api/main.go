package main

import (
	"flag"
	"log/slog"
	"os"
	"runtime"

	"github.com/go-kratos/kratos/v3"
	"github.com/go-kratos/kratos/v3/config"
	"github.com/go-kratos/kratos/v3/config/env"
	"github.com/go-kratos/kratos/v3/config/file"
	"github.com/go-kratos/kratos/v3/log"
	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	"github.com/example/scaffold/services/internal/conf"
	"github.com/example/scaffold/services/internal/service"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the compiled software name.
	Name = "scaffold"
	// Version is injected at build time.
	Version = "dev"
	// flagconf is the config path.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../configs/config.local.yaml", "config path, eg: -conf ../configs/config.local.yaml")
}

func newApp(logger *slog.Logger, gs *grpc.Server, hs *http.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(gs, hs),
	)
}

func main() {
	flag.Parse()

	logger := log.NewLogger(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	})).With(
		slog.String("service.id", id),
		slog.String("service.name", Name),
		slog.String("service.version", Version),
	)
	log.SetDefault(logger)

	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
			env.NewSource("SCAFFOLD"),
		),
		config.WithResolveActualTypes(true),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	if os.Getenv("APP_ENV") == "production" {
		if secret := bc.GetAuth().GetJwt().GetSecret(); secret == "" || secret == "dev-only-change-me" {
			panic("production requires SCAFFOLD_JWT_SECRET to be set to a non-default value")
		}
		if secret := bc.GetAuth().GetOauth().GetStateSecret(); secret == "" || secret == "dev-only-oauth-state-secret" {
			panic("production requires SCAFFOLD_OAUTH_STATE_SECRET to be set to a non-default value")
		}
	}

	build := &service.BuildInfo{Version: Version, GoVersion: runtime.Version()}
	app, cleanup, err := wireApp(&bc, logger, build)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
