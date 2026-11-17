package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"gorm.io/gorm"
)

type coListingRepository struct {
	db *gorm.DB
}

// NewCoListingRepository creates a new assignment repository.
func NewCoListingRepository(db *gorm.DB) repository.CoListingRepository {
	return &coListingRepository{db: db}
}

func (r *coListingRepository) Create(ctx context.Context, a *entity.PropertyMarketingAssignment) error {
	return conn(ctx, r.db).Omit("Marketing").Create(a).Error
}

func (r *coListingRepository) Find(ctx context.Context, propertyID, marketingID uuid.UUID) (*entity.PropertyMarketingAssignment, error) {
	var a entity.PropertyMarketingAssignment
	err := conn(ctx, r.db).Where("property_id = ? AND marketing_id = ?", propertyID, marketingID).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *coListingRepository) ListByProperty(ctx context.Context, propertyID uuid.UUID) ([]entity.PropertyMarketingAssignment, error) {
	var list []entity.PropertyMarketingAssignment
	err := conn(ctx, r.db).
		Preload("Marketing", func(db *gorm.DB) *gorm.DB { return db.Select("id", "name") }).
		Where("property_id = ?", propertyID).
		Order("CASE assignment_type WHEN 'primary' THEN 0 ELSE 1 END, assigned_at ASC").
		Find(&list).Error
	return list, err
}

func (r *coListingRepository) CountCoListing(ctx context.Context, propertyID uuid.UUID) (int64, error) {
	var n int64
	err := conn(ctx, r.db).Model(&entity.PropertyMarketingAssignment{}).
		Where("property_id = ? AND assignment_type = ?", propertyID, entity.AssignmentTypeCoListing).
		Count(&n).Error
	return n, err
}

func (r *coListingRepository) DeleteCoListing(ctx context.Context, propertyID, marketingID uuid.UUID) (bool, error) {
	res := conn(ctx, r.db).
		Where("property_id = ? AND marketing_id = ? AND assignment_type = ?", propertyID, marketingID, entity.AssignmentTypeCoListing).
		Delete(&entity.PropertyMarketingAssignment{})
	return res.RowsAffected > 0, res.Error
}

// isUniqueViolation reports a PostgreSQL unique_violation (23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return strings.Contains(err.Error(), "SQLSTATE 23505")
}

func (r *coListingRepository) Delete(ctx context.Context, propertyID, marketingID uuid.UUID) error {
	return conn(ctx, r.db).
		Where("property_id = ? AND marketing_id = ?", propertyID, marketingID).
		Delete(&entity.PropertyMarketingAssignment{}).Error
}
