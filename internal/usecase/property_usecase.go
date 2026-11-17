package usecase

import (
	"context"
	"mime/multipart"
	"strings"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/internal/service"
	"github.com/leventa-backend/internal/storage"
	"github.com/leventa-backend/pkg/errs"
)

// MaxPageSize is the largest page size accepted by list endpoints.
const MaxPageSize = 100

// Limits that keep public search cheap: anonymous callers can reach it.
const (
	MaxSearchLength = 100   // characters in the free-text search
	MaxListOffset   = 10000 // (page-1)*limit; deeper pages must narrow the filters
	MaxFacilityIDs  = 20
)

// PropertyUsecase handles property business logic.
type PropertyUsecase struct {
	propertyAccess
	imageRepo    repository.PropertyImageRepository
	ownerRepo    repository.OwnerRepository
	masterRepo   repository.MasterDataRepository
	docRepo      repository.PropertyDocumentRepository
	tx           repository.Transactor
	storage      storage.FileStorage
	auditService *service.AuditService
}

// NewPropertyUsecase creates a new property usecase.
func NewPropertyUsecase(
	propertyRepo repository.PropertyRepository,
	imageRepo repository.PropertyImageRepository,
	ownerRepo repository.OwnerRepository,
	coRepo repository.CoListingRepository,
	masterRepo repository.MasterDataRepository,
	docRepo repository.PropertyDocumentRepository,
	tx repository.Transactor,
	storage storage.FileStorage,
	auditService *service.AuditService,
) *PropertyUsecase {
	return &PropertyUsecase{
		propertyAccess: propertyAccess{propertyRepo: propertyRepo, coRepo: coRepo},
		imageRepo:      imageRepo,
		ownerRepo:      ownerRepo,
		masterRepo:     masterRepo,
		docRepo:        docRepo,
		tx:             tx,
		storage:        storage,
		auditService:   auditService,
	}
}

// OwnerInput holds owner PII supplied by the primary marketing user.
type OwnerInput struct {
	Name           string `json:"name" validate:"required,min=2,max=255"`
	Email          string `json:"email" validate:"omitempty,email,max=255"`
	Phone          string `json:"phone" validate:"omitempty,max=50"`
	PrivateAddress string `json:"private_address" validate:"omitempty,max=2000"`
}

// CreatePropertyInput holds the data for creating a property.
// There is intentionally no primary_marketing_id / owner_id: ownership comes
// from the authenticated actor.
type CreatePropertyInput struct {
	Owner           OwnerInput  `json:"owner"`
	Code            string      `json:"code" validate:"required,min=2,max=50"`
	Name            string      `json:"name" validate:"required,min=2,max=255"`
	Type            string      `json:"type" validate:"required,oneof=house apartment shop_house factory"`
	TransactionType string      `json:"transaction_type" validate:"required,oneof=sale rent"`
	Status          string      `json:"status" validate:"omitempty,oneof=draft available"`
	Price           *float64    `json:"price" validate:"omitempty,gte=0"`
	RentalPrice     *float64    `json:"rental_price" validate:"omitempty,gte=0"`
	Description     string      `json:"description"`
	ProvinceID      *uuid.UUID  `json:"province_id"`
	CityID          *uuid.UUID  `json:"city_id"`
	DistrictID      *uuid.UUID  `json:"district_id"`
	FacilityIDs     []uuid.UUID `json:"facility_ids"`
	GeneralLocation string      `json:"general_location" validate:"omitempty,max=255"`
	BlockNumber     string      `json:"block_number" validate:"omitempty,max=50"`
	ClusterName     string      `json:"cluster_name" validate:"omitempty,max=255"`
	StreetName      string      `json:"street_name" validate:"omitempty,max=255"`
	HouseNumber     string      `json:"house_number" validate:"omitempty,max=50"`
	FloorNumber     string      `json:"floor_number" validate:"omitempty,max=50"`
	RoomNumber      string      `json:"room_number" validate:"omitempty,max=50"`
	Latitude        *float64    `json:"latitude" validate:"omitempty,gte=-90,lte=90"`
	Longitude       *float64    `json:"longitude" validate:"omitempty,gte=-180,lte=180"`
	LandArea        *float64    `json:"land_area" validate:"omitempty,gte=0"`
	BuildingArea    *float64    `json:"building_area" validate:"omitempty,gte=0"`
	Floor           *int        `json:"floor"`
	Bedrooms        int         `json:"bedrooms" validate:"gte=0"`
	Bathrooms       int         `json:"bathrooms" validate:"gte=0"`
	YearBuilt       *int        `json:"year_built"`
	PropertyDetails
}

