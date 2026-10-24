package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"

	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
)

// Authn parses the bearer token when present. It deliberately does not reject
// anonymous requests: the authorization middleware decides per operation.
func Authn(tokens *authpkg.TokenManager) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if token := bearerToken(ctx); token != "" {
				if claims, err := tokens.ParseAccess(token); err == nil {
					ctx = authpkg.WithPrincipal(ctx, &authpkg.Principal{
						UserID:             claims.UserID,
						Username:           claims.Username,
						SessionID:          claims.SessionID,
						IsAdmin:            claims.Admin,
						MustChangePassword: claims.MustChangePassword,
					})
				}
			}
			return handler(ctx, req)
		}
	}
}

func bearerToken(ctx context.Context) string {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return ""
	}
	return authpkg.Bearer(tr.RequestHeader().Get("Authorization"))
}
