package biz

import (
	"time"

	"github.com/google/wire"

	"github.com/example/scaffold/services/internal/auth/oauth"
	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
	"github.com/example/scaffold/services/internal/pkg/rbac"
)

// NowMillis returns the current time as Unix milliseconds.
func NowMillis() int64 { return time.Now().UnixMilli() }

// ProviderSet is the domain layer provider set.
var ProviderSet = wire.NewSet(
	authpkg.NewTokenManager,
	rbac.NewChecker,
	oauth.NewMemoryStore,
	wire.Bind(new(oauth.Store), new(*oauth.MemoryStore)),
	NewAuthUsecase,
	NewUserUsecase,
	NewRoleUsecase,
	NewOAuthUsecase,
	NewPluginUsecase,
)
