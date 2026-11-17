package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/internal/service"
	"github.com/leventa-backend/pkg/errs"
)

// CoListingUsecase manages co-listing membership. Admin only. It never touches
// listing content, owner data, or the primary owner.
type CoListingUsecase struct {
	propertyAccess
	userRepo     repository.UserRepository
	tx           repository.Transactor
	auditService *service.AuditService
}

// NewCoListingUsecase creates a new co-listing usecase.
func NewCoListingUsecase(
	propertyRepo repository.PropertyRepository,
	coRepo repository.CoListingRepository,
	userRepo repository.UserRepository,
	tx repository.Transactor,
	auditService *service.AuditService,
) *CoListingUsecase {
	return &CoListingUsecase{
		propertyAccess: propertyAccess{propertyRepo: propertyRepo, coRepo: coRepo},
		userRepo:       userRepo,
		tx:             tx,
		auditService:   auditService,
	}
}

// List returns the marketing team of a property to admin, its primary
// marketing owner and its co-listing members.
func (uc *CoListingUsecase) List(ctx context.Context, actor Actor, propertyID uuid.UUID) ([]TeamMember, error) {
	if !actor.Authenticated() {
		return nil, errs.ErrUnauthorized
	}
	_, lvl, err := uc.load(ctx, actor, propertyID)
	if err != nil {
		return nil, err
	}
	if lvl != AccessAdmin && lvl != AccessPrimary && lvl != AccessCoListing {
		return nil, errs.Forbidden("You do not have access to this listing's marketing team")
	}
	team, err := uc.coRepo.ListByProperty(ctx, propertyID)
	if err != nil {
		return nil, errs.Internal(err)
	}
	return toTeam(team), nil
}

func requireAdmin(actor Actor) error {
	if !actor.Authenticated() {
		return errs.ErrUnauthorized
	}
	if !actor.IsAdmin() {
		return errs.Forbidden("Only admin can manage co-listing")
	}
	return nil
}

// eligibleMarketing checks the user exists, is active and has the sales role.
func (uc *CoListingUsecase) eligibleMarketing(ctx context.Context, id uuid.UUID) error {
	u, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return errs.Validation("Marketing user not found")
	}
	if u.Role != entity.RoleSales {
		return errs.Validation("User must have the marketing (sales) role")
	}
	if !u.IsActive() {
		return errs.Validation("User must be active")
	}
	return nil
}

// addLocked validates and inserts a co-listing row. Caller holds the property lock.
func (uc *CoListingUsecase) addLocked(ctx context.Context, actor Actor, propertyID, marketingID uuid.UUID) error {
	if err := uc.eligibleMarketing(ctx, marketingID); err != nil {
		return err
	}
	existing, err := uc.coRepo.Find(ctx, propertyID, marketingID)
	if err != nil {
		return err
	}
	if existing != nil {
		return errs.Conflict("User is already assigned to this property")
	}
	n, err := uc.coRepo.CountCoListing(ctx, propertyID)
	if err != nil {
		return err
	}
	if n >= entity.MaxCoListingPerProperty {
		return errs.Validation("A property can have at most 2 co-listing marketing users (3 including primary)")
	}
	return uc.coRepo.Create(ctx, &entity.PropertyMarketingAssignment{
		PropertyID:     propertyID,
		MarketingID:    marketingID,
		AssignmentType: entity.AssignmentTypeCoListing,
		AssignedBy:     actor.ID,
	})
}

func wrapTxErr(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := errs.IsAppError(err); ok {
		return err
	}
	return errs.Internal(err)
}

// Add assigns a co-listing marketing user. Admin only; transactional with a
// property row lock so concurrent requests cannot exceed the limit.
func (uc *CoListingUsecase) Add(ctx context.Context, actor Actor, propertyID, marketingID uuid.UUID, meta RequestMeta) ([]TeamMember, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	err := uc.tx.Do(ctx, func(ctx context.Context) error {
		if err := uc.propertyRepo.LockByID(ctx, propertyID); err != nil {
			return err
		}
		if err := uc.addLocked(ctx, actor, propertyID, marketingID); err != nil {
			return err
		}
		return uc.auditService.LogSync(ctx, &actor.ID, "added_co_listing", "co_listing", &propertyID,
			map[string]string{"marketing_id": marketingID.String()}, meta.IP, meta.UserAgent)
	})
	if err := wrapTxErr(err); err != nil {
		return nil, err
	}
	return uc.List(ctx, actor, propertyID)
}

// Remove removes a co-listing user. The primary owner cannot be removed here.
func (uc *CoListingUsecase) Remove(ctx context.Context, actor Actor, propertyID, marketingID uuid.UUID, meta RequestMeta) error {
	if err := requireAdmin(actor); err != nil {
		return err
	}
	err := uc.tx.Do(ctx, func(ctx context.Context) error {
		if err := uc.propertyRepo.LockByID(ctx, propertyID); err != nil {
			return err
		}
		if err := uc.requireCoListing(ctx, propertyID, marketingID); err != nil {
			return err
		}
		if _, err := uc.coRepo.DeleteCoListing(ctx, propertyID, marketingID); err != nil {
			return err
		}
		return uc.auditService.LogSync(ctx, &actor.ID, "removed_co_listing", "co_listing", &propertyID,
			map[string]string{"marketing_id": marketingID.String()}, meta.IP, meta.UserAgent)
	})
	return wrapTxErr(err)
}

