package authz

import (
	"context"
	"testing"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	identityv1 "github.com/example/scaffold/services/api/identity/v1"
	systemv1 "github.com/example/scaffold/services/api/system/v1"
)

type fakePolicyRepo struct {
	policies    map[string]string
	permissions map[string]bool
}

func newFakePolicyRepo() *fakePolicyRepo {
	return &fakePolicyRepo{policies: map[string]string{}, permissions: map[string]bool{}}
}

func (f *fakePolicyRepo) ListOperationPolicies(context.Context) (map[string]string, error) {
	out := make(map[string]string, len(f.policies))
	for k, v := range f.policies {
		out[k] = v
	}
	return out, nil
}

func (f *fakePolicyRepo) GetOperationPolicy(_ context.Context, operation string) (string, bool, error) {
	code, ok := f.policies[operation]
	return code, ok, nil
}

func (f *fakePolicyRepo) SetOperationPolicy(_ context.Context, operation, code string) error {
	f.policies[operation] = code
	return nil
}

func (f *fakePolicyRepo) DeleteOperationPolicy(_ context.Context, operation string) error {
	delete(f.policies, operation)
	return nil
}

func (f *fakePolicyRepo) PermissionExists(_ context.Context, code string) (bool, error) {
	return f.permissions[code], nil
}

func reasonOf(err error) string {
	return kerrors.FromError(err).Reason
}

func TestResolvePrefersDatabaseBinding(t *testing.T) {
	repo := newFakePolicyRepo()
	manager := NewManager(repo)
	ctx := context.Background()

	rule, ok, err := manager.Resolve(ctx, identityv1.OperationUserServiceListUsers)
	if err != nil || !ok || rule.Code != "user:list" {
		t.Fatalf("static resolve = %+v, %v, %v", rule, ok, err)
	}

	repo.policies[identityv1.OperationUserServiceListUsers] = "user:delete"
	manager.Invalidate(identityv1.OperationUserServiceListUsers)

	rule, ok, err = manager.Resolve(ctx, identityv1.OperationUserServiceListUsers)
	if err != nil || !ok || rule.Code != "user:delete" {
		t.Fatalf("database resolve = %+v, %v, %v", rule, ok, err)
	}
}

func TestResolvePublicAndAuthenticated(t *testing.T) {
	manager := NewManager(newFakePolicyRepo())
	ctx := context.Background()

	rule, ok, err := manager.Resolve(ctx, systemv1.OperationSystemServiceGetHealth)
	if err != nil || !ok || !rule.Public {
		t.Fatalf("public resolve = %+v, %v, %v", rule, ok, err)
	}

	rule, ok, err = manager.Resolve(ctx, identityv1.OperationAuthServiceLogout)
	if err != nil || !ok || rule.Code != "" || rule.Public {
		t.Fatalf("authenticated resolve = %+v, %v, %v", rule, ok, err)
	}
}

func TestResolveUnknownOperationFailsClosed(t *testing.T) {
	manager := NewManager(newFakePolicyRepo())
	_, ok, err := manager.Resolve(context.Background(), "/unknown.Service/Operation")
	if err != nil || ok {
		t.Fatalf("unknown operation = ok %v, err %v", ok, err)
	}
}

func TestSetRejectsLockedOperations(t *testing.T) {
	manager := NewManager(newFakePolicyRepo())
	ctx := context.Background()

	if _, err := manager.Set(ctx, systemv1.OperationSystemServiceGetHealth, "user:list"); reasonOf(err) != "FORBIDDEN" {
		t.Fatalf("public reason = %q", reasonOf(err))
	}
	if _, err := manager.Set(ctx, identityv1.OperationAuthServiceLogout, "user:list"); reasonOf(err) != "FORBIDDEN" {
		t.Fatalf("authenticated reason = %q", reasonOf(err))
	}
	if _, err := manager.Set(ctx, "/unknown.Service/Operation", "user:list"); reasonOf(err) != "INVALID_ARGUMENT" {
		t.Fatalf("unknown reason = %q", reasonOf(err))
	}
}

func TestSetRejectsUnknownPermission(t *testing.T) {
	manager := NewManager(newFakePolicyRepo())
	_, err := manager.Set(context.Background(), identityv1.OperationUserServiceListUsers, "missing:code")
	if reasonOf(err) != "PERMISSION_NOT_FOUND" {
		t.Fatalf("reason = %q", reasonOf(err))
	}
}

func TestSetAndReset(t *testing.T) {
	repo := newFakePolicyRepo()
	repo.permissions["report:export"] = true
	manager := NewManager(repo)
	ctx := context.Background()

	view, err := manager.Set(ctx, identityv1.OperationUserServiceListUsers, "report:export")
	if err != nil {
		t.Fatalf("set: %v", err)
	}
	if view.Permission != "report:export" || view.Source != "database" || view.StaticPermission != "user:list" {
		t.Fatalf("view after set = %+v", view)
	}

	rule, ok, err := manager.Resolve(ctx, identityv1.OperationUserServiceListUsers)
	if err != nil || !ok || rule.Code != "report:export" {
		t.Fatalf("resolve after set = %+v, %v, %v", rule, ok, err)
	}

	view, err = manager.Reset(ctx, identityv1.OperationUserServiceListUsers)
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	if view.Permission != "user:list" || view.Source != "static" {
		t.Fatalf("view after reset = %+v", view)
	}

	rule, ok, err = manager.Resolve(ctx, identityv1.OperationUserServiceListUsers)
	if err != nil || !ok || rule.Code != "user:list" {
		t.Fatalf("resolve after reset = %+v, %v, %v", rule, ok, err)
	}
}

func TestListReportsSources(t *testing.T) {
	repo := newFakePolicyRepo()
	repo.policies[identityv1.OperationUserServiceListUsers] = "user:delete"
	manager := NewManager(repo)

	views, err := manager.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	byOperation := make(map[string]OperationView, len(views))
	for _, view := range views {
		byOperation[view.Operation] = view
	}

	if got := byOperation[systemv1.OperationSystemServiceGetHealth]; got.Source != "public" || !got.Locked {
		t.Fatalf("public view = %+v", got)
	}
	if got := byOperation[identityv1.OperationAuthServiceLogout]; got.Source != "authenticated" || !got.Locked {
		t.Fatalf("authenticated view = %+v", got)
	}
	if got := byOperation[identityv1.OperationUserServiceListUsers]; got.Source != "database" || got.Permission != "user:delete" {
		t.Fatalf("database view = %+v", got)
	}
}
