package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/internal/service"
	"github.com/leventa-backend/pkg/errs"
)

// CommissionUsecase serves commission records. Marketing sees only their own;
// admin sees all and is the only role that can change percentage or status.
type CommissionUsecase struct {
	repo         repository.CommissionRepository
	auditService *service.AuditService
}

// NewCommissionUsecase creates a commission usecase.
func NewCommissionUsecase(repo repository.CommissionRepository, auditService *service.AuditService) *CommissionUsecase {
	return &CommissionUsecase{repo: repo, auditService: auditService}
}

// Get returns a commission. Another marketing user's commission is reported as not found.
func (uc *CommissionUsecase) Get(ctx context.Context, actor Actor, id uuid.UUID) (*entity.Commission, error) {
	if !actor.IsSales() && !actor.IsAdmin() {
		return nil, errs.Forbidden("You do not have access to commissions")
	}
	c, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if actor.IsSales() && c.SalesID != actor.ID {
		return nil, errs.NotFound("Commission")
	}
	return c, nil
}

// List returns commissions scoped to the actor.
func (uc *CommissionUsecase) List(ctx context.Context, actor Actor, page, limit int, status string) ([]entity.Commission, int64, error) {
	var salesID *uuid.UUID
	switch {
	case actor.IsSales():
		salesID = &actor.ID
	case actor.IsAdmin():
	default:
		return nil, 0, errs.Forbidden("You do not have access to commissions")
	}
	return uc.repo.List(ctx, page, limit, salesID, status)
}

// UpdateStatus changes the status (admin only). pending → paid|cancelled; final states are immutable.
func (uc *CommissionUsecase) UpdateStatus(ctx context.Context, actor Actor, id uuid.UUID, status string, meta RequestMeta) (*entity.Commission, error) {
	if !actor.IsAdmin() {
		return nil, errs.Forbidden("Only admin can change commission status")
	}
	c, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Status != entity.CommissionStatusPending {
		return nil, errs.Conflict("Commission is already " + c.Status)
	}
	if status != entity.CommissionStatusPaid && status != entity.CommissionStatusCancelled {
		return nil, errs.Validation("status must be paid or cancelled")
	}
	old := c.Status
	c.Status = status
	if err := uc.repo.Update(ctx, c); err != nil {
		return nil, errs.Internal(err)
	}
	uc.auditService.Log(&actor.ID, "changed_commission_status", "commissions", &id,
		map[string]string{"old_status": old, "new_status": status}, meta.IP, meta.UserAgent)
	return c, nil
}

// UpdatePercentage overrides the percentage of a pending commission (admin only)
// and recalculates the amount from the deal price.
func (uc *CommissionUsecase) UpdatePercentage(ctx context.Context, actor Actor, id uuid.UUID, pct float64, meta RequestMeta) (*entity.Commission, error) {
	if !actor.IsAdmin() {
		return nil, errs.Forbidden("Only admin can change commission percentage")
	}
	if pct < 0 || pct > 100 {
		return nil, errs.Validation("percentage must be between 0 and 100")
	}
	c, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if c.Status != entity.CommissionStatusPending {
		return nil, errs.Conflict("Only a pending commission can be changed")
	}
	if c.Deal == nil || c.Deal.Price == nil {
		return nil, errs.ErrInternalServer
	}
	old := c.Percentage
	c.Percentage = pct
	c.Amount = entity.CalculateCommission(*c.Deal.Price, pct)
	if err := uc.repo.Update(ctx, c); err != nil {
		return nil, errs.Internal(err)
	}
	uc.auditService.Log(&actor.ID, "changed_commission_percentage", "commissions", &id,
		map[string]float64{"old_percentage": old, "new_percentage": pct}, meta.IP, meta.UserAgent)
	return c, nil
}
