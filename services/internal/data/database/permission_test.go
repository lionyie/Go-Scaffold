package database

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v3/errors"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"

	commonv1 "github.com/example/scaffold/services/api/common/v1"
	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/data/model"
)

func newPermissionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.PermissionPO{},
		&model.OperationPolicyPO{},
		&model.RolePO{},
		&model.RolePermissionPO{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func TestPermissionCRUD(t *testing.T) {
	db := newPermissionTestDB(t)
	repo := NewRoleRepo(&Data{DB: db})
	ctx := context.Background()

	created, err := repo.CreatePermission(ctx, &biz.Permission{Code: "report:export", Name: "Export reports"})
	if err != nil {
		t.Fatalf("create permission: %v", err)
	}
	if created.Source != biz.PermissionSourceCustom {
		t.Fatalf("source = %q, want %q", created.Source, biz.PermissionSourceCustom)
	}

	_, err = repo.CreatePermission(ctx, &biz.Permission{Code: "report:export", Name: "Duplicate"})
	if reason := errors.FromError(err).Reason; reason != commonv1.ErrorReason_PERMISSION_ALREADY_EXISTS.String() {
		t.Fatalf("duplicate reason = %q, want PERMISSION_ALREADY_EXISTS", reason)
	}

	updated, err := repo.UpdatePermission(ctx, &biz.Permission{
		Code:     "report:export",
		Name:     "Export CSV",
		Resource: "report",
		Action:   "export",
	})
	if err != nil {
		t.Fatalf("update permission: %v", err)
	}
	if updated.Name != "Export CSV" || updated.Source != biz.PermissionSourceCustom {
		t.Fatalf("unexpected updated permission: %+v", updated)
	}

	items, err := repo.ListPermissions(ctx)
	if err != nil {
		t.Fatalf("list permissions: %v", err)
	}
	if len(items) != 1 || items[0].Source != biz.PermissionSourceCustom {
		t.Fatalf("unexpected permissions: %+v", items)
	}

	role := &model.RolePO{ID: "role-1", Name: "role-1"}
	if err := db.Create(role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	if err := db.Create(&model.RolePermissionPO{RoleID: role.ID, PermissionCode: "report:export"}).Error; err != nil {
		t.Fatalf("link role permission: %v", err)
	}
	if err := db.Create(&model.OperationPolicyPO{
		Operation:      "/identity.v1.UserService/ListUsers",
		PermissionCode: "report:export",
	}).Error; err != nil {
		t.Fatalf("bind operation policy: %v", err)
	}

	if err := repo.DeletePermission(ctx, "report:export"); err != nil {
		t.Fatalf("delete permission: %v", err)
	}
	var links int64
	if err := db.Model(&model.RolePermissionPO{}).
		Where("permission_code = ?", "report:export").
		Count(&links).Error; err != nil {
		t.Fatalf("count role permissions: %v", err)
	}
	if links != 0 {
		t.Fatalf("role_permissions still has %d rows", links)
	}
	var policies int64
	if err := db.Model(&model.OperationPolicyPO{}).
		Where("permission_code = ?", "report:export").
		Count(&policies).Error; err != nil {
		t.Fatalf("count operation policies: %v", err)
	}
	if policies != 0 {
		t.Fatalf("operation_permissions still has %d rows", policies)
	}
}

func TestPermissionDeleteRejectsBuiltin(t *testing.T) {
	db := newPermissionTestDB(t)
	repo := NewRoleRepo(&Data{DB: db})
	ctx := context.Background()

	if err := db.Create(&model.PermissionPO{
		Code:   "user:list",
		Name:   "List users",
		Source: biz.PermissionSourceBuiltin,
	}).Error; err != nil {
		t.Fatalf("seed builtin permission: %v", err)
	}

	err := repo.DeletePermission(ctx, "user:list")
	if reason := errors.FromError(err).Reason; reason != commonv1.ErrorReason_PERMISSION_READ_ONLY.String() {
		t.Fatalf("reason = %q, want PERMISSION_READ_ONLY", reason)
	}
}

func TestSeedPermissionsPreservesCustomAndPlugin(t *testing.T) {
	db := newPermissionTestDB(t)

	rows := []model.PermissionPO{
		{Code: "custom:one", Name: "Custom", Source: biz.PermissionSourceCustom},
		{Code: "plugin:one", Name: "Plugin", Source: biz.PermissionSourcePlugin},
		{Code: "stale:builtin", Name: "Stale", Source: biz.PermissionSourceBuiltin},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed permissions: %v", err)
	}
	if err := seedPermissions(db); err != nil {
		t.Fatalf("seedPermissions: %v", err)
	}

	exists := func(code string) bool {
		var count int64
		if err := db.Model(&model.PermissionPO{}).Where("code = ?", code).Count(&count).Error; err != nil {
			t.Fatalf("count %s: %v", code, err)
		}
		return count > 0
	}
	if !exists("custom:one") || !exists("plugin:one") {
		t.Fatal("seeder pruned custom or plugin permissions")
	}
	if exists("stale:builtin") {
		t.Fatal("seeder kept a stale builtin permission")
	}
	if !exists("user:list") {
		t.Fatal("seeder did not insert the builtin catalog")
	}
}
