package auth

import (
	"context"
	"strings"
)

// TokenTypeAccess and TokenTypeRefresh distinguish the two issued tokens.
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

// Principal is the authenticated caller extracted from an access token.
type Principal struct {
	UserID             string
	Username           string
	SessionID          string
	IsAdmin            bool
	MustChangePassword bool
}

type principalKey struct{}

// WithPrincipal stores the principal in the context.
func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

// PrincipalFromContext returns the principal stored in the context.
func PrincipalFromContext(ctx context.Context) (*Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(*Principal)
	return p, ok
}

// Bearer extracts the token from an Authorization header value.
func Bearer(header string) string {
	const prefix = "Bearer "
	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return strings.TrimSpace(header[len(prefix):])
	}
	return ""
}
