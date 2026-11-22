package database

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/data/model"
)

type policyRepo struct {
	data *Data
}

// NewPolicyRepo returns the runtime operation-policy repository.
func NewPolicyRepo(d *Data) biz.PolicyRepo {
	return &policyRepo{data: d}
}

func (r *policyRepo) ListOperationPolicies(ctx context.Context) (map[string]string, error) {
	var rows []model.OperationPolicyPO
	if err := r.data.DB.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for i := range rows {
		out[rows[i].Operation] = rows[i].PermissionCode
	}
	return out, nil
}

func (r *policyRepo) GetOperationPolicy(ctx context.Context, operation string) (string, bool, error) {
	var row model.OperationPolicyPO
	err := r.data.DB.WithContext(ctx).Where("operation = ?", operation).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return row.PermissionCode, true, nil
}

func (r *policyRepo) SetOperationPolicy(ctx context.Context, operation, permissionCode string) error {
	row := &model.OperationPolicyPO{Operation: operation, PermissionCode: permissionCode}
	return r.data.DB.WithContext(ctx).
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(row).Error
}

func (r *policyRepo) DeleteOperationPolicy(ctx context.Context, operation string) error {
	return r.data.DB.WithContext(ctx).
		Where("operation = ?", operation).
		Delete(&model.OperationPolicyPO{}).Error
}

func (r *policyRepo) PermissionExists(ctx context.Context, code string) (bool, error) {
	var count int64
	if err := r.data.DB.WithContext(ctx).
		Model(&model.PermissionPO{}).
		Where("code = ?", code).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
