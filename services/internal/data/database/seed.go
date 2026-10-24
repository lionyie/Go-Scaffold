package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/example/scaffold/services/internal/data/model"
	authpkg "github.com/example/scaffold/services/internal/pkg/auth"
	"github.com/example/scaffold/services/internal/pkg/id"
	"github.com/example/scaffold/services/internal/pkg/rbac"
)

// Seeder inserts idempotent development data.
type Seeder struct {
	data *Data
	log  *slog.Logger
}

// Run executes the seed routine.
func (s *Seeder) Run(ctx context.Context) error {
	return s.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := seedPermissions(tx); err != nil {
			return err
		}
		if err := seedRoles(tx); err != nil {
			return err
		}
		if err := s.seedAdmin(tx); err != nil {
			return err
		}
		return s.verifyProductionAdmin(tx)
	})
}

func (s *Seeder) verifyProductionAdmin(tx *gorm.DB) error {
	if os.Getenv("APP_ENV") != "production" {
		return nil
	}
	var count int64
	if err := tx.Model(&model.UserPO{}).
		Where("username = ? AND must_change_password = ?", "admin", true).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("production refuses to start with the seeded administrator password; set SCAFFOLD_SEED_ADMIN_PASSWORD before the first start or change the password")
	}
	return nil
}

func seedPermissions(tx *gorm.DB) error {
	catalog := rbac.Catalog()
	rows := make([]model.PermissionPO, 0, len(catalog))
	codes := make([]string, 0, len(catalog))
	for _, p := range catalog {
		codes = append(codes, p.Code)
		rows = append(rows, model.PermissionPO{
			Code:     p.Code,
			Name:     p.Name,
			Resource: p.Resource,
			Action:   p.Action,
		})
	}
	if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&rows).Error; err != nil {
		return err
	}
	return tx.Where("code NOT IN ?", codes).Delete(&model.PermissionPO{}).Error
}

func seedRoles(tx *gorm.DB) error {
	all := make([]string, 0, len(rbac.Catalog()))
	for _, p := range rbac.Catalog() {
		all = append(all, p.Code)
	}
	readOnly := []string{
		rbac.PermUserList,
		rbac.PermRoleList,
		rbac.PermRolePermissionList,
		rbac.PermPluginList,
		rbac.PermPluginRead,
	}

	seeds := []struct {
		name        string
		displayName string
		permissions []string
	}{
		{"admin", "Administrator", all},
		{"viewer", "Read-only administrator", readOnly},
	}

	for _, seed := range seeds {
		var role model.RolePO
		err := tx.Where("name = ?", seed.name).First(&role).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			role = model.RolePO{
				ID:          id.New(),
				Name:        seed.name,
				DisplayName: seed.displayName,
				IsSystem:    true,
			}
			if err := tx.Create(&role).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
		if err := replacePermissionsTx(tx, role.ID, seed.permissions); err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) seedAdmin(tx *gorm.DB) error {
	seedPassword := os.Getenv("SCAFFOLD_SEED_ADMIN_PASSWORD")
	password := seedPassword
	mustChange := false
	if password == "" {
		password = "admin123"
		mustChange = true
	}

	var user model.UserPO
	err := tx.Where("username = ?", "admin").First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hash, hashErr := authpkg.HashPassword(password)
		if hashErr != nil {
			return hashErr
		}
		user = model.UserPO{
			ID:                 id.New(),
			Username:           "admin",
			DisplayName:        "Administrator",
			PasswordHash:       hash,
			IsAdmin:            true,
			MustChangePassword: mustChange,
			Status:             "active",
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		if mustChange {
			s.log.Warn("seeded default administrator account; change the password on first login",
				"username", "admin", "password", "admin123")
		}
	} else if err != nil {
		return err
	} else if seedPassword != "" && user.MustChangePassword {
		hash, hashErr := authpkg.HashPassword(seedPassword)
		if hashErr != nil {
			return hashErr
		}
		user.PasswordHash = hash
		user.MustChangePassword = false
		if err := tx.Save(&user).Error; err != nil {
			return err
		}
	}

	var role model.RolePO
	if err := tx.Where("name = ?", "admin").First(&role).Error; err != nil {
		return err
	}
	link := &model.UserRolePO{UserID: user.ID, RoleID: role.ID}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(link).Error
}
