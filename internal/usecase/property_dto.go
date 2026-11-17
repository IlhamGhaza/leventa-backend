package usecase

import (
	"time"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
)

// Access levels reported in internal projections.
const (
	AccessPrimary   = "primary"
	AccessCoListing = "co_listing"
	AccessAdmin     = "admin"
	AccessPublic    = "public"
)

// LocationRef is a province/city reference.
type LocationRef struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// MarketingContact is the representing marketing user's contact channel.
// It never contains owner data.
type MarketingContact struct {
	Name           string `json:"name"`
	WhatsAppNumber string `json:"whatsapp_number,omitempty"`
	WhatsAppURL    string `json:"whatsapp_url,omitempty"`
}

// PublicProperty is the projection for guests (no account) and
// marketing users without a relationship to the listing.
// It contains no owner data, no restricted address and no coordinates.
type PublicProperty struct {
	ID              uuid.UUID              `json:"id"`
	Code            string                 `json:"code"`
	Name            string                 `json:"name"`
	Type            string                 `json:"type"`
	TransactionType string                 `json:"transaction_type"`
	Status          string                 `json:"status"`
	StatusBadge     string                 `json:"status_badge,omitempty"`
	IsAvailable     bool                   `json:"is_available"`
	Price           *float64               `json:"price,omitempty"`
	RentalPrice     *float64               `json:"rental_price,omitempty"`
	Description     string                 `json:"description,omitempty"`
	Province        *LocationRef           `json:"province,omitempty"`
	City            *LocationRef           `json:"city,omitempty"`
	District        *LocationRef           `json:"district,omitempty"`
	Facilities      []LocationRef          `json:"facilities,omitempty"`
	GeneralLocation string                 `json:"general_location,omitempty"`
	LandArea        *float64               `json:"land_area,omitempty"`
	BuildingArea    *float64               `json:"building_area,omitempty"`
	Floor           *int                   `json:"floor,omitempty"`
	Bedrooms        int                    `json:"bedrooms"`
	Bathrooms       int                    `json:"bathrooms"`
	YearBuilt       *int                   `json:"year_built,omitempty"`
	Thumbnail       string                 `json:"thumbnail,omitempty"`
	Images          []entity.PropertyImage `json:"images,omitempty"`
	Contact         *MarketingContact      `json:"contact,omitempty"`
	PropertyDetailsView
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TeamMember is a marketing user assigned to a listing.
type TeamMember struct {
	MarketingID    uuid.UUID `json:"marketing_id"`
	Name           string    `json:"name"`
	AssignmentType string    `json:"assignment_type"`
	AssignedAt     time.Time `json:"assigned_at"`
}

// RestrictedAddress is the exact address, visible to the primary marketing only.
type RestrictedAddress struct {
	BlockNumber string `json:"block_number,omitempty"`
	ClusterName string `json:"cluster_name,omitempty"`
	StreetName  string `json:"street_name,omitempty"`
	HouseNumber string `json:"house_number,omitempty"`
	FloorNumber string `json:"floor_number,omitempty"`
	RoomNumber  string `json:"room_number,omitempty"`
}

// PrimaryProperty is the projection for the primary marketing owner.
// Owner PII is still served only by the dedicated owner endpoint.
type PrimaryProperty struct {
	PublicProperty
	AccessLevel string            `json:"access_level"`
	Address     RestrictedAddress `json:"address"`
	Latitude    *float64          `json:"latitude,omitempty"`
	Longitude   *float64          `json:"longitude,omitempty"`
	Team        []TeamMember      `json:"marketing_team"`
}

// InternalProperty is the projection for admin.
// No restricted address, coordinates or owner data.
type InternalProperty struct {
	PublicProperty
	AccessLevel string       `json:"access_level"`
	Team        []TeamMember `json:"marketing_team"`
}

func toPublicProperty(p *entity.Property) PublicProperty {
	v := PublicProperty{
		ID:                  p.ID,
		Code:                p.Code,
		Name:                p.Name,
		Type:                p.Type,
		TransactionType:     p.TransactionType,
		Status:              p.Status,
		Price:               p.Price,
		RentalPrice:         p.RentalPrice,
		Description:         p.Description,
		GeneralLocation:     p.GeneralLocation,
		LandArea:            p.LandArea,
		BuildingArea:        p.BuildingArea,
		Floor:               p.Floor,
		Bedrooms:            p.Bedrooms,
		Bathrooms:           p.Bathrooms,
		YearBuilt:           p.YearBuilt,
		Thumbnail:           p.Thumbnail,
		Images:              p.Images,
		PropertyDetailsView: toDetailsView(p),
		CreatedAt:           p.CreatedAt,
		UpdatedAt:           p.UpdatedAt,
	}
	v.IsAvailable = p.Status == entity.PropertyStatusAvailable || p.Status == entity.PropertyStatusReserved
	switch p.Status {
	case entity.PropertyStatusSold:
		v.StatusBadge = "SOLD"
	case entity.PropertyStatusRented:
		v.StatusBadge = "RENTED"
	}
	if p.Province != nil {
		v.Province = &LocationRef{ID: p.Province.ID, Name: p.Province.Name}
	}
	if p.City != nil {
		v.City = &LocationRef{ID: p.City.ID, Name: p.City.Name}
	}
	if p.District != nil {
		v.District = &LocationRef{ID: p.District.ID, Name: p.District.Name}
	}
	for _, f := range p.Facilities {
		v.Facilities = append(v.Facilities, LocationRef{ID: f.ID, Name: f.Name})
	}
	// The representing marketing user's contact is public (never the owner's).
	if p.PrimaryMarketing != nil {
		v.Contact = buildContact(p)
	}
	return v
}

func buildContact(p *entity.Property) *MarketingContact {
	m := p.PrimaryMarketing
	c := &MarketingContact{Name: m.Name}
	number, ok := NormalizeWhatsApp(m.WhatsAppNumber)
	if !ok {
		return c
	}
	c.WhatsAppNumber = number
	loc := p.GeneralLocation
	if loc == "" && p.City != nil {
		loc = p.City.Name
		if p.Province != nil {
			loc += ", " + p.Province.Name
		}
	}
	price := p.Price
	if p.TransactionType == entity.TransactionTypeRent {
		price = p.RentalPrice
	}
	c.WhatsAppURL = BuildWhatsAppURL(number, p.Name, p.Code, price, loc)
	return c
}

func toTeam(list []entity.PropertyMarketingAssignment) []TeamMember {
	out := make([]TeamMember, 0, len(list))
	for _, a := range list {
		m := TeamMember{MarketingID: a.MarketingID, AssignmentType: a.AssignmentType, AssignedAt: a.AssignedAt}
		if a.Marketing != nil {
			m.Name = a.Marketing.Name
		}
		out = append(out, m)
	}
	return out
}

func toPrimaryProperty(p *entity.Property, team []entity.PropertyMarketingAssignment) PrimaryProperty {
	return PrimaryProperty{
		PublicProperty: toPublicProperty(p),
		AccessLevel:    AccessPrimary,
		Address: RestrictedAddress{
			BlockNumber: p.BlockNumber, ClusterName: p.ClusterName, StreetName: p.StreetName,
			HouseNumber: p.HouseNumber, FloorNumber: p.FloorNumber, RoomNumber: p.RoomNumber,
		},
		Latitude:  p.Latitude,
		Longitude: p.Longitude,
		Team:      toTeam(team),
	}
}

func toInternalProperty(p *entity.Property, level string, team []entity.PropertyMarketingAssignment) InternalProperty {
	return InternalProperty{
		PublicProperty: toPublicProperty(p),
		AccessLevel:    level,
		Team:           toTeam(team),
	}
}

// OwnerView is owner PII. Returned only to the primary marketing of the listing.
type OwnerView struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email,omitempty"`
	Phone          string    `json:"phone,omitempty"`
	PrivateAddress string    `json:"private_address,omitempty"`
}

func toOwnerView(o *entity.Owner) OwnerView {
	return OwnerView{ID: o.ID, Name: o.Name, Email: o.Email, Phone: o.Phone, PrivateAddress: o.PrivateAddress}
}
