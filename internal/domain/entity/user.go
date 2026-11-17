package entity

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User roles
const (
	RoleSales = "sales"
	RoleAdmin = "admin"
)

// User statuses
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
)

// User represents a system user.
type User struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name                string         `gorm:"size:255;not null" json:"name"`
	Email               string         `gorm:"size:255;not null;uniqueIndex" json:"email"`
	Phone               string         `gorm:"size:50" json:"phone,omitempty"`
	WhatsAppNumber      string         `gorm:"column:whatsapp_number;size:20" json:"whatsapp_number,omitempty"`
	PasswordHash        string         `gorm:"size:255;not null" json:"-"`
	Role                string         `gorm:"size:20;not null;default:'sales'" json:"role"`
	Avatar              string         `gorm:"size:500" json:"avatar,omitempty"`
	Status              string         `gorm:"size:20;not null;default:'active'" json:"status"`
	FailedLoginAttempts int            `gorm:"not null;default:0" json:"-"`
	LockedUntil         *time.Time     `json:"-"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for User.
func (User) TableName() string {
	return "users"
}

// SetPassword hashes and sets the user's password.
func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

// CheckPassword verifies the password against the hash.
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// IsAdmin checks if the user has admin role.
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsSales checks if the user has sales role.
func (u *User) IsSales() bool {
	return u.Role == RoleSales
}

// IsActive checks if the user is active.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// ValidRoles returns all valid user roles.
func ValidRoles() []string {
	return []string{RoleSales, RoleAdmin}
}
