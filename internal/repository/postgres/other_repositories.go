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

// --- Property Image Repository ---

type propertyImageRepository struct {
	db *gorm.DB
}

func NewPropertyImageRepository(db *gorm.DB) repository.PropertyImageRepository {
	return &propertyImageRepository{db: db}
}

func (r *propertyImageRepository) Create(ctx context.Context, image *entity.PropertyImage) error {
	return r.db.WithContext(ctx).Create(image).Error
}

func (r *propertyImageRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.PropertyImage, error) {
	var image entity.PropertyImage
	err := r.db.WithContext(ctx).First(&image, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Property image")
		}
		return nil, err
	}
	return &image, nil
}

func (r *propertyImageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.PropertyImage{}, "id = ?", id).Error
}

func (r *propertyImageRepository) ListByPropertyID(ctx context.Context, propertyID uuid.UUID) ([]entity.PropertyImage, error) {
	var images []entity.PropertyImage
	err := r.db.WithContext(ctx).Where("property_id = ?", propertyID).Order("sort_order ASC").Find(&images).Error
	return images, err
}

func (r *propertyImageRepository) UpdateSortOrder(ctx context.Context, id uuid.UUID, sortOrder int) error {
	return r.db.WithContext(ctx).Model(&entity.PropertyImage{}).Where("id = ?", id).Update("sort_order", sortOrder).Error
}

func (r *propertyImageRepository) SetThumbnail(ctx context.Context, propertyID uuid.UUID, imageID uuid.UUID) error {
	db := conn(ctx, r.db)
	if err := db.Model(&entity.PropertyImage{}).
		Where("property_id = ? AND is_thumbnail", propertyID).
		Update("is_thumbnail", false).Error; err != nil {
		return err
	}
	url := ""
	if imageID != uuid.Nil {
		var img entity.PropertyImage
		if err := db.Where("id = ? AND property_id = ?", imageID, propertyID).First(&img).Error; err != nil {
			return err
		}
		if err := db.Model(&img).Update("is_thumbnail", true).Error; err != nil {
			return err
		}
		url = img.URL
	}
	return db.Model(&entity.Property{}).Where("id = ?", propertyID).Update("thumbnail", url).Error
}
