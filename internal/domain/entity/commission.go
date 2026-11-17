package entity

import (
	"math"
	"time"

	"github.com/google/uuid"
)

// Commission statuses
const (
	CommissionStatusPending   = "pending"
	CommissionStatusPaid      = "paid"
	CommissionStatusCancelled = "cancelled"
)

// DefaultCommissionPercentage is the default commission rate (2.5%).
const DefaultCommissionPercentage = 2.5

// Commission represents a sales commission from a deal.
type Commission struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DealID     uuid.UUID `gorm:"type:uuid;not null;index" json:"deal_id"`
	SalesID    uuid.UUID `gorm:"type:uuid;not null;index" json:"sales_id"`
	Percentage float64   `gorm:"type:decimal(5,2);not null" json:"percentage"`
	Amount     float64   `gorm:"type:decimal(15,2);not null" json:"amount"`
	Status     string    `gorm:"size:20;not null;default:'pending'" json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relations
	Deal  *Deal `gorm:"foreignKey:DealID" json:"deal,omitempty"`
	Sales *User `gorm:"foreignKey:SalesID" json:"sales,omitempty"`
}

// TableName returns the table name for Commission.
func (Commission) TableName() string {
	return "commissions"
}

// CalculateCommission calculates the commission amount from a deal price.
func CalculateCommission(dealPrice float64, percentage float64) float64 {
	return math.Round(dealPrice*percentage) / 100
}
