package usecase

import (
	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
)

// Actor is the authenticated caller, built only from a verified token.
// The zero value is an anonymous (unauthenticated) caller.
type Actor struct {
	ID   uuid.UUID
	Role string
}

// Authenticated reports whether the actor carries a verified identity.
func (a Actor) Authenticated() bool { return a.ID != uuid.Nil && a.Role != "" }

// IsSales reports whether the actor is a marketing (sales) user.
func (a Actor) IsSales() bool { return a.Authenticated() && a.Role == entity.RoleSales }

// IsAdmin reports whether the actor is an admin.
func (a Actor) IsAdmin() bool { return a.Authenticated() && a.Role == entity.RoleAdmin }

// RequestMeta carries request details for audit logging.
type RequestMeta struct {
	IP        string
	UserAgent string
}
