package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/pkg/errs"
	"gorm.io/gorm"
)

type ownerRepository struct {
	db *gorm.DB
}

// NewOwnerRepository creates a new owner repository.
func NewOwnerRepository(db *gorm.DB) repository.OwnerRepository {
	return &ownerRepository{db: db}
}

func (r *ownerRepository) Create(ctx context.Context, owner *entity.Owner) error {
	return conn(ctx, r.db).Create(owner).Error
}

func (r *ownerRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Owner, error) {
	var owner entity.Owner
	if err := conn(ctx, r.db).First(&owner, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Owner")
		}
		return nil, err
	}
	return &owner, nil
}

func (r *ownerRepository) Update(ctx context.Context, owner *entity.Owner) error {
	return conn(ctx, r.db).Save(owner).Error
}
