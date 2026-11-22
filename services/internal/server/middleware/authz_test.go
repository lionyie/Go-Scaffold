package middleware

import (
	"context"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"
	"github.com/go-kratos/kratos/v3/transport"

	identityv1 "github.com/example/scaffold/services/api/identity/v1"
	systemv1 "github.com/example/scaffold/services/api/system/v1"
	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
	"github.com/example/scaffold/services/internal/server/authz"
)

type fakePolicyRepo struct{}

func (fakePolicyRepo) ListOperationPolicies(context.Context) (map[string]string, error) {
	return map[string]string{}, nil
}

func (fakePolicyRepo) GetOperationPolicy(context.Context, string) (string, bool, error) {
	return "", false, nil
}

func (fakePolicyRepo) SetOperationPolicy(context.Context, string, string) error { return nil }

func (fakePolicyRepo) DeleteOperationPolicy(context.Context, string) error { return nil }

func (fakePolicyRepo) PermissionExists(context.Context, string) (bool, error) { return true, nil }

type fakeChecker struct {
	allowed map[string]bool
}

func (f fakeChecker) HasGlobal(_ context.Context, userID, code string) (bool, error) {
	return f.allowed[userID+"|"+code], nil
}

func (f fakeChecker) Permissions(_ context.Context, userID string) ([]string, error) {
	out := []string{}
	for key, allowed := range f.allowed {
		if allowed && len(key) > len(userID) && key[:len(userID)] == userID {
			out = append(out, key[len(userID)+1:])
		}
	}
	return out, nil
}

func (f fakeChecker) Invalidate(string) {}

type fakeHeader map[string][]string

func (h fakeHeader) Get(key string) string {
	if values := h[key]; len(values) > 0 {
		return values[0]
	}
	return ""
}
func (h fakeHeader) Set(key, value string) { h[key] = []string{value} }
func (h fakeHeader) Add(key, value string) { h[key] = append(h[key], value) }
func (h fakeHeader) Keys() []string {
	out := make([]string, 0, len(h))
	for key := range h {
		out = append(out, key)
	}
	return out
}
func (h fakeHeader) Values(key string) []string { return h[key] }

type fakeTransport struct {
	operation string
}

func (f *fakeTransport) Kind() transport.Kind            { return transport.KindHTTP }
func (f *fakeTransport) Endpoint() string                { return "" }
func (f *fakeTransport) Operation() string               { return f.operation }
func (f *fakeTransport) RequestHeader() transport.Header { return fakeHeader{} }
func (f *fakeTransport) ReplyHeader() transport.Header   { return fakeHeader{} }

func runAuthz(t *testing.T, _ fakeChecker, operation string, principal *authpkg.Principal) (any, error) {
	t.Helper()
	ctx := transport.NewServerContext(context.Background(), &fakeTransport{operation: operation})
	if principal != nil {
		ctx = authpkg.WithPrincipal(ctx, principal)
	}
	handler := Authz(authz.NewManager(fakePolicyRepo{}))(func(context.Context, any) (any, error) { return "ok", nil })
	return handler(ctx, struct{}{})
}

func assertCode(t *testing.T, err error, code int32, reason string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error %d/%s, got nil", code, reason)
	}
	parsed := kerrors.FromError(err)
	if parsed.Code != code || parsed.Reason != reason {
		t.Fatalf("got %d/%s, want %d/%s", parsed.Code, parsed.Reason, code, reason)
	}
}

func TestAuthzPublic(t *testing.T) {
	checker := fakeChecker{allowed: map[string]bool{}}
	if _, err := runAuthz(t, checker, systemv1.OperationSystemServiceGetHealth, nil); err != nil {
		t.Fatalf("public operation failed: %v", err)
	}
}

func TestAuthzUnknownOperation(t *testing.T) {
	checker := fakeChecker{allowed: map[string]bool{}}
	_, err := runAuthz(t, checker, "/unknown.Service/Operation", nil)
	assertCode(t, err, 403, "FORBIDDEN")
}

func TestAuthzRequiresPrincipal(t *testing.T) {
	checker := fakeChecker{allowed: map[string]bool{}}
	_, err := runAuthz(t, checker, identityv1.OperationUserServiceListUsers, nil)
	assertCode(t, err, 401, "UNAUTHORIZED")
}

func TestAuthzDeniesWithoutPermission(t *testing.T) {
	checker := fakeChecker{allowed: map[string]bool{}}
	_, err := runAuthz(t, checker, identityv1.OperationUserServiceListUsers, &authpkg.Principal{UserID: "u1", Username: "viewer"})
	assertCode(t, err, 403, "FORBIDDEN")
}

func TestAuthzAllowsWithPermission(t *testing.T) {
	checker := fakeChecker{allowed: map[string]bool{"u1|user:list": true}}
	principal := &authpkg.Principal{UserID: "u1", Username: "viewer", Roles: []string{"user:list"}}
	if _, err := runAuthz(t, checker, identityv1.OperationUserServiceListUsers, principal); err != nil {
		t.Fatalf("permission should allow: %v", err)
	}
}

func TestAuthzAdminBypass(t *testing.T) {
	checker := fakeChecker{allowed: map[string]bool{}}
	principal := &authpkg.Principal{UserID: "admin", Username: "admin", IsAdmin: true}
	if _, err := runAuthz(t, checker, identityv1.OperationUserServiceListUsers, principal); err != nil {
		t.Fatalf("admin should bypass: %v", err)
	}
}

func TestAuthzMustChangePassword(t *testing.T) {
	checker := fakeChecker{allowed: map[string]bool{"u1|user:list": true}}
	principal := &authpkg.Principal{UserID: "u1", Username: "admin", MustChangePassword: true, Roles: []string{"user:list"}}
	_, err := runAuthz(t, checker, identityv1.OperationUserServiceListUsers, principal)
	assertCode(t, err, 403, "MUST_CHANGE_PASSWORD")

	if _, err := runAuthz(t, checker, identityv1.OperationAuthServiceChangePassword, principal); err != nil {
		t.Fatalf("change-password must be allowed: %v", err)
	}
	if _, err := runAuthz(t, checker, identityv1.OperationAuthServiceGetCurrentUser, principal); err != nil {
		t.Fatalf("get-current-user must be allowed: %v", err)
	}
	if _, err := runAuthz(t, checker, identityv1.OperationAuthServiceLogout, principal); err != nil {
		t.Fatalf("logout must be allowed: %v", err)
	}
}
