package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, page, limit int, search, role, status string) ([]entity.User, int64, error)
	UpdateRole(ctx context.Context, id uuid.UUID, role string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	// UpdateWhatsAppNumber changes only the whatsapp_number column.
	UpdateWhatsAppNumber(ctx context.Context, id uuid.UUID, number string) error
	// RegisterLoginFailure atomically counts a failed login and locks the account
	// for lockFor once maxAttempts is reached.
	RegisterLoginFailure(ctx context.Context, id uuid.UUID, maxAttempts int, lockFor time.Duration) error
	// ResetLoginFailures clears the failure counter and any lock.
	ResetLoginFailures(ctx context.Context, id uuid.UUID) error
}
