package database

import (
	"context"
	"testing"

	"github.com/example/scaffold/services/internal/data/model"
)

func TestOperationPolicyRepo(t *testing.T) {
	db := newPermissionTestDB(t)
	repo := NewPolicyRepo(&Data{DB: db})
	ctx := context.Background()

	if err := db.Create(&model.PermissionPO{Code: "report:export", Name: "Export", Source: "custom"}).Error; err != nil {
		t.Fatalf("seed permission: %v", err)
	}

	exists, err := repo.PermissionExists(ctx, "report:export")
	if err != nil || !exists {
		t.Fatalf("permission exists = %v, %v", exists, err)
	}
	exists, err = repo.PermissionExists(ctx, "missing:code")
	if err != nil || exists {
		t.Fatalf("missing permission exists = %v, %v", exists, err)
	}

	const operation = "/identity.v1.UserService/ListUsers"
	if err := repo.SetOperationPolicy(ctx, operation, "report:export"); err != nil {
		t.Fatalf("set policy: %v", err)
	}
	code, found, err := repo.GetOperationPolicy(ctx, operation)
	if err != nil || !found || code != "report:export" {
		t.Fatalf("get policy = %q, %v, %v", code, found, err)
	}

	if err := repo.SetOperationPolicy(ctx, operation, "user:list"); err != nil {
		t.Fatalf("update policy: %v", err)
	}
	policies, err := repo.ListOperationPolicies(ctx)
	if err != nil || policies[operation] != "user:list" {
		t.Fatalf("list policies = %+v, %v", policies, err)
	}

	if err := repo.DeleteOperationPolicy(ctx, operation); err != nil {
		t.Fatalf("delete policy: %v", err)
	}
	if _, found, err := repo.GetOperationPolicy(ctx, operation); err != nil || found {
		t.Fatalf("policy still found = %v, %v", found, err)
	}
}
