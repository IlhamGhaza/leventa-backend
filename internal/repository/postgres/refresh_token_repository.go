package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/pkg/errs"
	"gorm.io/gorm"
)

type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository.
func NewRefreshTokenRepository(db *gorm.DB) repository.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *entity.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *refreshTokenRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error) {
	var token entity.RefreshToken
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Refresh token")
		}
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	// Atomic compare-and-set: only one concurrent caller can revoke (and so use) a token.
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&entity.RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", &now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.ErrUnauthorized
	}
	return nil
}

func (r *refreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&entity.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", &now).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&entity.RefreshToken{}).Error
}

func (r *refreshTokenRepository) PruneActive(ctx context.Context, userID uuid.UUID, keep int) error {
	return r.db.WithContext(ctx).Exec(`
UPDATE refresh_tokens SET revoked_at = NOW()
WHERE user_id = ? AND revoked_at IS NULL
  AND id NOT IN (SELECT id FROM refresh_tokens WHERE user_id = ? AND revoked_at IS NULL ORDER BY created_at DESC LIMIT ?)`,
		userID, userID, keep).Error
}
