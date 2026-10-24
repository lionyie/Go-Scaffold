package database

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/data/model"
	"github.com/example/scaffold/services/internal/pkg/id"
)

type oauthAccountRepo struct {
	data *Data
}

// NewOAuthAccountRepo returns the provider-link repository.
func NewOAuthAccountRepo(d *Data) biz.OAuthAccountRepo {
	return &oauthAccountRepo{data: d}
}

func (r *oauthAccountRepo) FindUserID(ctx context.Context, provider, providerUserID string) (string, error) {
	var po model.OAuthAccountPO
	err := r.data.DB.WithContext(ctx).
		Where("provider = ? AND provider_user_id = ?", provider, providerUserID).
		First(&po).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return po.UserID, nil
}

func (r *oauthAccountRepo) Link(ctx context.Context, account *biz.OAuthAccount) error {
	po := &model.OAuthAccountPO{
		ID:             id.New(),
		UserID:         account.UserID,
		Provider:       account.Provider,
		ProviderUserID: account.ProviderUserID,
		Email:          account.Email,
		EmailVerified:  account.EmailVerified,
		Name:           account.Name,
		Avatar:         account.Avatar,
	}
	return r.data.DB.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "provider"}, {Name: "provider_user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"user_id", "email", "email_verified", "name", "avatar"}),
		}).
		Create(po).Error
}

func (r *oauthAccountRepo) ListByUser(ctx context.Context, userID string) ([]*biz.OAuthAccount, error) {
	var rows []model.OAuthAccountPO
	if err := r.data.DB.WithContext(ctx).Where("user_id = ?", userID).Order("provider").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*biz.OAuthAccount, 0, len(rows))
	for i := range rows {
		out = append(out, &biz.OAuthAccount{
			Provider:       rows[i].Provider,
			ProviderUserID: rows[i].ProviderUserID,
			Email:          rows[i].Email,
			EmailVerified:  rows[i].EmailVerified,
			Name:           rows[i].Name,
			Avatar:         rows[i].Avatar,
			UserID:         rows[i].UserID,
		})
	}
	return out, nil
}

func (r *oauthAccountRepo) Unlink(ctx context.Context, userID, provider string) error {
	return r.data.DB.WithContext(ctx).
		Where("user_id = ? AND provider = ?", userID, provider).
		Delete(&model.OAuthAccountPO{}).Error
}
