package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/internal/service"
	"github.com/leventa-backend/pkg/errs"
)

// OwnerUsecase serves owner PII. Every operation is scoped to the primary
// marketing owner of the property; no other role (including admin and staff)
// has access, and there is no owner list endpoint.
type OwnerUsecase struct {
	propertyAccess
	ownerRepo    repository.OwnerRepository
	auditService *service.AuditService
}

// NewOwnerUsecase creates a new owner usecase.
func NewOwnerUsecase(
	propertyRepo repository.PropertyRepository,
	coRepo repository.CoListingRepository,
	ownerRepo repository.OwnerRepository,
	auditService *service.AuditService,
) *OwnerUsecase {
	return &OwnerUsecase{
		propertyAccess: propertyAccess{propertyRepo: propertyRepo, coRepo: coRepo},
		ownerRepo:      ownerRepo,
		auditService:   auditService,
	}
}

// UpdateOwnerInput is a partial owner update.
type UpdateOwnerInput struct {
	Name           *string `json:"name" validate:"omitempty,min=2,max=255"`
	Email          *string `json:"email" validate:"omitempty,email,max=255"`
	Phone          *string `json:"phone" validate:"omitempty,max=50"`
	PrivateAddress *string `json:"private_address" validate:"omitempty,max=2000"`
}

// GetByProperty returns the owner of a property to its primary marketing owner.
func (uc *OwnerUsecase) GetByProperty(ctx context.Context, actor Actor, propertyID uuid.UUID) (*OwnerView, error) {
	p, err := uc.requirePrimary(ctx, actor, propertyID)
	if err != nil {
		return nil, err
	}
	owner, err := uc.ownerRepo.FindByID(ctx, p.OwnerID)
	if err != nil {
		return nil, err
	}
	v := toOwnerView(owner)
	return &v, nil
}

// UpdateByProperty updates owner details. Primary marketing owner only.
// Audit metadata carries only the property ID, never owner values.
func (uc *OwnerUsecase) UpdateByProperty(ctx context.Context, actor Actor, propertyID uuid.UUID, in UpdateOwnerInput, meta RequestMeta) (*OwnerView, error) {
	p, err := uc.requirePrimary(ctx, actor, propertyID)
	if err != nil {
		return nil, err
	}
	owner, err := uc.ownerRepo.FindByID(ctx, p.OwnerID)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		owner.Name = *in.Name
	}
	if in.Email != nil {
		owner.Email = *in.Email
	}
	if in.Phone != nil {
		owner.Phone = *in.Phone
	}
	if in.PrivateAddress != nil {
		owner.PrivateAddress = *in.PrivateAddress
	}
	if err := uc.ownerRepo.Update(ctx, owner); err != nil {
		return nil, errs.Internal(err)
	}
	uc.auditService.Log(&actor.ID, "updated_owner", "owners", &propertyID, nil, meta.IP, meta.UserAgent)
	v := toOwnerView(owner)
	return &v, nil
}
