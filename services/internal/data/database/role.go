package database

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/data/model"
	"github.com/example/scaffold/services/internal/pkg/pagination"
)

type roleRepo struct {
	data *Data
}

// NewRoleRepo returns the database-backed role repository.
func NewRoleRepo(d *Data) biz.RoleRepo {
	return &roleRepo{data: d}
}

func (r *roleRepo) List(ctx context.Context, opts biz.ListOptions) ([]*biz.Role, int, error) {
	offset, _ := pagination.DecodeOffset(opts.PageToken)
	base := r.data.DB.WithContext(ctx).Model(&model.RolePO{})
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.RolePO
	if err := base.Order("name ASC").Offset(offset).Limit(opts.PageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	ids := make([]string, 0, len(rows))
	for i := range rows {
		ids = append(ids, rows[i].ID)
	}
	perms, err := loadPermissionsByRole(r.data.DB.WithContext(ctx), ids)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*biz.Role, 0, len(rows))
	for i := range rows {
		items = append(items, toBizRole(&rows[i], perms[rows[i].ID]))
	}
	return items, int(total), nil
}

func (r *roleRepo) GetByName(ctx context.Context, name string) (*biz.Role, error) {
	var po model.RolePO
	if err := r.data.DB.WithContext(ctx).Where("name = ?", name).First(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrRoleNotFound()
		}
		return nil, err
	}
	perms, err := loadPermissionsByRole(r.data.DB.WithContext(ctx), []string{po.ID})
	if err != nil {
		return nil, err
	}
	return toBizRole(&po, perms[po.ID]), nil
}

func (r *roleRepo) Create(ctx context.Context, role *biz.Role) (*biz.Role, error) {
	po := &model.RolePO{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		IsSystem:    role.IsSystem,
	}
	err := r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(po).Error; err != nil {
			return err
		}
		return replacePermissionsTx(tx, po.ID, role.Permissions)
	})
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, biz.ErrRoleAlreadyExists()
		}
		return nil, err
	}
	return toBizRole(po, role.Permissions), nil
}

func (r *roleRepo) Update(ctx context.Context, role *biz.Role) (*biz.Role, error) {
	var po model.RolePO
	if err := r.data.DB.WithContext(ctx).Where("id = ?", role.ID).First(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrRoleNotFound()
		}
		return nil, err
	}
	if role.DisplayName != "" {
		po.DisplayName = role.DisplayName
	}
	if role.Description != "" {
		po.Description = role.Description
	}
	err := r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&po).Error; err != nil {
			return err
		}
		if role.Permissions != nil {
			return replacePermissionsTx(tx, po.ID, role.Permissions)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	perms, err := loadPermissionsByRole(r.data.DB.WithContext(ctx), []string{po.ID})
	if err != nil {
		return nil, err
	}
	return toBizRole(&po, perms[po.ID]), nil
}

func (r *roleRepo) Delete(ctx context.Context, roleID string) error {
	var po model.RolePO
	if err := r.data.DB.WithContext(ctx).Where("id = ?", roleID).First(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return biz.ErrRoleNotFound()
		}
		return err
	}
	if po.IsSystem {
		return biz.ErrForbidden()
	}
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermissionPO{}).Error; err != nil {
			return err
		}
		if err := tx.Where("role_id = ?", roleID).Delete(&model.UserRolePO{}).Error; err != nil {
			return err
		}
		return tx.Delete(&po).Error
	})
}

func (r *roleRepo) ListPermissions(ctx context.Context) ([]*biz.Permission, error) {
	var rows []model.PermissionPO
	if err := r.data.DB.WithContext(ctx).Order("code ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]*biz.Permission, 0, len(rows))
	for i := range rows {
		items = append(items, &biz.Permission{
			Code:     rows[i].Code,
			Name:     rows[i].Name,
			Resource: rows[i].Resource,
			Action:   rows[i].Action,
		})
	}
	return items, nil
}

func (r *roleRepo) UpsertPermissions(ctx context.Context, permissions []*biz.Permission) error {
	if len(permissions) == 0 {
		return nil
	}
	rows := make([]model.PermissionPO, 0, len(permissions))
	for _, p := range permissions {
		rows = append(rows, model.PermissionPO{
			Code:     p.Code,
			Name:     p.Name,
			Resource: p.Resource,
			Action:   p.Action,
		})
	}
	return r.data.DB.WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&rows).Error
}

func (r *roleRepo) AssignRole(ctx context.Context, userID, roleName string) error {
	var role model.RolePO
	if err := r.data.DB.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return biz.ErrRoleNotFound()
		}
		return err
	}
	link := &model.UserRolePO{UserID: userID, RoleID: role.ID}
	return r.data.DB.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(link).Error
}

func replacePermissionsTx(tx *gorm.DB, roleID string, permissions []string) error {
	if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermissionPO{}).Error; err != nil {
		return err
	}
	seen := map[string]struct{}{}
	rows := make([]model.RolePermissionPO, 0, len(permissions))
	for _, code := range permissions {
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		if _, ok := seen[code]; ok {
			continue
		}
		seen[code] = struct{}{}
		rows = append(rows, model.RolePermissionPO{RoleID: roleID, PermissionCode: code})
	}
	if len(rows) == 0 {
		return nil
	}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

func loadPermissionsByRole(db *gorm.DB, roleIDs []string) (map[string][]string, error) {
	out := make(map[string][]string, len(roleIDs))
	if len(roleIDs) == 0 {
		return out, nil
	}
	type row struct {
		RoleID         string
		PermissionCode string
	}
	var rows []row
	err := db.Table("role_permissions").
		Select("role_id, permission_code").
		Where("role_id IN ?", roleIDs).
		Order("permission_code").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.RoleID] = append(out[r.RoleID], r.PermissionCode)
	}
	return out, nil
}

func toBizRole(po *model.RolePO, permissions []string) *biz.Role {
	return &biz.Role{
		ID:          po.ID,
		Name:        po.Name,
		DisplayName: po.DisplayName,
		Description: po.Description,
		IsSystem:    po.IsSystem,
		Permissions: permissions,
	}
}