// UpdatePropertyInput is a partial update (PATCH). Nil fields are unchanged.
// Code, owner and primary marketing cannot be changed here.
type UpdatePropertyInput struct {
	Name            *string      `json:"name" validate:"omitempty,min=2,max=255"`
	Type            *string      `json:"type" validate:"omitempty,oneof=house apartment shop_house factory"`
	TransactionType *string      `json:"transaction_type" validate:"omitempty,oneof=sale rent"`
	Price           *float64     `json:"price" validate:"omitempty,gte=0"`
	RentalPrice     *float64     `json:"rental_price" validate:"omitempty,gte=0"`
	Description     *string      `json:"description"`
	ProvinceID      *uuid.UUID   `json:"province_id"`
	CityID          *uuid.UUID   `json:"city_id"`
	DistrictID      *uuid.UUID   `json:"district_id"`
	FacilityIDs     *[]uuid.UUID `json:"facility_ids"`
	GeneralLocation *string      `json:"general_location" validate:"omitempty,max=255"`
	BlockNumber     *string      `json:"block_number" validate:"omitempty,max=50"`
	ClusterName     *string      `json:"cluster_name" validate:"omitempty,max=255"`
	StreetName      *string      `json:"street_name" validate:"omitempty,max=255"`
	HouseNumber     *string      `json:"house_number" validate:"omitempty,max=50"`
	FloorNumber     *string      `json:"floor_number" validate:"omitempty,max=50"`
	RoomNumber      *string      `json:"room_number" validate:"omitempty,max=50"`
	Latitude        *float64     `json:"latitude" validate:"omitempty,gte=-90,lte=90"`
	Longitude       *float64     `json:"longitude" validate:"omitempty,gte=-180,lte=180"`
	LandArea        *float64     `json:"land_area" validate:"omitempty,gte=0"`
	BuildingArea    *float64     `json:"building_area" validate:"omitempty,gte=0"`
	Floor           *int         `json:"floor"`
	Bedrooms        *int         `json:"bedrooms" validate:"omitempty,gte=0"`
	Bathrooms       *int         `json:"bathrooms" validate:"omitempty,gte=0"`
	YearBuilt       *int         `json:"year_built"`
	PropertyDetailsPatch
}

func validatePricing(transactionType string, price, rental *float64) error {
	if transactionType == entity.TransactionTypeSale && price == nil {
		return errs.Validation("price is required for a sale listing")
	}
	if transactionType == entity.TransactionTypeRent && rental == nil {
		return errs.Validation("rental_price is required for a rent listing")
	}
	return nil
}

// validateLocation checks province → city → district consistency.
func (uc *PropertyUsecase) validateLocation(ctx context.Context, province, city, district *uuid.UUID) error {
	var cityEnt *entity.City
	if city != nil {
		c, err := uc.masterRepo.FindCityByID(ctx, *city)
		if err != nil {
			return errs.Validation("city_id not found")
		}
		cityEnt = c
		if province != nil && c.ProvinceID != *province {
			return errs.Validation("city does not belong to the province")
		}
	}
	if province != nil {
		if _, err := uc.masterRepo.FindProvinceByID(ctx, *province); err != nil {
			return errs.Validation("province_id not found")
		}
	}
	if district != nil {
		d, err := uc.masterRepo.FindDistrictByID(ctx, *district)
		if err != nil {
			return errs.Validation("district_id not found")
		}
		if cityEnt == nil || d.CityID != cityEnt.ID {
			return errs.Validation("district must belong to the selected city")
		}
	}
	return nil
}

// resolveFacilities loads facilities and rejects unknown IDs.
func (uc *PropertyUsecase) resolveFacilities(ctx context.Context, ids []uuid.UUID) ([]entity.Facility, error) {
	uniq := map[uuid.UUID]bool{}
	var list []uuid.UUID
	for _, id := range ids {
		if !uniq[id] {
			uniq[id] = true
			list = append(list, id)
		}
	}
	found, err := uc.masterRepo.FindFacilitiesByIDs(ctx, list)
	if err != nil {
		return nil, errs.Internal(err)
	}
	if len(found) != len(list) {
		return nil, errs.Validation("facility_ids contains an unknown facility")
	}
	return found, nil
}

