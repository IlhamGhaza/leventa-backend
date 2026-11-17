package entity

import (
	"time"

	"github.com/google/uuid"
)

// PropertyImage represents an image associated with a property.
type PropertyImage struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PropertyID  uuid.UUID `gorm:"type:uuid;not null;index" json:"property_id"`
	URL         string    `gorm:"size:500;not null" json:"url"`
	SortOrder   int       `gorm:"not null;default:0" json:"sort_order"`
	IsThumbnail bool      `gorm:"not null;default:false" json:"is_thumbnail"`
	CreatedAt   time.Time `json:"created_at"`
}

// TableName returns the table name for PropertyImage.
func (PropertyImage) TableName() string {
	return "property_images"
}
