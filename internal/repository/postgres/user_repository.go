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

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository.
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("User")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("User")
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// UpdateWhatsAppNumber changes only the whatsapp_number column, so a concurrent
// login-failure update (lockout state) is never overwritten.
func (r *userRepository) UpdateWhatsAppNumber(ctx context.Context, id uuid.UUID, number string) error {
	res := r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("whatsapp_number", number)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errs.NotFound("User")
	}
	return nil
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.User{}, "id = ?", id).Error
}

func (r *userRepository) List(ctx context.Context, page, limit int, search, role, status string) ([]entity.User, int64, error) {
	var users []entity.User
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.User{})

	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) UpdateRole(ctx context.Context, id uuid.UUID, role string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("role", role).Error
}

func (r *userRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&entity.User{}).Where("id = ?", id).Update("status", status).Error
}

func (r *userRepository) RegisterLoginFailure(ctx context.Context, id uuid.UUID, maxAttempts int, lockFor time.Duration) error {
	return r.db.WithContext(ctx).Exec(`
UPDATE users SET
  failed_login_attempts = failed_login_attempts + 1,
  locked_until = CASE WHEN failed_login_attempts + 1 >= ? THEN NOW() + (? * INTERVAL '1 second') ELSE locked_until END
WHERE id = ?`, maxAttempts, int(lockFor.Seconds()), id).Error
}

func (r *userRepository) ResetLoginFailures(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Exec(
		"UPDATE users SET failed_login_attempts = 0, locked_until = NULL WHERE id = ?", id).Error
}
