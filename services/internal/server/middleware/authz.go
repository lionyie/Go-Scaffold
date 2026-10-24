package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"

	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/pkg/auth"
	"github.com/example/scaffold/services/internal/pkg/rbac"
	"github.com/example/scaffold/services/internal/server/authz"
)

// Authz enforces the operation-to-permission map. Unknown operations fail
// closed so that a new endpoint cannot silently become public.
func Authz(checker rbac.Checker) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			operation := contextOperation(ctx)
			rule, ok := authz.Requirements[operation]
			if !ok {
				return nil, biz.ErrForbidden()
			}
			if rule.Public {
				return handler(ctx, req)
			}

			principal, ok := auth.PrincipalFromContext(ctx)
			if !ok {
				return nil, biz.ErrUnauthorized()
			}
			// Authenticated-only operations such as GetCurrentUser and Logout
			// stay available so the client can render and complete the forced
			// password rotation flow. Permissioned operations are blocked.
			if principal.MustChangePassword && rule.Code != "" && operation != identityChangePasswordOperation {
				return nil, biz.ErrMustChangePassword()
			}
			if principal.IsAdmin || rule.Code == "" {
				return handler(ctx, req)
			}

			allowed, err := checker.HasGlobal(ctx, principal.UserID, rule.Code)
			if err != nil {
				return nil, biz.ErrInternal("authorization check failed")
			}
			if !allowed {
				return nil, biz.ErrForbidden()
			}
			return handler(ctx, req)
		}
	}
}

// identityChangePasswordOperation is the only operation allowed while a user
// must rotate the seeded password.
const identityChangePasswordOperation = "/identity.v1.AuthService/ChangePassword"

func contextOperation(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		return tr.Operation()
	}
	return ""
}
