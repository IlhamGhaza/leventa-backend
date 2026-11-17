package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
)

// DealRepository defines the interface for deal data access.
type DealRepository interface {
	Create(ctx context.Context, deal *entity.Deal) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Deal, error)
	List(ctx context.Context, page, limit int, salesID *uuid.UUID) ([]entity.Deal, int64, error)
	ExistsByPropertyID(ctx context.Context, propertyID uuid.UUID) (bool, error)
}

// CommissionRepository defines the interface for commission data access.
type CommissionRepository interface {
	Create(ctx context.Context, commission *entity.Commission) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Commission, error)
	Update(ctx context.Context, commission *entity.Commission) error
	List(ctx context.Context, page, limit int, salesID *uuid.UUID, status string) ([]entity.Commission, int64, error)
}

// AuditLogRepository defines the interface for audit log data access.
type AuditLogRepository interface {
	Create(ctx context.Context, log *entity.AuditLog) error
	List(ctx context.Context, page, limit int, module, action string, userID *uuid.UUID) ([]entity.AuditLog, int64, error)
}

// MasterDataRepository defines the interface for master data access.
type MasterDataRepository interface {
	// Provinces
	CreateProvince(ctx context.Context, province *entity.Province) error
	ListProvinces(ctx context.Context) ([]entity.Province, error)
	FindProvinceByID(ctx context.Context, id uuid.UUID) (*entity.Province, error)
	UpdateProvince(ctx context.Context, province *entity.Province) error
	DeleteProvince(ctx context.Context, id uuid.UUID) error

	// Cities
	CreateCity(ctx context.Context, city *entity.City) error
	ListCities(ctx context.Context, provinceID *uuid.UUID) ([]entity.City, error)
	FindCityByID(ctx context.Context, id uuid.UUID) (*entity.City, error)
	UpdateCity(ctx context.Context, city *entity.City) error
	DeleteCity(ctx context.Context, id uuid.UUID) error

	// Districts
	CreateDistrict(ctx context.Context, d *entity.District) error
	ListDistricts(ctx context.Context, cityID *uuid.UUID) ([]entity.District, error)
	FindDistrictByID(ctx context.Context, id uuid.UUID) (*entity.District, error)
	UpdateDistrict(ctx context.Context, d *entity.District) error
	DeleteDistrict(ctx context.Context, id uuid.UUID) error

	// Facilities
	CreateFacility(ctx context.Context, f *entity.Facility) error
	ListFacilities(ctx context.Context) ([]entity.Facility, error)
	FindFacilityByID(ctx context.Context, id uuid.UUID) (*entity.Facility, error)
	UpdateFacility(ctx context.Context, f *entity.Facility) error
	DeleteFacility(ctx context.Context, id uuid.UUID) error
	// FindFacilitiesByIDs returns the facilities that exist for the given IDs.
	FindFacilitiesByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.Facility, error)
}