// validateCertificates checks the cert_* flags against the property type and,
// on update, refuses to clear a flag while its document is still stored.
func (uc *PropertyUsecase) validateCertificates(ctx context.Context, p *entity.Property, existing bool) error {
	for _, t := range entity.ValidDocTypes() {
		if p.CertFlag(t) && !entity.DocTypeAllowed(p.Type, t) {
			return errs.Validation("cert_" + t + " does not apply to a " + p.Type + " listing")
		}
		if existing && !p.CertFlag(t) {
			d, err := uc.docRepo.FindByType(ctx, p.ID, t)
			if err != nil {
				return errs.Internal(err)
			}
			if d != nil {
				return errs.Conflict("delete the " + t + " document before setting cert_" + t + " to false")
			}
		}
	}
	return nil
}

// Create creates a listing. The authenticated marketing user becomes primary owner.
func (uc *PropertyUsecase) Create(ctx context.Context, actor Actor, input CreatePropertyInput, meta RequestMeta) (*PrimaryProperty, error) {
	if !actor.IsSales() {
		return nil, errs.Forbidden("Only marketing users can create listings")
	}
	if err := validatePricing(input.TransactionType, input.Price, input.RentalPrice); err != nil {
		return nil, err
	}
	if (input.Latitude == nil) != (input.Longitude == nil) {
		return nil, errs.Validation("latitude and longitude must be provided together")
	}
	if existing, _ := uc.propertyRepo.FindByCode(ctx, input.Code); existing != nil {
		return nil, errs.Conflict("Property code already exists")
	}

	if err := uc.validateLocation(ctx, input.ProvinceID, input.CityID, input.DistrictID); err != nil {
		return nil, err
	}
	facilities, err := uc.resolveFacilities(ctx, input.FacilityIDs)
	if err != nil {
		return nil, err
	}

	status := input.Status
	if status == "" {
		status = entity.PropertyStatusAvailable
	}

	owner := &entity.Owner{
		Name:           input.Owner.Name,
		Email:          input.Owner.Email,
		Phone:          input.Owner.Phone,
		PrivateAddress: input.Owner.PrivateAddress,
		CreatedBy:      &actor.ID,
	}
	property := &entity.Property{
		PrimaryMarketingID: actor.ID, // from verified token only
		Code:               input.Code,
		Name:               input.Name,
		Type:               input.Type,
		TransactionType:    input.TransactionType,
		Status:             status,
		Price:              input.Price,
		RentalPrice:        input.RentalPrice,
		Description:        input.Description,
		ProvinceID:         input.ProvinceID,
		CityID:             input.CityID,
		DistrictID:         input.DistrictID,
		GeneralLocation:    input.GeneralLocation,
		BlockNumber:        input.BlockNumber,
		ClusterName:        input.ClusterName,
		StreetName:         input.StreetName,
		HouseNumber:        input.HouseNumber,
		FloorNumber:        input.FloorNumber,
		RoomNumber:         input.RoomNumber,
		Latitude:           input.Latitude,
		Longitude:          input.Longitude,
		LandArea:           input.LandArea,
		BuildingArea:       input.BuildingArea,
		Floor:              input.Floor,
		Bedrooms:           input.Bedrooms,
		Bathrooms:          input.Bathrooms,
		YearBuilt:          input.YearBuilt,
	}
	input.PropertyDetails.applyTo(property)
	if err := uc.validateCertificates(ctx, property, false); err != nil {
		return nil, err
	}

	err = uc.tx.Do(ctx, func(ctx context.Context) error {
		if err := uc.ownerRepo.Create(ctx, owner); err != nil {
			return err
		}
		property.OwnerID = owner.ID
		if err := uc.propertyRepo.Create(ctx, property); err != nil {
			return err
		}
		if err := uc.coRepo.Create(ctx, &entity.PropertyMarketingAssignment{
			PropertyID:     property.ID,
			MarketingID:    actor.ID,
			AssignmentType: entity.AssignmentTypePrimary,
			AssignedBy:     actor.ID,
		}); err != nil {
			return err
		}
		if len(facilities) > 0 {
			if err := uc.propertyRepo.SetFacilities(ctx, property.ID, facilities); err != nil {
				return err
			}
		}
		return uc.auditService.LogSync(ctx, &actor.ID, "created_property", "properties", &property.ID,
			map[string]string{"code": property.Code}, meta.IP, meta.UserAgent)
	})
	if err != nil {
		if _, ok := errs.IsAppError(err); ok {
			return nil, err
		}
		return nil, errs.Internal(err)
	}

	return uc.getPrimary(ctx, property.ID)
}

