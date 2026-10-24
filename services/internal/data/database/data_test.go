package database

import (
	"log/slog"
	"testing"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"

	"github.com/example/scaffold/services/internal/data/model"
)

func TestOAuthAccountTable(t *testing.T) {
	db, err := gorm.Open(gormlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.UserPO{}, &model.OAuthAccountPO{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	if !db.Migrator().HasTable(&model.OAuthAccountPO{}) {
		t.Fatal("oauth_accounts table was not created")
	}
	if !db.Migrator().HasTable(&model.UserPO{}) {
		t.Fatal("users table was not created")
	}
}

func TestRunMigrationsReadsDirectory(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(&testWriter{t: t}, nil))
	if err := runMigrations(t.TempDir(), logger); err != nil {
		t.Fatalf("empty migrations directory: %v", err)
	}
}

type testWriter struct{ t *testing.T }

func (w *testWriter) Write(p []byte) (int, error) {
	w.t.Log(string(p))
	return len(p), nil
}