// requireCoListing ensures the assignment exists and is not the primary one.
func (uc *CoListingUsecase) requireCoListing(ctx context.Context, propertyID, marketingID uuid.UUID) error {
	a, err := uc.coRepo.Find(ctx, propertyID, marketingID)
	if err != nil {
		return err
	}
	if a == nil {
		return errs.NotFound("Co-listing assignment")
	}
	if a.AssignmentType == entity.AssignmentTypePrimary {
		return errs.Validation("The primary marketing owner cannot be changed through co-listing management")
	}
	return nil
}

// Replace swaps one co-listing user for another atomically.
func (uc *CoListingUsecase) Replace(ctx context.Context, actor Actor, propertyID, oldMarketingID, newMarketingID uuid.UUID, meta RequestMeta) ([]TeamMember, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	if oldMarketingID == newMarketingID {
		return nil, errs.Validation("Replacement user must differ from the current user")
	}
	err := uc.tx.Do(ctx, func(ctx context.Context) error {
		if err := uc.propertyRepo.LockByID(ctx, propertyID); err != nil {
			return err
		}
		if err := uc.requireCoListing(ctx, propertyID, oldMarketingID); err != nil {
			return err
		}
		if _, err := uc.coRepo.DeleteCoListing(ctx, propertyID, oldMarketingID); err != nil {
			return err
		}
		if err := uc.addLocked(ctx, actor, propertyID, newMarketingID); err != nil {
			return err
		}
		return uc.auditService.LogSync(ctx, &actor.ID, "replaced_co_listing", "co_listing", &propertyID,
			map[string]string{"old_marketing_id": oldMarketingID.String(), "new_marketing_id": newMarketingID.String()},
			meta.IP, meta.UserAgent)
	})
	if err := wrapTxErr(err); err != nil {
		return nil, err
	}
	return uc.List(ctx, actor, propertyID)
}

// TransferPrimary moves primary ownership of a listing to another marketing user
// (admin only, audited). The previous primary becomes a co-listing member, so the
// team size never changes; if the new primary is not yet in the team there must be
// a free co-listing slot for the previous primary. Owner data follows the listing.
func (uc *CoListingUsecase) TransferPrimary(ctx context.Context, actor Actor, propertyID, newPrimaryID uuid.UUID, meta RequestMeta) ([]TeamMember, error) {
	if err := requireAdmin(actor); err != nil {
		return nil, err
	}
	err := uc.tx.Do(ctx, func(ctx context.Context) error {
		if err := uc.propertyRepo.LockByID(ctx, propertyID); err != nil {
			return err
		}
		p, err := uc.propertyRepo.FindByID(ctx, propertyID)
		if err != nil {
			return err
		}
		oldID := p.PrimaryMarketingID
		if newPrimaryID == oldID {
			return errs.Validation("The user is already the primary marketing owner")
		}
		if err := uc.eligibleMarketing(ctx, newPrimaryID); err != nil {
			return err
		}
		existing, err := uc.coRepo.Find(ctx, propertyID, newPrimaryID)
		if err != nil {
			return err
		}
		if existing == nil {
			n, err := uc.coRepo.CountCoListing(ctx, propertyID)
			if err != nil {
				return err
			}
			if n >= entity.MaxCoListingPerProperty {
				return errs.Validation("No free co-listing slot for the previous primary; remove a co-listing member first")
			}
		}
		if err := uc.coRepo.Delete(ctx, propertyID, oldID); err != nil {
			return err
		}
		if existing != nil {
			if err := uc.coRepo.Delete(ctx, propertyID, newPrimaryID); err != nil {
				return err
			}
		}
		if err := uc.propertyRepo.SetPrimaryMarketing(ctx, propertyID, newPrimaryID); err != nil {
			return err
		}
		if err := uc.coRepo.Create(ctx, &entity.PropertyMarketingAssignment{
			PropertyID: propertyID, MarketingID: newPrimaryID, AssignmentType: entity.AssignmentTypePrimary, AssignedBy: actor.ID,
		}); err != nil {
			return err
		}
		if err := uc.coRepo.Create(ctx, &entity.PropertyMarketingAssignment{
			PropertyID: propertyID, MarketingID: oldID, AssignmentType: entity.AssignmentTypeCoListing, AssignedBy: actor.ID,
		}); err != nil {
			return err
		}
		return uc.auditService.LogSync(ctx, &actor.ID, "transferred_primary_ownership", "properties", &propertyID,
			map[string]string{"old_primary_id": oldID.String(), "new_primary_id": newPrimaryID.String()}, meta.IP, meta.UserAgent)
	})
	if err := wrapTxErr(err); err != nil {
		return nil, err
	}
	return uc.List(ctx, actor, propertyID)
}