func (uc *PropertyUsecase) getPrimary(ctx context.Context, id uuid.UUID) (*PrimaryProperty, error) {
	p, err := uc.propertyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	team, err := uc.coRepo.ListByProperty(ctx, id)
	if err != nil {
		return nil, errs.Internal(err)
	}
	v := toPrimaryProperty(p, team)
	return &v, nil
}

// Update applies a partial update. Primary marketing owner only.
func (uc *PropertyUsecase) Update(ctx context.Context, actor Actor, id uuid.UUID, in UpdatePropertyInput, meta RequestMeta) (*PrimaryProperty, error) {
	p, err := uc.requirePrimary(ctx, actor, id)
	if err != nil {
		return nil, err
	}

	if in.Name != nil {
		p.Name = *in.Name
	}
	if in.Type != nil {
		p.Type = *in.Type
	}
	if in.TransactionType != nil {
		p.TransactionType = *in.TransactionType
	}
	if in.Price != nil {
		p.Price = in.Price
	}
	if in.RentalPrice != nil {
		p.RentalPrice = in.RentalPrice
	}
	if in.Description != nil {
		p.Description = *in.Description
	}
	if in.ProvinceID != nil {
		p.ProvinceID = in.ProvinceID
	}
	if in.CityID != nil {
		p.CityID = in.CityID
	}
	if in.DistrictID != nil {
		p.DistrictID = in.DistrictID
	}
	if in.ProvinceID != nil || in.CityID != nil || in.DistrictID != nil {
		if err := uc.validateLocation(ctx, p.ProvinceID, p.CityID, p.DistrictID); err != nil {
			return nil, err
		}
	}
	if in.GeneralLocation != nil {
		p.GeneralLocation = *in.GeneralLocation
	}
	if in.BlockNumber != nil {
		p.BlockNumber = *in.BlockNumber
	}
	if in.ClusterName != nil {
		p.ClusterName = *in.ClusterName
	}
	if in.StreetName != nil {
		p.StreetName = *in.StreetName
	}
	if in.HouseNumber != nil {
		p.HouseNumber = *in.HouseNumber
	}
	if in.FloorNumber != nil {
		p.FloorNumber = *in.FloorNumber
	}
	if in.RoomNumber != nil {
		p.RoomNumber = *in.RoomNumber
	}
	if in.Latitude != nil {
		p.Latitude = in.Latitude
	}
	if in.Longitude != nil {
		p.Longitude = in.Longitude
	}
	if (p.Latitude == nil) != (p.Longitude == nil) {
		return nil, errs.Validation("latitude and longitude must be provided together")
	}
	if in.LandArea != nil {
		p.LandArea = in.LandArea
	}
	if in.BuildingArea != nil {
		p.BuildingArea = in.BuildingArea
	}
	if in.Floor != nil {
		p.Floor = in.Floor
	}
	if in.Bedrooms != nil {
		p.Bedrooms = *in.Bedrooms
	}
	if in.Bathrooms != nil {
		p.Bathrooms = *in.Bathrooms
	}
	if in.YearBuilt != nil {
		p.YearBuilt = in.YearBuilt
	}
	in.PropertyDetailsPatch.applyTo(p)
	if err := uc.validateCertificates(ctx, p, true); err != nil {
		return nil, err
	}
	if err := validatePricing(p.TransactionType, p.Price, p.RentalPrice); err != nil {
		return nil, err
	}

	var facilities []entity.Facility
	if in.FacilityIDs != nil {
		if facilities, err = uc.resolveFacilities(ctx, *in.FacilityIDs); err != nil {
			return nil, err
		}
	}
	err = uc.tx.Do(ctx, func(ctx context.Context) error {
		if err := uc.propertyRepo.Update(ctx, p); err != nil {
			return err
		}
		if in.FacilityIDs != nil {
			return uc.propertyRepo.SetFacilities(ctx, id, facilities)
		}
		return nil
	})
	if err != nil {
		return nil, errs.Internal(err)
	}
	uc.auditService.Log(&actor.ID, "updated_property", "properties", &id, nil, meta.IP, meta.UserAgent)
	return uc.getPrimary(ctx, id)
}

