package server

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3/middleware/logging"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	kgrpc "github.com/go-kratos/kratos/v3/transport/grpc"

	identityv1 "github.com/example/scaffold/services/api/identity/v1"
	pluginv1 "github.com/example/scaffold/services/api/plugin/v1"
	systemv1 "github.com/example/scaffold/services/api/system/v1"
	"github.com/example/scaffold/services/internal/conf"
	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
	"github.com/example/scaffold/services/internal/pkg/rbac"
	"github.com/example/scaffold/services/internal/server/middleware"
	"github.com/example/scaffold/services/internal/service"
)

// NewGRPCServer builds the gRPC transport with the same auth middleware as HTTP.
func NewGRPCServer(
	bc *conf.Bootstrap,
	svcs *service.Services,
	tokens *authpkg.TokenManager,
	checker rbac.Checker,
	log *slog.Logger,
) *kgrpc.Server {
	cfg := bc.GetServer().GetGrpc()
	opts := []kgrpc.ServerOption{
		kgrpc.Middleware(
			recovery.Recovery(),
			logging.Server(log),
			middleware.LoginRateLimit(),
			middleware.Authn(tokens),
			middleware.Authz(checker),
		),
	}
	if cfg.GetNetwork() != "" {
		opts = append(opts, kgrpc.Network(cfg.GetNetwork()))
	}
	if cfg.GetAddr() != "" {
		opts = append(opts, kgrpc.Address(cfg.GetAddr()))
	}
	if timeout := cfg.GetTimeout().AsDuration(); timeout > 0 {
		opts = append(opts, kgrpc.Timeout(timeout))
	}

	srv := kgrpc.NewServer(opts...)
	identityv1.RegisterAuthServiceServer(srv, svcs.Auth)
	identityv1.RegisterUserServiceServer(srv, svcs.User)
	identityv1.RegisterRoleServiceServer(srv, svcs.Role)
	pluginv1.RegisterPluginServiceServer(srv, svcs.Plugin)
	systemv1.RegisterSystemServiceServer(srv, svcs.System)
	return srv
}
