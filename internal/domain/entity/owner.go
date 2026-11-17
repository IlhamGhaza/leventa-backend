package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Owner holds private owner PII. Every personal field is tagged json:"-";
// responses must use an explicit DTO returned only to the primary marketing user.
type Owner struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"-"`
	Name           string         `gorm:"size:255;not null" json:"-"`
	Email          string         `gorm:"size:255" json:"-"`
	Phone          string         `gorm:"size:50" json:"-"`
	PrivateAddress string         `gorm:"type:text" json:"-"`
	CreatedBy      *uuid.UUID     `gorm:"type:uuid;index" json:"-"`
	CreatedAt      time.Time      `json:"-"`
	UpdatedAt      time.Time      `json:"-"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for Owner.
func (Owner) TableName() string {
	return "owners"
}
