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

type propertyDocumentRepository struct {
	db *gorm.DB
}

// NewPropertyDocumentRepository creates a property document repository.
func NewPropertyDocumentRepository(db *gorm.DB) repository.PropertyDocumentRepository {
	return &propertyDocumentRepository{db: db}
}

func (r *propertyDocumentRepository) Create(ctx context.Context, d *entity.PropertyDocument) error {
	err := conn(ctx, r.db).Create(d).Error
	if err != nil && isUniqueViolation(err) {
		return errs.Conflict("A document of this type already exists")
	}
	return err
}

func (r *propertyDocumentRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.PropertyDocument, error) {
	var d entity.PropertyDocument
	if err := conn(ctx, r.db).First(&d, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Document")
		}
		return nil, err
	}
	return &d, nil
}

func (r *propertyDocumentRepository) FindByType(ctx context.Context, propertyID uuid.UUID, docType string) (*entity.PropertyDocument, error) {
	var d entity.PropertyDocument
	err := conn(ctx, r.db).First(&d, "property_id = ? AND doc_type = ?", propertyID, docType).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *propertyDocumentRepository) ListByProperty(ctx context.Context, propertyID uuid.UUID) ([]entity.PropertyDocument, error) {
	var list []entity.PropertyDocument
	err := conn(ctx, r.db).Where("property_id = ?", propertyID).Order("created_at ASC").Find(&list).Error
	return list, err
}

func (r *propertyDocumentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return conn(ctx, r.db).Delete(&entity.PropertyDocument{}, "id = ?", id).Error
}
