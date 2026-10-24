package server

import (
	"log/slog"

	"github.com/go-kratos/kratos/v3/middleware/logging"
	"github.com/go-kratos/kratos/v3/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v3/transport/http"

	identityv1 "github.com/example/scaffold/services/api/identity/v1"
	pluginv1 "github.com/example/scaffold/services/api/plugin/v1"
	systemv1 "github.com/example/scaffold/services/api/system/v1"
	"github.com/example/scaffold/services/internal/auth/oauth"
	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/conf"
	extregistry "github.com/example/scaffold/services/internal/extensions/registry"
	_ "github.com/example/scaffold/services/internal/extensions/registry_gen"
	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
	"github.com/example/scaffold/services/internal/pkg/rbac"
	"github.com/example/scaffold/services/internal/plugin"
	"github.com/example/scaffold/services/internal/server/middleware"
	"github.com/example/scaffold/services/internal/service"
	"github.com/example/scaffold/services/internal/webui"
)

// NewHTTPServer builds the HTTP transport with API routes, plugin routes, and
// the SPA fallback.
func NewHTTPServer(
	bc *conf.Bootstrap,
	svcs *service.Services,
	registry *plugin.Registry,
	web *webui.Handler,
	tokens *authpkg.TokenManager,
	checker rbac.Checker,
	oauthStore oauth.Store,
	oauthUC *biz.OAuthUsecase,
	log *slog.Logger,
) *khttp.Server {
	cfg := bc.GetServer().GetHttp()
	opts := []khttp.ServerOption{
		khttp.RequestDecoder(protojsonRequestDecoder),
		khttp.ResponseEncoder(protojsonResponseEncoder),
		khttp.Middleware(
			recovery.Recovery(),
			logging.Server(log),
			middleware.LoginRateLimit(),
			middleware.Authn(tokens),
			middleware.Authz(checker),
		),
		khttp.NotFoundHandler(web),
	}
	if cfg.GetNetwork() != "" {
		opts = append(opts, khttp.Network(cfg.GetNetwork()))
	}
	if cfg.GetAddr() != "" {
		opts = append(opts, khttp.Address(cfg.GetAddr()))
	}
	if timeout := cfg.GetTimeout().AsDuration(); timeout > 0 {
		opts = append(opts, khttp.Timeout(timeout))
	}

	srv := khttp.NewServer(opts...)
	identityv1.RegisterAuthServiceHTTPServer(srv, svcs.Auth)
	identityv1.RegisterUserServiceHTTPServer(srv, svcs.User)
	identityv1.RegisterRoleServiceHTTPServer(srv, svcs.Role)
	pluginv1.RegisterPluginServiceHTTPServer(srv, svcs.Plugin)
	systemv1.RegisterSystemServiceHTTPServer(srv, svcs.System)

	// HandlePrefix bypasses the Kratos middleware chain; the registry performs
	// its own authentication and authorization for /api requests.
	srv.HandlePrefix("/proxy/plugins/", registry.Handler())
	srv.HandlePrefix("/api/v1/auth/oauth/", NewOAuthHandler(bc, oauthStore, oauthUC, tokens))
	extregistry.RegisterAll(srv)
	return srv
}
