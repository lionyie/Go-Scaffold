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

type userRepo struct {
	data *Data
}

// NewUserRepo returns the database-backed user repository.
func NewUserRepo(d *Data) biz.UserRepo {
	return &userRepo{data: d}
}

func (r *userRepo) Create(ctx context.Context, user *biz.User) (*biz.User, error) {
	po := &model.UserPO{
		ID:                 user.ID,
		Username:           user.Username,
		DisplayName:        user.DisplayName,
		Email:              user.Email,
		PasswordHash:       user.PasswordHash,
		IsAdmin:            user.IsAdmin,
		MustChangePassword: user.MustChangePassword,
		Status:             user.Status,
	}
	err := r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(po).Error; err != nil {
			return err
		}
		return assignRolesTx(tx, user.ID, user.Roles)
	})
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, biz.ErrUserAlreadyExists()
		}
		if errors.Is(err, errRoleMissing) {
			return nil, biz.ErrRoleNotFound()
		}
		return nil, err
	}
	return toBizUser(po, user.Roles), nil
}

func (r *userRepo) GetByID(ctx context.Context, id string) (*biz.User, error) {
	var po model.UserPO
	if err := r.data.DB.WithContext(ctx).Where("id = ?", id).First(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound()
		}
		return nil, err
	}
	roles, err := loadRoleNames(r.data.DB.WithContext(ctx), po.ID)
	if err != nil {
		return nil, err
	}
	return toBizUser(&po, roles), nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*biz.User, error) {
	var po model.UserPO
	if err := r.data.DB.WithContext(ctx).Where("username = ?", username).First(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound()
		}
		return nil, err
	}
	roles, err := loadRoleNames(r.data.DB.WithContext(ctx), po.ID)
	if err != nil {
		return nil, err
	}
	return toBizUser(&po, roles), nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*biz.User, error) {
	var po model.UserPO
	if err := r.data.DB.WithContext(ctx).Where("email = ?", email).First(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound()
		}
		return nil, err
	}
	roles, err := loadRoleNames(r.data.DB.WithContext(ctx), po.ID)
	if err != nil {
		return nil, err
	}
	return toBizUser(&po, roles), nil
}

func (r *userRepo) List(ctx context.Context, opts biz.ListOptions) ([]*biz.User, int, error) {
	offset, _ := pagination.DecodeOffset(opts.PageToken)
	base := r.data.DB.WithContext(ctx).Model(&model.UserPO{})
	if keyword := strings.TrimSpace(opts.Keyword); keyword != "" {
		like := "%" + keyword + "%"
		base = base.Where("username LIKE ? OR display_name LIKE ?", like, like)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []model.UserPO
	if err := base.Order("created_at DESC").Offset(offset).Limit(opts.PageSize).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*biz.User, 0, len(rows))
	for i := range rows {
		roles, err := loadRoleNames(r.data.DB.WithContext(ctx), rows[i].ID)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, toBizUser(&rows[i], roles))
	}
	return items, int(total), nil
}

func (r *userRepo) Update(ctx context.Context, user *biz.User) (*biz.User, error) {
	var po model.UserPO
	if err := r.data.DB.WithContext(ctx).Where("id = ?", user.ID).First(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrUserNotFound()
		}
		return nil, err
	}
	if user.DisplayName != "" {
		po.DisplayName = user.DisplayName
	}
	if user.Email != "" {
		po.Email = user.Email
	}
	if user.Status != "" {
		po.Status = user.Status
	}
	if user.PasswordHash != "" {
		po.PasswordHash = user.PasswordHash
	}
	po.IsAdmin = user.IsAdmin
	po.MustChangePassword = user.MustChangePassword
	err := r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&po).Error; err != nil {
			return err
		}
		if len(user.Roles) > 0 {
			if err := tx.Where("user_id = ?", user.ID).Delete(&model.UserRolePO{}).Error; err != nil {
				return err
			}
			return assignRolesTx(tx, user.ID, user.Roles)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, errRoleMissing) {
			return nil, biz.ErrRoleNotFound()
		}
		return nil, err
	}
	roles, err := loadRoleNames(r.data.DB.WithContext(ctx), po.ID)
	if err != nil {
		return nil, err
	}
	return toBizUser(&po, roles), nil
}

func (r *userRepo) Delete(ctx context.Context, id string) error {
	return r.data.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("id = ?", id).Delete(&model.UserPO{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return biz.ErrUserNotFound()
		}
		return tx.Where("user_id = ?", id).Delete(&model.UserRolePO{}).Error
	})
}

var errRoleMissing = errors.New("role missing")

func assignRolesTx(tx *gorm.DB, userID string, roleNames []string) error {
	for _, name := range roleNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		var role model.RolePO
		if err := tx.Where("name = ?", name).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errRoleMissing
			}
			return err
		}
		link := &model.UserRolePO{UserID: userID, RoleID: role.ID}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(link).Error; err != nil {
			return err
		}
	}
	return nil
}

func loadRoleNames(db *gorm.DB, userID string) ([]string, error) {
	var names []string
	err := db.Table("roles AS r").
		Select("DISTINCT r.name").
		Joins("JOIN user_roles AS ur ON ur.role_id = r.id").
		Where("ur.user_id = ?", userID).
		Order("r.name").
		Scan(&names).Error
	return names, err
}

func toBizUser(po *model.UserPO, roles []string) *biz.User {
	return &biz.User{
		ID:                 po.ID,
		Username:           po.Username,
		DisplayName:        po.DisplayName,
		Email:              po.Email,
		PasswordHash:       po.PasswordHash,
		IsAdmin:            po.IsAdmin,
		MustChangePassword: po.MustChangePassword,
		Status:             po.Status,
		Roles:              roles,
		CreatedAt:          po.CreatedAt.UnixMilli(),
	}
}
