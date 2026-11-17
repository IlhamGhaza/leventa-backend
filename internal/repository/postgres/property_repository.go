package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/pkg/errs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type propertyRepository struct {
	db *gorm.DB
}

// NewPropertyRepository creates a new property repository.
func NewPropertyRepository(db *gorm.DB) repository.PropertyRepository {
	return &propertyRepository{db: db}
}

// priceExpr is the effective price used for filtering/sorting.
const priceExpr = "(CASE WHEN properties.transaction_type = 'rent' THEN properties.rental_price ELSE properties.price END)"

// marketingContact loads only the columns needed to build a contact link.
func marketingContact(db *gorm.DB) *gorm.DB {
	return db.Select("id", "name", "whatsapp_number")
}

func (r *propertyRepository) Create(ctx context.Context, property *entity.Property) error {
	err := conn(ctx, r.db).Omit(clause.Associations).Create(property).Error
	if err != nil && isUniqueViolation(err) {
		return errs.Conflict("Property code already exists")
	}
	return err
}

func (r *propertyRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Property, error) {
	var property entity.Property
	err := conn(ctx, r.db).
		Preload("Images", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).
		Preload("Province").
		Preload("City").
		Preload("District").
		Preload("Facilities").
		Preload("PrimaryMarketing", marketingContact).
		First(&property, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Property")
		}
		return nil, err
	}
	return &property, nil
}

func (r *propertyRepository) FindByCode(ctx context.Context, code string) (*entity.Property, error) {
	var property entity.Property
	err := conn(ctx, r.db).Where("code = ?", code).First(&property).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Property")
		}
		return nil, err
	}
	return &property, nil
}

func (r *propertyRepository) Update(ctx context.Context, property *entity.Property) error {
	// Omit relations and immutable columns: ownership can never change here.
	return conn(ctx, r.db).
		Omit(clause.Associations, "primary_marketing_id", "owner_id", "code").
		Save(property).Error
}

func (r *propertyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return conn(ctx, r.db).Delete(&entity.Property{}, "id = ?", id).Error
}

