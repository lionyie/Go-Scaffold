package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v3/middleware"
	"github.com/go-kratos/kratos/v3/transport"

	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/pkg/auth"
	"github.com/example/scaffold/services/internal/server/authz"
)

// Authz enforces the operation-to-permission map. The manager resolves runtime
// database bindings over the compiled mapping, and the permission check reads
// the roles carried by the session token. Unknown operations fail closed so
// that a new endpoint cannot silently become public.
func Authz(manager *authz.Manager) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			operation := contextOperation(ctx)
			rule, ok, err := manager.Resolve(ctx, operation)
			if err != nil {
				return nil, biz.ErrInternal("authorization check failed")
			}
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
			if principal.IsAdmin || rule.Code == "" || principal.HasRole(rule.Code) {
				return handler(ctx, req)
			}
			return nil, biz.ErrForbidden()
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
