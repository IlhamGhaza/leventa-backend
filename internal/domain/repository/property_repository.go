package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"gorm.io/gorm"
)

// PropertyVisibility limits which statuses a query may return.
type PropertyVisibility int

const (
	// VisibilityAvailable returns only publicly available listings (available, reserved).
	VisibilityAvailable PropertyVisibility = iota
	// VisibilityHistory returns only sold/rented historical listings.
	VisibilityHistory
	// VisibilityAvailableAndHistory returns available, reserved, sold and rented listings.
	VisibilityAvailableAndHistory
	// VisibilityAll returns every non-deleted status (internal roles only).
	VisibilityAll
)

// PropertyFilter holds filter options for property listing.
type PropertyFilter struct {
	Search          string
	Type            string
	TransactionType string
	Status          string // honoured only with VisibilityAll
	ProvinceID      *uuid.UUID
	CityID          *uuid.UUID
	DistrictID      *uuid.UUID
	FacilityIDs     []uuid.UUID // listing must have ALL of these facilities
	MinPrice        *float64
	MaxPrice        *float64
	MinLandArea     *float64
	MaxLandArea     *float64
	MinBuildingArea *float64
	MaxBuildingArea *float64
	Bedrooms        *int
	Bathrooms       *int
	Visibility      PropertyVisibility
	PrimaryOf       *uuid.UUID // only listings where this user is primary marketing
	CoListingOf     *uuid.UUID // only listings where this user is co-listing marketing
	Sort            string
	Page            int
	Limit           int
}

// PropertyRepository defines the interface for property data access.
type PropertyRepository interface {
	Create(ctx context.Context, property *entity.Property) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Property, error)
	FindByCode(ctx context.Context, code string) (*entity.Property, error)
	Update(ctx context.Context, property *entity.Property) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter PropertyFilter) ([]entity.Property, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateStatusTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, status string) error
	HasDeal(ctx context.Context, id uuid.UUID) (bool, error)
	// LockByID takes a row lock (SELECT ... FOR UPDATE); call inside a transaction.
	LockByID(ctx context.Context, id uuid.UUID) error
	// SetFacilities replaces the facilities of a property.
	// SetPrimaryMarketing changes the primary marketing owner and the creator of the listing's owner record.
	SetPrimaryMarketing(ctx context.Context, propertyID, marketingID uuid.UUID) error
	SetFacilities(ctx context.Context, propertyID uuid.UUID, facilities []entity.Facility) error
}

// PropertyImageRepository defines the interface for property image data access.
type PropertyImageRepository interface {
	Create(ctx context.Context, image *entity.PropertyImage) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.PropertyImage, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListByPropertyID(ctx context.Context, propertyID uuid.UUID) ([]entity.PropertyImage, error)
	UpdateSortOrder(ctx context.Context, id uuid.UUID, sortOrder int) error
	// SetThumbnail makes imageID the only thumbnail and mirrors its URL into
	// properties.thumbnail; uuid.Nil clears both.
	SetThumbnail(ctx context.Context, propertyID uuid.UUID, imageID uuid.UUID) error
}

// PropertyDocumentRepository defines ownership document metadata access.
type PropertyDocumentRepository interface {
	Create(ctx context.Context, d *entity.PropertyDocument) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.PropertyDocument, error)
	// FindByType returns nil, nil when no document of that type exists.
	FindByType(ctx context.Context, propertyID uuid.UUID, docType string) (*entity.PropertyDocument, error)
	ListByProperty(ctx context.Context, propertyID uuid.UUID) ([]entity.PropertyDocument, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// OwnerRepository defines owner data access. Owners are reachable only through
// the property they belong to.
type OwnerRepository interface {
	Create(ctx context.Context, owner *entity.Owner) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Owner, error)
	Update(ctx context.Context, owner *entity.Owner) error
}

// CoListingRepository manages property marketing assignments.
type CoListingRepository interface {
	Create(ctx context.Context, a *entity.PropertyMarketingAssignment) error
	// Find returns nil, nil when the user is not assigned to the property.
	Find(ctx context.Context, propertyID, marketingID uuid.UUID) (*entity.PropertyMarketingAssignment, error)
	ListByProperty(ctx context.Context, propertyID uuid.UUID) ([]entity.PropertyMarketingAssignment, error)
	CountCoListing(ctx context.Context, propertyID uuid.UUID) (int64, error)
	DeleteCoListing(ctx context.Context, propertyID, marketingID uuid.UUID) (bool, error)
	// Delete removes any assignment (primary or co-listing) of the user on the property.
	Delete(ctx context.Context, propertyID, marketingID uuid.UUID) error
}
