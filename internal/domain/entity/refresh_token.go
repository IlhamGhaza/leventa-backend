package entity

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a stored refresh token.
type RefreshToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string     `gorm:"size:255;not null;index" json:"-"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// TableName returns the table name for RefreshToken.
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// IsExpired checks if the token has expired.
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsRevoked checks if the token has been revoked.
func (rt *RefreshToken) IsRevoked() bool {
	return rt.RevokedAt != nil
}
