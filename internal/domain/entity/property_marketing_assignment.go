package entity

import (
	"time"

	"github.com/google/uuid"
)

// Assignment types
const (
	AssignmentTypePrimary   = "primary"
	AssignmentTypeCoListing = "co_listing"
)

// Marketing team limits per property.
const (
	MaxCoListingPerProperty = 2
	MaxMarketingPerProperty = 3
)

// PropertyMarketingAssignment links a marketing user to a property.
type PropertyMarketingAssignment struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PropertyID     uuid.UUID `gorm:"type:uuid;not null;index"`
	MarketingID    uuid.UUID `gorm:"type:uuid;not null;index"`
	AssignmentType string    `gorm:"size:20;not null"`
	AssignedBy     uuid.UUID `gorm:"type:uuid;not null"`
	AssignedAt     time.Time `gorm:"not null;default:now()"`

	Marketing *User `gorm:"foreignKey:MarketingID"`
}

// TableName returns the table name.
func (PropertyMarketingAssignment) TableName() string {
	return "property_marketing_assignments"
}