// DeleteResult tells the caller what happened to the listing.
type DeleteResult struct {
	Archived bool
}

// Delete removes a listing. Primary marketing owner only. A listing that is
// part of a deal is archived instead of deleted (D-015).
func (uc *PropertyUsecase) Delete(ctx context.Context, actor Actor, id uuid.UUID, meta RequestMeta) (DeleteResult, error) {
	if _, err := uc.requirePrimary(ctx, actor, id); err != nil {
		return DeleteResult{}, err
	}
	hasDeal, err := uc.propertyRepo.HasDeal(ctx, id)
	if err != nil {
		return DeleteResult{}, errs.Internal(err)
	}
	if hasDeal {
		if err := uc.propertyRepo.UpdateStatus(ctx, id, entity.PropertyStatusArchived); err != nil {
			return DeleteResult{}, errs.Internal(err)
		}
		uc.auditService.Log(&actor.ID, "archived_property", "properties", &id, nil, meta.IP, meta.UserAgent)
		return DeleteResult{Archived: true}, nil
	}
	if err := uc.propertyRepo.Delete(ctx, id); err != nil {
		return DeleteResult{}, errs.Internal(err)
	}
	uc.auditService.Log(&actor.ID, "deleted_property", "properties", &id, nil, meta.IP, meta.UserAgent)
	return DeleteResult{}, nil
}

