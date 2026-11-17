package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
)

// RefreshTokenRepository defines the interface for refresh token data access.
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *entity.RefreshToken) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*entity.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	// PruneActive revokes the oldest active tokens of a user beyond keep.
	PruneActive(ctx context.Context, userID uuid.UUID, keep int) error
}
