package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// AuditLog represents a system audit log entry.
type AuditLog struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    *uuid.UUID     `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action    string         `gorm:"size:100;not null" json:"action"`
	Module    string         `gorm:"size:100;not null" json:"module"`
	EntityID  *uuid.UUID     `gorm:"type:uuid" json:"entity_id,omitempty"`
	Metadata  datatypes.JSON `gorm:"type:jsonb" json:"metadata,omitempty"`
	IPAddress string         `gorm:"size:50" json:"ip_address,omitempty"`
	UserAgent string         `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt time.Time      `json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName returns the table name for AuditLog.
func (AuditLog) TableName() string {
	return "audit_logs"
}