// GetByID returns the projection that matches the actor's access level.
func (uc *PropertyUsecase) GetByID(ctx context.Context, actor Actor, id uuid.UUID) (any, error) {
	p, lvl, err := uc.load(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if lvl == AccessPublic {
		return toPublicProperty(p), nil
	}
	team, err := uc.coRepo.ListByProperty(ctx, id)
	if err != nil {
		return nil, errs.Internal(err)
	}
	if lvl == AccessPrimary {
		return toPrimaryProperty(p, team), nil
	}
	if lvl == AccessCoListing {
		// Same view as the primary marketing (including exact address); owner data is never included.
		v := toPrimaryProperty(p, team)
		v.AccessLevel = AccessCoListing
		return v, nil
	}
	return toInternalProperty(p, lvl, team), nil
}

// ListPropertiesInput holds list/search parameters (already parsed by the handler).
type ListPropertiesInput struct {
	Search          string
	Type            string
	TransactionType string
	Status          string // internal roles only
	Scope           string // "", "mine", "co_listing" (marketing only)
	History         bool   // only the historical (sold/rented) section
	IncludeHistory  bool   // available + historical in one list; items carry is_available
	FacilityIDs     []uuid.UUID
	ProvinceID      *uuid.UUID
	CityID          *uuid.UUID
	DistrictID      *uuid.UUID
	MinPrice        *float64
	MaxPrice        *float64
	MinLandArea     *float64
	MaxLandArea     *float64
	MinBuildingArea *float64
	MaxBuildingArea *float64
	Bedrooms        *int
	Bathrooms       *int
	Sort            string
	Page            int
	Limit           int
}

var allowedSorts = map[string]bool{
	"": true, "newest": true, "oldest": true,
	"price_asc": true, "price_desc": true, "name_asc": true, "name_desc": true,
}

func rangeOK(min, max *float64) bool {
	return min == nil || max == nil || *min <= *max
}

// List returns a page of listings. Availability semantics are enforced here
// regardless of client input: non-internal callers never receive draft,
// maintenance or archived listings, and sold/rented only via History.
func (uc *PropertyUsecase) List(ctx context.Context, actor Actor, in ListPropertiesInput) ([]PublicProperty, int64, error) {
	if !allowedSorts[in.Sort] {
		return nil, 0, errs.Validation("invalid sort value")
	}
	if !rangeOK(in.MinPrice, in.MaxPrice) || !rangeOK(in.MinLandArea, in.MaxLandArea) || !rangeOK(in.MinBuildingArea, in.MaxBuildingArea) {
		return nil, 0, errs.Validation("minimum value must not exceed maximum value")
	}
	if in.TransactionType != "" && in.TransactionType != entity.TransactionTypeSale && in.TransactionType != entity.TransactionTypeRent {
		return nil, 0, errs.Validation("invalid transaction_type")
	}
	if in.Type != "" {
		valid := false
		for _, t := range entity.ValidPropertyTypes() {
			if t == in.Type {
				valid = true
			}
		}
		if !valid {
			return nil, 0, errs.Validation("invalid type")
		}
	}
	if len([]rune(in.Search)) > MaxSearchLength {
		return nil, 0, errs.Validation("search must be at most 100 characters")
	}
	if len(in.FacilityIDs) > MaxFacilityIDs {
		return nil, 0, errs.Validation("at most 20 facility_ids")
	}
	if in.Page < 1 {
		in.Page = 1
	}
	if in.Limit < 1 || in.Limit > MaxPageSize {
		in.Limit = 20
	}
	if (in.Page-1)*in.Limit > MaxListOffset {
		return nil, 0, errs.Validation("page is too deep; narrow the search filters")
	}

	f := repository.PropertyFilter{
		Search:          strings.TrimSpace(in.Search),
		Type:            in.Type,
		TransactionType: in.TransactionType,
		ProvinceID:      in.ProvinceID,
		CityID:          in.CityID,
		DistrictID:      in.DistrictID,
		MinPrice:        in.MinPrice,
		MaxPrice:        in.MaxPrice,
		MinLandArea:     in.MinLandArea,
		MaxLandArea:     in.MaxLandArea,
		MinBuildingArea: in.MinBuildingArea,
		MaxBuildingArea: in.MaxBuildingArea,
		Bedrooms:        in.Bedrooms,
		Bathrooms:       in.Bathrooms,
		Sort:            in.Sort,
		Page:            in.Page,
		Limit:           in.Limit,
		Visibility:      repository.VisibilityAvailable,
		FacilityIDs:     in.FacilityIDs,
	}

	switch {
	case actor.IsAdmin():
		f.Visibility = repository.VisibilityAll
		f.Status = in.Status
	case actor.IsSales() && in.Scope == "mine":
		f.Visibility = repository.VisibilityAll
		f.PrimaryOf = &actor.ID
		f.Status = in.Status
	case actor.IsSales() && in.Scope == "co_listing":
		f.Visibility = repository.VisibilityAll
		f.CoListingOf = &actor.ID
		f.Status = in.Status
	case in.History:
		f.Visibility = repository.VisibilityHistory
	case in.IncludeHistory:
		f.Visibility = repository.VisibilityAvailableAndHistory
	}

	items, total, err := uc.propertyRepo.List(ctx, f)
	if err != nil {
		return nil, 0, errs.Internal(err)
	}
	out := make([]PublicProperty, 0, len(items))
	for i := range items {
		out = append(out, toPublicProperty(&items[i]))
	}
	return out, total, nil
}

// Statuses a marketing owner may set manually. sold/rented come only from deal finalization.
var manualStatuses = map[string]bool{
	entity.PropertyStatusDraft:       true,
	entity.PropertyStatusAvailable:   true,
	entity.PropertyStatusReserved:    true,
	entity.PropertyStatusMaintenance: true,
	entity.PropertyStatusArchived:    true,
}

// coListingStatuses are the statuses a co-listing marketing user may set.
var coListingStatuses = map[string]bool{
	entity.PropertyStatusAvailable:   true,
	entity.PropertyStatusReserved:    true,
	entity.PropertyStatusMaintenance: true,
}

// UpdateStatus changes the listing status. The primary marketing owner may set
// any manual status; a co-listing marketing user may only toggle available /
// reserved / maintenance (never draft, archived, sold or rented), and cannot
// change a draft or archived listing.
func (uc *PropertyUsecase) UpdateStatus(ctx context.Context, actor Actor, id uuid.UUID, status string, meta RequestMeta) error {
	if !actor.Authenticated() {
		return errs.ErrUnauthorized
	}
	p, lvl, err := uc.load(ctx, actor, id)
	if err != nil {
		return err
	}
	if lvl != AccessPrimary && lvl != AccessCoListing {
		return errs.Forbidden("Only the primary or a co-listing marketing user can change the status")
	}
	if !manualStatuses[status] {
		return errs.Validation("status cannot be set manually; sold and rented are set by deal finalization")
	}
	if lvl == AccessCoListing {
		if !coListingStatuses[status] {
			return errs.Forbidden("Co-listing marketing can only set available, reserved or maintenance")
		}
		if p.Status == entity.PropertyStatusDraft || p.Status == entity.PropertyStatusArchived {
			return errs.Forbidden("Only the primary marketing owner can change a draft or archived listing")
		}
	}
	if (p.Status == entity.PropertyStatusSold || p.Status == entity.PropertyStatusRented) && status != entity.PropertyStatusArchived {
		return errs.Validation("a sold or rented listing can only be archived")
	}
	if p.Status == status {
		return nil
	}
	if err := uc.propertyRepo.UpdateStatus(ctx, id, status); err != nil {
		return errs.Internal(err)
	}
	uc.auditService.Log(&actor.ID, "updated_property_status", "properties", &id,
		map[string]string{"old_status": p.Status, "new_status": status, "access": lvl}, meta.IP, meta.UserAgent)
	return nil
}

// UploadImage uploads an image. Primary marketing owner only.
func (uc *PropertyUsecase) UploadImage(ctx context.Context, actor Actor, propertyID uuid.UUID, file multipart.File, header *multipart.FileHeader) (*entity.PropertyImage, error) {
	p, err := uc.requirePrimary(ctx, actor, propertyID)
	if err != nil {
		return nil, err
	}
	path, err := uc.storage.Upload(ctx, file, header)
	if err != nil {
		return nil, errs.BadRequest(err.Error())
	}
	image := &entity.PropertyImage{PropertyID: propertyID, URL: uc.storage.GetURL(path)}
	err = uc.tx.Do(ctx, func(ctx context.Context) error {
		if err := uc.imageRepo.Create(ctx, image); err != nil {
			return err
		}
		if p.Thumbnail != "" {
			return nil
		}
		// The first photo becomes the cover until the marketing picks another.
		image.IsThumbnail = true
		return uc.imageRepo.SetThumbnail(ctx, propertyID, image.ID)
	})
	if err != nil {
		_ = uc.storage.Delete(ctx, image.URL)
		return nil, errs.Internal(err)
	}
	return image, nil
}

// SetThumbnail makes one of the property's images its cover. Primary marketing owner only.
func (uc *PropertyUsecase) SetThumbnail(ctx context.Context, actor Actor, propertyID, imageID uuid.UUID) error {
	if _, err := uc.requirePrimary(ctx, actor, propertyID); err != nil {
		return err
	}
	image, err := uc.imageRepo.FindByID(ctx, imageID)
	if err != nil {
		return err
	}
	if image.PropertyID != propertyID {
		return errs.NotFound("Property image")
	}
	if err := uc.tx.Do(ctx, func(ctx context.Context) error {
		return uc.imageRepo.SetThumbnail(ctx, propertyID, imageID)
	}); err != nil {
		return errs.Internal(err)
	}
	return nil
}

// DeleteImage deletes an image. Primary marketing owner only.
func (uc *PropertyUsecase) DeleteImage(ctx context.Context, actor Actor, propertyID, imageID uuid.UUID) error {
	if _, err := uc.requirePrimary(ctx, actor, propertyID); err != nil {
		return err
	}
	image, err := uc.imageRepo.FindByID(ctx, imageID)
	if err != nil {
		return err
	}
	if image.PropertyID != propertyID {
		return errs.NotFound("Property image")
	}
	err = uc.tx.Do(ctx, func(ctx context.Context) error {
		if err := uc.imageRepo.Delete(ctx, imageID); err != nil {
			return err
		}
		if !image.IsThumbnail {
			return nil
		}
		// Promote the next photo so the listing keeps a cover.
		rest, err := uc.imageRepo.ListByPropertyID(ctx, propertyID)
		if err != nil {
			return err
		}
		next := uuid.Nil
		if len(rest) > 0 {
			next = rest[0].ID
		}
		return uc.imageRepo.SetThumbnail(ctx, propertyID, next)
	})
	if err != nil {
		return errs.Internal(err)
	}
	_ = uc.storage.Delete(ctx, image.URL)
	return nil
}
