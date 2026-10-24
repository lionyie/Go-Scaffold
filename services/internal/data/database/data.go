package database

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/example/scaffold/services/internal/conf"
	"github.com/example/scaffold/services/internal/data/model"
)

// Data holds the long-lived database handle shared by repos.
type Data struct {
	DB *gorm.DB
}

// NewData opens the configured database, runs development migrations, and
// returns a cleanup function.
func NewData(bc *conf.Bootstrap, log *slog.Logger) (*Data, func(), error) {
	dc := bc.GetData().GetDatabase()
	driver := dc.GetDriver()
	source := dc.GetSource()

	if driver == "sqlite" || driver == "sqlite3" {
		if err := ensureSQLiteDir(source); err != nil {
			return nil, nil, err
		}
	}

	var dialector gorm.Dialector
	switch driver {
	case "sqlite", "sqlite3":
		dialector = gormlite.Open(source)
	case "mysql":
		dialector = mysql.Open(source)
	default:
		return nil, nil, fmt.Errorf("unsupported database driver %q", driver)
	}

	gormLogger := logger.Default.LogMode(logger.Warn)
	if dc.GetDebug() {
		gormLogger = logger.Default.LogMode(logger.Info)
	}
	db, err := gorm.Open(dialector, &gorm.Config{
		Logger:         gormLogger,
		TranslateError: true,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get sql database: %w", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if driver == "sqlite" || driver == "sqlite3" {
		// SQLite serializes writes; a single writer avoids busy errors.
		sqlDB.SetMaxOpenConns(1)
	}

	if dc.GetAutoMigrate() {
		if err := runMigrations("internal/data/migrations", log); err != nil {
			return nil, nil, fmt.Errorf("run migrations: %w", err)
		}
		if err := db.AutoMigrate(
			&model.UserPO{},
			&model.RolePO{},
			&model.PermissionPO{},
			&model.RolePermissionPO{},
			&model.UserRolePO{},
			&model.RefreshTokenPO{},
			&model.OAuthAccountPO{},
		); err != nil {
			return nil, nil, fmt.Errorf("auto migrate: %w", err)
		}
	}

	cleanup := func() {
		log.Info("closing the data resources")
		if err := sqlDB.Close(); err != nil && log != nil {
			log.Error("failed closing the database", "err", err)
		}
	}
	data := &Data{DB: db}
	if dc.GetSeed() {
		seeder := &Seeder{data: data, log: log}
		if err := seeder.Run(context.Background()); err != nil {
			return nil, nil, fmt.Errorf("seed data: %w", err)
		}
	}
	return data, cleanup, nil
}

// runMigrations reads the migration directory before AutoMigrate. The scaffold
// currently ships an empty directory; this hook keeps the upgrade path explicit.
func runMigrations(dir string, log *slog.Logger) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			count++
		}
	}
	if count > 0 && log != nil {
		log.Warn("SQL migration files are present but the migrator is not enabled",
			"dir", dir, "files", count)
	}
	return nil
}

func ensureSQLiteDir(source string) error {
	path := source
	if len(path) > 5 && path[:5] == "file:" {
		path = path[5:]
	}
	for i := 0; i < len(path); i++ {
		if path[i] == '?' {
			path = path[:i]
			break
		}
	}
	if path == "" || path == ":memory:" {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
