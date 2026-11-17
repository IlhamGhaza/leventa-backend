package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/pkg/errs"
)

// propertyAccess resolves what an actor may do with a property.
// Authorization is decided here (use case layer), not in route middleware.
type propertyAccess struct {
	propertyRepo repository.PropertyRepository
	coRepo       repository.CoListingRepository
}

// publicStatus reports whether guests may see a listing with this status.
func publicStatus(status string) bool {
	switch status {
	case entity.PropertyStatusAvailable, entity.PropertyStatusReserved,
		entity.PropertyStatusSold, entity.PropertyStatusRented:
		return true
	}
	return false
}

// level returns the actor's access level for p.
func (a propertyAccess) level(ctx context.Context, actor Actor, p *entity.Property) (string, error) {
	switch {
	case actor.IsSales():
		if p.PrimaryMarketingID == actor.ID {
			return AccessPrimary, nil
		}
		asg, err := a.coRepo.Find(ctx, p.ID, actor.ID)
		if err != nil {
			return "", errs.Internal(err)
		}
		if asg != nil && asg.AssignmentType == entity.AssignmentTypeCoListing {
			return AccessCoListing, nil
		}
	case actor.IsAdmin():
		return AccessAdmin, nil
	}
	return AccessPublic, nil
}

// load returns the property and the actor's access level. A listing that is
// not public (draft/maintenance/archived) is reported as not found to actors
// with no relationship to it, so its existence does not leak.
func (a propertyAccess) load(ctx context.Context, actor Actor, id uuid.UUID) (*entity.Property, string, error) {
	p, err := a.propertyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	lvl, err := a.level(ctx, actor, p)
	if err != nil {
		return nil, "", err
	}
	if lvl == AccessPublic && !publicStatus(p.Status) {
		return nil, "", errs.NotFound("Property")
	}
	return p, lvl, nil
}

// requirePrimary returns the property only when the actor is its primary
// marketing owner. Otherwise: 401 when anonymous, 404 when the listing is
// hidden from the actor, 403 when visible but not editable.
func (a propertyAccess) requirePrimary(ctx context.Context, actor Actor, id uuid.UUID) (*entity.Property, error) {
	if !actor.Authenticated() {
		return nil, errs.ErrUnauthorized
	}
	p, lvl, err := a.load(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if lvl != AccessPrimary {
		return nil, errs.Forbidden("Only the primary marketing owner can perform this action")
	}
	return p, nil
}
