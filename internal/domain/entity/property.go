package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Property types
const (
	PropertyTypeHouse     = "house"
	PropertyTypeApartment = "apartment"
	PropertyTypeShopHouse = "shop_house" // ruko
	PropertyTypeFactory   = "factory"    // pabrik
)

// Property statuses
const (
	PropertyStatusDraft       = "draft"
	PropertyStatusAvailable   = "available"
	PropertyStatusReserved    = "reserved"
	PropertyStatusSold        = "sold"
	PropertyStatusRented      = "rented"
	PropertyStatusMaintenance = "maintenance"
	PropertyStatusArchived    = "archived"
)

// Transaction types
const (
	TransactionTypeSale = "sale"
	TransactionTypeRent = "rent"
)

// Property represents a real estate property.
//
// Restricted fields (exact address, coordinates, owner, primary marketing)
// are tagged json:"-" so that an accidental serialization of the entity can
// never leak them. Responses must use the DTO projections in the usecase package.
type Property struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerID            uuid.UUID  `gorm:"type:uuid;not null" json:"-"`
	PrimaryMarketingID uuid.UUID  `gorm:"type:uuid;not null;index" json:"-"`
	Code               string     `gorm:"size:50;not null;uniqueIndex" json:"code"`
	Name               string     `gorm:"size:255;not null" json:"name"`
	Type               string     `gorm:"size:50;not null" json:"type"`
	TransactionType    string     `gorm:"size:10;not null;default:'sale'" json:"transaction_type"`
	Status             string     `gorm:"size:50;not null;default:'available'" json:"status"`
	Price              *float64   `gorm:"type:decimal(15,2)" json:"price,omitempty"`
	RentalPrice        *float64   `gorm:"type:decimal(15,2)" json:"rental_price,omitempty"`
	Description        string     `gorm:"type:text" json:"description,omitempty"`
	ProvinceID         *uuid.UUID `gorm:"type:uuid" json:"province_id,omitempty"`
	CityID             *uuid.UUID `gorm:"type:uuid" json:"city_id,omitempty"`
	DistrictID         *uuid.UUID `gorm:"type:uuid" json:"district_id,omitempty"`
	GeneralLocation    string     `gorm:"size:255" json:"general_location,omitempty"`

	// Restricted address (primary marketing only).
	BlockNumber string   `gorm:"size:50" json:"-"`
	ClusterName string   `gorm:"size:255" json:"-"`
	StreetName  string   `gorm:"size:255" json:"-"`
	HouseNumber string   `gorm:"size:50" json:"-"`
	FloorNumber string   `gorm:"size:50" json:"-"`
	RoomNumber  string   `gorm:"size:50" json:"-"`
	Latitude    *float64 `gorm:"type:decimal(10,8)" json:"-"`
	Longitude   *float64 `gorm:"type:decimal(11,8)" json:"-"`

	LandArea     *float64 `gorm:"type:decimal(10,2)" json:"land_area,omitempty"`
	BuildingArea *float64 `gorm:"type:decimal(10,2)" json:"building_area,omitempty"`
	Floor        *int     `json:"floor,omitempty"`
	Bedrooms     int      `gorm:"default:0" json:"bedrooms"`
	Bathrooms    int      `gorm:"default:0" json:"bathrooms"`
	YearBuilt    *int     `json:"year_built,omitempty"`

	// Listing details (D-028). Public except where noted in the DTO projections.
	ShortDescription string   `gorm:"size:500" json:"short_description,omitempty"`
	AreaLabel        string   `gorm:"size:255" json:"area_label,omitempty"`
	SubDistrict      string   `gorm:"size:255" json:"sub_district,omitempty"`
	PostalCode       string   `gorm:"size:10" json:"postal_code,omitempty"`
	Orientation      string   `gorm:"size:50" json:"orientation,omitempty"`
	MaintenanceFee   *float64 `gorm:"type:decimal(15,2)" json:"maintenance_fee,omitempty"`

	CertSHM    bool `gorm:"column:cert_shm" json:"cert_shm"`
	CertHGB    bool `gorm:"column:cert_hgb" json:"cert_hgb"`
	CertIMB    bool `gorm:"column:cert_imb" json:"cert_imb"`
	CertPBB    bool `gorm:"column:cert_pbb" json:"cert_pbb"`
	CertGirik  bool `gorm:"column:cert_girik" json:"cert_girik"`
	CertOther  bool `gorm:"column:cert_other" json:"cert_other"`
	CertSHMSRS bool `gorm:"column:cert_shmsrs" json:"cert_shmsrs"`
	CertHGU    bool `gorm:"column:cert_hgu" json:"cert_hgu"`
	CertSLF    bool `gorm:"column:cert_slf" json:"cert_slf"`

	LivingRooms     int `json:"living_rooms"`
	DiningRooms     int `json:"dining_rooms"`
	MasterBedrooms  int `json:"master_bedrooms"`
	ChildBedrooms   int `json:"child_bedrooms"`
	OtherBedrooms   int `json:"other_bedrooms"`
	MasterBathrooms int `json:"master_bathrooms"`
	ChildBathrooms  int `json:"child_bathrooms"`
	OtherBathrooms  int `json:"other_bathrooms"`
	FamilyRooms     int `json:"family_rooms"`
	StudyRooms      int `json:"study_rooms"`
	Kitchens        int `json:"kitchens"`

	ElectricityCapacity string `gorm:"size:50" json:"electricity_capacity,omitempty"`
	AirConditioners     int    `json:"air_conditioners"`
	TVProvider          string `gorm:"column:tv_provider;size:100" json:"tv_provider,omitempty"`
	WaterSource         string `gorm:"size:100" json:"water_source,omitempty"`
	FurnishingStatus    string `gorm:"size:30" json:"furnishing_status,omitempty"`

	HasPool      bool `json:"has_pool"`
	HasGarage    bool `json:"has_garage"`
	HasFrontYard bool `json:"has_front_yard"`
	HasBackYard  bool `json:"has_back_yard"`
	HasBalcony   bool `json:"has_balcony"`

	DistSchool      *int `json:"dist_school,omitempty"`
	DistBusTerminal *int `json:"dist_bus_terminal,omitempty"`
	DistMrtLrt      *int `json:"dist_mrt_lrt,omitempty"`
	DistToll        *int `json:"dist_toll,omitempty"`
	DistHospital    *int `json:"dist_hospital,omitempty"`

	// Free-form additions entered by the primary marketing (not filterable).
	ExtraOutdoor   []string      `gorm:"serializer:json;type:jsonb;not null;default:'[]'" json:"extra_outdoor"`
	ExtraDistances []NearbyPlace `gorm:"serializer:json;type:jsonb;not null;default:'[]'" json:"extra_distances"`

	Thumbnail string         `gorm:"size:500" json:"thumbnail,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Owner            *Owner          `gorm:"foreignKey:OwnerID" json:"-"`
	PrimaryMarketing *User           `gorm:"foreignKey:PrimaryMarketingID" json:"-"`
	Images           []PropertyImage `gorm:"foreignKey:PropertyID" json:"images,omitempty"`
	Province         *Province       `gorm:"foreignKey:ProvinceID" json:"province,omitempty"`
	City             *City           `gorm:"foreignKey:CityID" json:"city,omitempty"`
	District         *District       `gorm:"foreignKey:DistrictID" json:"district,omitempty"`
	Facilities       []Facility      `gorm:"many2many:property_facilities" json:"facilities,omitempty"`
}

// TableName returns the table name for Property.
func (Property) TableName() string {
	return "properties"
}

// ValidPropertyTypes returns all valid property types.
func ValidPropertyTypes() []string {
	return []string{
		PropertyTypeHouse, PropertyTypeApartment, PropertyTypeShopHouse, PropertyTypeFactory,
	}
}

// ValidPropertyStatuses returns all valid property statuses.
func ValidPropertyStatuses() []string {
	return []string{
		PropertyStatusDraft, PropertyStatusAvailable, PropertyStatusReserved,
		PropertyStatusSold, PropertyStatusRented, PropertyStatusMaintenance,
		PropertyStatusArchived,
	}
}

// Furnishing statuses
const (
	FurnishingUnfurnished   = "unfurnished"
	FurnishingSemiFurnished = "semi_furnished"
	FurnishingFurnished     = "furnished"
)

// NearbyPlace is a custom public place with its distance from the property.
type NearbyPlace struct {
	Name   string `json:"name" validate:"required,max=50"`
	Meters int    `json:"meters" validate:"gte=0,lte=1000000"`
}