func (r *propertyRepository) List(ctx context.Context, f repository.PropertyFilter) ([]entity.Property, int64, error) {
	var properties []entity.Property
	var total int64

	q := conn(ctx, r.db).Model(&entity.Property{})

	switch f.Visibility {
	case repository.VisibilityAvailable:
		q = q.Where("properties.status IN ?", []string{entity.PropertyStatusAvailable, entity.PropertyStatusReserved})
	case repository.VisibilityAvailableAndHistory:
		q = q.Where("properties.status IN ?", []string{entity.PropertyStatusAvailable, entity.PropertyStatusReserved, entity.PropertyStatusSold, entity.PropertyStatusRented})
	case repository.VisibilityHistory:
		q = q.Where("properties.status IN ?", []string{entity.PropertyStatusSold, entity.PropertyStatusRented})
	case repository.VisibilityAll:
		if f.Status != "" {
			q = q.Where("properties.status = ?", f.Status)
		} else {
			// Cancelled/archived listings stay out of lists unless asked for explicitly.
			q = q.Where("properties.status <> ?", entity.PropertyStatusArchived)
		}
	}

	if f.PrimaryOf != nil {
		q = q.Where("properties.primary_marketing_id = ?", *f.PrimaryOf)
	}
	if f.CoListingOf != nil {
		q = q.Where("EXISTS (SELECT 1 FROM property_marketing_assignments a WHERE a.property_id = properties.id AND a.marketing_id = ? AND a.assignment_type = ?)",
			*f.CoListingOf, entity.AssignmentTypeCoListing)
	}

	if s := strings.TrimSpace(f.Search); s != "" {
		like := "%" + escapeLike(s) + "%"
		// Restricted address columns are deliberately never searched.
		q = q.Where("(properties.name ILIKE ? OR properties.code ILIKE ? OR properties.general_location ILIKE ?)", like, like, like)
	}
	if f.Type != "" {
		q = q.Where("properties.type = ?", f.Type)
	}
	if f.TransactionType != "" {
		q = q.Where("properties.transaction_type = ?", f.TransactionType)
	}
	if f.ProvinceID != nil {
		q = q.Where("properties.province_id = ?", *f.ProvinceID)
	}
	if f.CityID != nil {
		q = q.Where("properties.city_id = ?", *f.CityID)
	}
	if len(f.FacilityIDs) > 0 {
		q = q.Where("properties.id IN (SELECT property_id FROM property_facilities WHERE facility_id IN ? GROUP BY property_id HAVING COUNT(DISTINCT facility_id) = ?)",
			f.FacilityIDs, len(f.FacilityIDs))
	}
	if f.DistrictID != nil {
		q = q.Where("properties.district_id = ?", *f.DistrictID)
	}
	if f.MinPrice != nil {
		q = q.Where(priceExpr+" >= ?", *f.MinPrice)
	}
	if f.MaxPrice != nil {
		q = q.Where(priceExpr+" <= ?", *f.MaxPrice)
	}
	if f.MinLandArea != nil {
		q = q.Where("properties.land_area >= ?", *f.MinLandArea)
	}
	if f.MaxLandArea != nil {
		q = q.Where("properties.land_area <= ?", *f.MaxLandArea)
	}
	if f.MinBuildingArea != nil {
		q = q.Where("properties.building_area >= ?", *f.MinBuildingArea)
	}
	if f.MaxBuildingArea != nil {
		q = q.Where("properties.building_area <= ?", *f.MaxBuildingArea)
	}
	if f.Bedrooms != nil {
		q = q.Where("properties.bedrooms >= ?", *f.Bedrooms)
	}
	if f.Bathrooms != nil {
		q = q.Where("properties.bathrooms >= ?", *f.Bathrooms)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.
		Preload("Province").
		Preload("City").
		Preload("District").
		Preload("Facilities").
		Preload("PrimaryMarketing", marketingContact).
		Offset((f.Page - 1) * f.Limit).Limit(f.Limit).
		Order(orderClause(f.Sort)).
		Find(&properties).Error
	if err != nil {
		return nil, 0, err
	}
	return properties, total, nil
}

// orderClause maps an allowlisted sort key to an ORDER BY clause.
func orderClause(sort string) string {
	switch sort {
	case "price_asc":
		return priceExpr + " ASC, properties.id ASC"
	case "price_desc":
		return priceExpr + " DESC, properties.id ASC"
	case "oldest":
		return "properties.created_at ASC, properties.id ASC"
	case "name_asc":
		return "properties.name ASC, properties.id ASC"
	case "name_desc":
		return "properties.name DESC, properties.id ASC"
	default:
		return "properties.created_at DESC, properties.id ASC"
	}
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func (r *propertyRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return conn(ctx, r.db).Model(&entity.Property{}).Where("id = ?", id).Update("status", status).Error
}

func (r *propertyRepository) UpdateStatusTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, status string) error {
	return tx.WithContext(ctx).Model(&entity.Property{}).Where("id = ?", id).Update("status", status).Error
}

func (r *propertyRepository) HasDeal(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := conn(ctx, r.db).Model(&entity.Deal{}).Where("property_id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *propertyRepository) LockByID(ctx context.Context, id uuid.UUID) error {
	var p entity.Property
	err := conn(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&p, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errs.NotFound("Property")
	}
	return err
}

func (r *propertyRepository) SetFacilities(ctx context.Context, propertyID uuid.UUID, facilities []entity.Facility) error {
	return conn(ctx, r.db).Model(&entity.Property{ID: propertyID}).Association("Facilities").Replace(facilities)
}

func (r *propertyRepository) SetPrimaryMarketing(ctx context.Context, propertyID, marketingID uuid.UUID) error {
	db := conn(ctx, r.db)
	if err := db.Model(&entity.Property{}).Where("id = ?", propertyID).Update("primary_marketing_id", marketingID).Error; err != nil {
		return err
	}
	return db.Exec("UPDATE owners SET created_by = ?, updated_at = NOW() WHERE id = (SELECT owner_id FROM properties WHERE id = ?)", marketingID, propertyID).Error
}
