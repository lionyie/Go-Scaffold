package biz

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v3/errors"

	commonv1 "github.com/example/scaffold/services/api/common/v1"
)

type stubRoleRepo struct {
	created *Permission
}

func (s *stubRoleRepo) List(context.Context, ListOptions) ([]*Role, int, error) {
	return nil, 0, nil
}

func (s *stubRoleRepo) GetByName(context.Context, string) (*Role, error) { return nil, nil }

func (s *stubRoleRepo) Create(context.Context, *Role) (*Role, error) { return nil, nil }

func (s *stubRoleRepo) Update(context.Context, *Role) (*Role, error) { return nil, nil }

func (s *stubRoleRepo) Delete(context.Context, string) error { return nil }

func (s *stubRoleRepo) ListPermissions(context.Context) ([]*Permission, error) { return nil, nil }

func (s *stubRoleRepo) CreatePermission(_ context.Context, permission *Permission) (*Permission, error) {
	s.created = permission
	return permission, nil
}

func (s *stubRoleRepo) UpdatePermission(_ context.Context, permission *Permission) (*Permission, error) {
	return permission, nil
}

func (s *stubRoleRepo) DeletePermission(context.Context, string) error { return nil }

func (s *stubRoleRepo) UpsertPermissions(context.Context, []*Permission) error { return nil }

func (s *stubRoleRepo) AssignRole(context.Context, string, string) error { return nil }

func TestCreatePermissionNormalizesAndMarksCustom(t *testing.T) {
	repo := &stubRoleRepo{}
	uc := NewRoleUsecase(repo)

	permission, err := uc.CreatePermission(context.Background(), &Permission{
		Code: "  report:export  ",
		Name: "  Export reports  ",
	})
	if err != nil {
		t.Fatalf("create permission: %v", err)
	}
	if permission.Code != "report:export" {
		t.Fatalf("code = %q", permission.Code)
	}
	if permission.Name != "Export reports" {
		t.Fatalf("name = %q", permission.Name)
	}
	if permission.Resource != "report" || permission.Action != "export" {
		t.Fatalf("resource/action = %q/%q", permission.Resource, permission.Action)
	}
	if permission.Source != PermissionSourceCustom {
		t.Fatalf("source = %q, want %q", permission.Source, PermissionSourceCustom)
	}
	if repo.created == nil {
		t.Fatal("repository was not called")
	}
}

func TestCreatePermissionRejectsInvalidInput(t *testing.T) {
	uc := NewRoleUsecase(&stubRoleRepo{})

	cases := []struct {
		name       string
		permission *Permission
	}{
		{"nil", nil},
		{"empty code", &Permission{Code: "", Name: "Name"}},
		{"no separator", &Permission{Code: "report", Name: "Name"}},
		{"uppercase", &Permission{Code: "Report:Export", Name: "Name"}},
		{"empty name", &Permission{Code: "report:export", Name: "  "}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := uc.CreatePermission(context.Background(), tc.permission)
			if reason := errors.FromError(err).Reason; reason != commonv1.ErrorReason_INVALID_ARGUMENT.String() {
				t.Fatalf("reason = %q, want INVALID_ARGUMENT", reason)
			}
		})
	}
}

func TestDeletePermissionRequiresCode(t *testing.T) {
	uc := NewRoleUsecase(&stubRoleRepo{})
	err := uc.DeletePermission(context.Background(), "   ")
	if reason := errors.FromError(err).Reason; reason != commonv1.ErrorReason_INVALID_ARGUMENT.String() {
		t.Fatalf("reason = %q, want INVALID_ARGUMENT", reason)
	}
}
