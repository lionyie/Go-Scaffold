package database

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/example/scaffold/services/internal/biz"
	"github.com/example/scaffold/services/internal/data/model"
)

type refreshTokenRepo struct {
	data *Data
}

// NewRefreshTokenRepo returns the database-backed refresh token repository.
func NewRefreshTokenRepo(d *Data) biz.RefreshTokenRepo {
	return &refreshTokenRepo{data: d}
}

func (r *refreshTokenRepo) Create(ctx context.Context, token *biz.RefreshToken) error {
	po := &model.RefreshTokenPO{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		SessionID: token.SessionID,
		ExpiresAt: time.UnixMilli(token.ExpiresAt),
	}
	return r.data.DB.WithContext(ctx).Create(po).Error
}

func (r *refreshTokenRepo) GetByHash(ctx context.Context, hash string) (*biz.RefreshToken, error) {
	var po model.RefreshTokenPO
	if err := r.data.DB.WithContext(ctx).Where("token_hash = ?", hash).First(&po).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, biz.ErrTokenInvalid()
		}
		return nil, err
	}
	return &biz.RefreshToken{
		ID:        po.ID,
		UserID:    po.UserID,
		TokenHash: po.TokenHash,
		SessionID: po.SessionID,
		ExpiresAt: po.ExpiresAt.UnixMilli(),
		Revoked:   po.RevokedAt != nil,
	}, nil
}

func (r *refreshTokenRepo) RevokeByHash(ctx context.Context, hash string) error {
	now := time.Now()
	return r.data.DB.WithContext(ctx).
		Model(&model.RefreshTokenPO{}).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		Update("revoked_at", now).Error
}

func (r *refreshTokenRepo) RevokeBySession(ctx context.Context, userID, sessionID string) error {
	now := time.Now()
	return r.data.DB.WithContext(ctx).
		Model(&model.RefreshTokenPO{}).
		Where("user_id = ? AND session_id = ? AND revoked_at IS NULL", userID, sessionID).
		Update("revoked_at", now).Error
}
