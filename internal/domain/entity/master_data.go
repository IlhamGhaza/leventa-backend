package entity

import (
	"time"

	"github.com/google/uuid"
)

// Province represents a province/state.
type Province struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"size:255;not null;uniqueIndex" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Cities []City `gorm:"foreignKey:ProvinceID" json:"cities,omitempty"`
}

// TableName returns the table name for Province.
func (Province) TableName() string {
	return "provinces"
}

// City represents a city within a province.
type City struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ProvinceID uuid.UUID `gorm:"type:uuid;not null;index" json:"province_id"`
	Name       string    `gorm:"size:255;not null" json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relations
	Province *Province `gorm:"foreignKey:ProvinceID" json:"province,omitempty"`
}

// TableName returns the table name for City.
func (City) TableName() string {
	return "cities"
}

// District represents a district (kecamatan) within a city.
type District struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CityID    uuid.UUID `gorm:"type:uuid;not null;index" json:"city_id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	City *City `gorm:"foreignKey:CityID" json:"city,omitempty"`
}

// TableName returns the table name for District.
func (District) TableName() string { return "districts" }

// Facility is a database-backed property facility (admin-managed master data).
type Facility struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"size:100;not null;uniqueIndex" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for Facility.
func (Facility) TableName() string { return "facilities" }
