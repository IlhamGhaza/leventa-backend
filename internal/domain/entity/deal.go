package entity

import (
	"time"

	"github.com/google/uuid"
)

// Deal types
const (
	DealTypeSale = "sale"
	DealTypeRent = "rent"
)

// Deal outcomes recorded by Admin (D-027).
const (
	DealOutcomeDeal      = "deal"
	DealOutcomeCancelled = "cancelled"
)

// Deal is the outcome Admin records for a property: a finalized transaction
// or a cancellation (the owner no longer sells/rents). Type, Price and
// BuyerName are set only for outcome "deal".
type Deal struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PropertyID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"property_id"`
	SalesID    uuid.UUID `gorm:"type:uuid;not null;index" json:"sales_id"` // primary marketing at record time
	RecordedBy uuid.UUID `gorm:"type:uuid;not null" json:"recorded_by"`    // Admin
	Outcome    string    `gorm:"size:20;not null" json:"outcome"`
	Type       *string   `gorm:"size:20" json:"type,omitempty"`
	Price      *float64  `gorm:"type:decimal(15,2)" json:"price,omitempty"`
	BuyerName  *string   `gorm:"size:255" json:"buyer_name,omitempty"`
	DealDate   time.Time `gorm:"type:date;not null" json:"deal_date"`
	Notes      string    `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`

	// Relations
	Property *Property `gorm:"foreignKey:PropertyID" json:"property,omitempty"`
	Sales    *User     `gorm:"foreignKey:SalesID" json:"sales,omitempty"`
}

// TableName returns the table name for Deal.
func (Deal) TableName() string {
	return "deals"
}
