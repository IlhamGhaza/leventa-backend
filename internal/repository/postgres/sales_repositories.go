package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/pkg/errs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// --- Deal Repository ---

type dealRepository struct {
	db *gorm.DB
}

func NewDealRepository(db *gorm.DB) repository.DealRepository {
	return &dealRepository{db: db}
}

func (r *dealRepository) Create(ctx context.Context, deal *entity.Deal) error {
	err := conn(ctx, r.db).Omit(clause.Associations).Create(deal).Error
	if err != nil && isUniqueViolation(err) {
		return errs.Conflict("An outcome is already recorded for this property")
	}
	return err
}

func (r *dealRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Deal, error) {
	var deal entity.Deal
	err := r.db.WithContext(ctx).
		Preload("Property").Preload("Sales").
		First(&deal, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Deal")
		}
		return nil, err
	}
	return &deal, nil
}

func (r *dealRepository) List(ctx context.Context, page, limit int, salesID *uuid.UUID) ([]entity.Deal, int64, error) {
	var deals []entity.Deal
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Deal{})
	if salesID != nil {
		query = query.Where("sales_id = ?", *salesID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Preload("Property").Preload("Sales").
		Offset(offset).Limit(limit).Order("created_at DESC").Find(&deals).Error; err != nil {
		return nil, 0, err
	}

	return deals, total, nil
}

func (r *dealRepository) ExistsByPropertyID(ctx context.Context, propertyID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Deal{}).Where("property_id = ?", propertyID).Count(&count).Error
	return count > 0, err
}

// --- Commission Repository ---

type commissionRepository struct {
	db *gorm.DB
}

func NewCommissionRepository(db *gorm.DB) repository.CommissionRepository {
	return &commissionRepository{db: db}
}

func (r *commissionRepository) Create(ctx context.Context, commission *entity.Commission) error {
	return conn(ctx, r.db).Omit(clause.Associations).Create(commission).Error
}

func (r *commissionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Commission, error) {
	var commission entity.Commission
	err := r.db.WithContext(ctx).Preload("Deal").Preload("Sales").First(&commission, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Commission")
		}
		return nil, err
	}
	return &commission, nil
}

func (r *commissionRepository) Update(ctx context.Context, commission *entity.Commission) error {
	return r.db.WithContext(ctx).Omit("Deal", "Sales").Save(commission).Error
}

func (r *commissionRepository) List(ctx context.Context, page, limit int, salesID *uuid.UUID, status string) ([]entity.Commission, int64, error) {
	var commissions []entity.Commission
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Commission{})
	if salesID != nil {
		query = query.Where("sales_id = ?", *salesID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Preload("Deal").Preload("Sales").
		Offset(offset).Limit(limit).Order("created_at DESC").Find(&commissions).Error; err != nil {
		return nil, 0, err
	}

	return commissions, total, nil
}

// --- Audit Log Repository ---

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) repository.AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, log *entity.AuditLog) error {
	return conn(ctx, r.db).Create(log).Error
}

func (r *auditLogRepository) List(ctx context.Context, page, limit int, module, action string, userID *uuid.UUID) ([]entity.AuditLog, int64, error) {
	var logs []entity.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.AuditLog{})
	if module != "" {
		query = query.Where("module = ?", module)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Preload("User").Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// --- Master Data Repository ---

type masterDataRepository struct {
	db *gorm.DB
}

func NewMasterDataRepository(db *gorm.DB) repository.MasterDataRepository {
	return &masterDataRepository{db: db}
}

func (r *masterDataRepository) CreateProvince(ctx context.Context, province *entity.Province) error {
	return r.db.WithContext(ctx).Create(province).Error
}

func (r *masterDataRepository) ListProvinces(ctx context.Context) ([]entity.Province, error) {
	var provinces []entity.Province
	err := r.db.WithContext(ctx).Order("name ASC").Find(&provinces).Error
	return provinces, err
}

func (r *masterDataRepository) FindProvinceByID(ctx context.Context, id uuid.UUID) (*entity.Province, error) {
	var province entity.Province
	err := r.db.WithContext(ctx).Preload("Cities").First(&province, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Province")
		}
		return nil, err
	}
	return &province, nil
}

func (r *masterDataRepository) UpdateProvince(ctx context.Context, province *entity.Province) error {
	return r.db.WithContext(ctx).Save(province).Error
}

func (r *masterDataRepository) DeleteProvince(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Province{}, "id = ?", id).Error
}

func (r *masterDataRepository) CreateCity(ctx context.Context, city *entity.City) error {
	return r.db.WithContext(ctx).Create(city).Error
}

func (r *masterDataRepository) ListCities(ctx context.Context, provinceID *uuid.UUID) ([]entity.City, error) {
	var cities []entity.City
	query := r.db.WithContext(ctx)
	if provinceID != nil {
		query = query.Where("province_id = ?", *provinceID)
	}
	err := query.Preload("Province").Order("name ASC").Find(&cities).Error
	return cities, err
}

func (r *masterDataRepository) FindCityByID(ctx context.Context, id uuid.UUID) (*entity.City, error) {
	var city entity.City
	err := r.db.WithContext(ctx).Preload("Province").First(&city, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("City")
		}
		return nil, err
	}
	return &city, nil
}

func (r *masterDataRepository) UpdateCity(ctx context.Context, city *entity.City) error {
	return r.db.WithContext(ctx).Save(city).Error
}

func (r *masterDataRepository) DeleteCity(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.City{}, "id = ?", id).Error
}

// --- Districts & facilities (master data) ---

func (r *masterDataRepository) CreateDistrict(ctx context.Context, d *entity.District) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *masterDataRepository) ListDistricts(ctx context.Context, cityID *uuid.UUID) ([]entity.District, error) {
	var list []entity.District
	q := r.db.WithContext(ctx)
	if cityID != nil {
		q = q.Where("city_id = ?", *cityID)
	}
	err := q.Order("name ASC").Find(&list).Error
	return list, err
}

func (r *masterDataRepository) FindDistrictByID(ctx context.Context, id uuid.UUID) (*entity.District, error) {
	var d entity.District
	if err := r.db.WithContext(ctx).First(&d, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("District")
		}
		return nil, err
	}
	return &d, nil
}

func (r *masterDataRepository) UpdateDistrict(ctx context.Context, d *entity.District) error {
	return r.db.WithContext(ctx).Omit("City").Save(d).Error
}

func (r *masterDataRepository) DeleteDistrict(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.District{}, "id = ?", id).Error
}

func (r *masterDataRepository) CreateFacility(ctx context.Context, f *entity.Facility) error {
	return r.db.WithContext(ctx).Create(f).Error
}

func (r *masterDataRepository) ListFacilities(ctx context.Context) ([]entity.Facility, error) {
	var list []entity.Facility
	err := r.db.WithContext(ctx).Order("name ASC").Find(&list).Error
	return list, err
}

func (r *masterDataRepository) FindFacilityByID(ctx context.Context, id uuid.UUID) (*entity.Facility, error) {
	var f entity.Facility
	if err := r.db.WithContext(ctx).First(&f, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errs.NotFound("Facility")
		}
		return nil, err
	}
	return &f, nil
}

func (r *masterDataRepository) UpdateFacility(ctx context.Context, f *entity.Facility) error {
	return r.db.WithContext(ctx).Save(f).Error
}

func (r *masterDataRepository) DeleteFacility(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Facility{}, "id = ?", id).Error
}

func (r *masterDataRepository) FindFacilitiesByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.Facility, error) {
	var list []entity.Facility
	if len(ids) == 0 {
		return list, nil
	}
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&list).Error
	return list, err
}
