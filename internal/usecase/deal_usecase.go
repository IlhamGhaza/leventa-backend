package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/internal/service"
	"github.com/leventa-backend/pkg/errs"
)

// DealUsecase lets Admin record the outcome of a property: a Deal (sold/rented
// plus commission) or Cancelled (the owner no longer sells; property archived).
// Marketing reports to Admin outside the system (D-027).
type DealUsecase struct {
	propertyRepo         repository.PropertyRepository
	dealRepo             repository.DealRepository
	commissionRepo       repository.CommissionRepository
	tx                   repository.Transactor
	auditService         *service.AuditService
	defaultCommissionPct float64
}

// NewDealUsecase creates a deal usecase.
func NewDealUsecase(
	propertyRepo repository.PropertyRepository,
	dealRepo repository.DealRepository,
	commissionRepo repository.CommissionRepository,
	tx repository.Transactor,
	auditService *service.AuditService,
	defaultCommissionPct float64,
) *DealUsecase {
	return &DealUsecase{
		propertyRepo:         propertyRepo,
		dealRepo:             dealRepo,
		commissionRepo:       commissionRepo,
		tx:                   tx,
		auditService:         auditService,
		defaultCommissionPct: defaultCommissionPct,
	}
}

// RecordDealInput is the Admin input for a finalized transaction.
type RecordDealInput struct {
	Type      string  `json:"type" validate:"required,oneof=sale rent"`
	Price     float64 `json:"price" validate:"required,gt=0"`
	BuyerName string  `json:"buyer_name" validate:"required,min=2,max=255"`
	DealDate  string  `json:"deal_date" validate:"omitempty,datetime=2006-01-02"`
	Notes     string  `json:"notes" validate:"omitempty,max=2000"`
}

// CancelInput is the Admin input for a cancellation.
type CancelInput struct {
	Notes string `json:"notes" validate:"omitempty,max=2000"`
}

// DealResult is returned after recording a deal.
type DealResult struct {
	Deal       *entity.Deal       `json:"deal"`
	Commission *entity.Commission `json:"commission,omitempty"`
}

func requireDealAdmin(actor Actor) error {
	if !actor.Authenticated() {
		return errs.ErrUnauthorized
	}
	if !actor.IsAdmin() {
		return errs.Forbidden("Only admin can record a deal or cancellation")
	}
	return nil
}

// recordable checks that a property can still receive an outcome.
func recordable(status string) error {
	switch status {
	case entity.PropertyStatusSold, entity.PropertyStatusRented:
		return errs.Conflict("Property already has a deal")
	case entity.PropertyStatusArchived:
		return errs.Conflict("Property is archived")
	}
	return nil
}

// RecordDeal records a deal in one transaction: lock the property, create the
// deal, set sold/rented, create the commission for the primary marketing and
// write the audit entry. Any failure rolls everything back.
func (uc *DealUsecase) RecordDeal(ctx context.Context, actor Actor, propertyID uuid.UUID, in RecordDealInput, meta RequestMeta) (*DealResult, error) {
	if err := requireDealAdmin(actor); err != nil {
		return nil, err
	}
	date, err := parseDealDate(in.DealDate)
	if err != nil {
		return nil, err
	}

	var result DealResult
	err = uc.tx.Do(ctx, func(ctx context.Context) error {
		if err := uc.propertyRepo.LockByID(ctx, propertyID); err != nil {
			return err
		}
		// Re-read under the lock so two concurrent outcomes cannot both succeed.
		p, err := uc.propertyRepo.FindByID(ctx, propertyID)
		if err != nil {
			return err
		}
		if err := recordable(p.Status); err != nil {
			return err
		}
		if p.TransactionType != in.Type {
			return errs.Validation("Deal type does not match the property's transaction type")
		}

		typ, price, buyer := in.Type, in.Price, in.BuyerName
		deal := &entity.Deal{
			PropertyID: propertyID,
			SalesID:    p.PrimaryMarketingID,
			RecordedBy: actor.ID,
			Outcome:    entity.DealOutcomeDeal,
			Type:       &typ,
			Price:      &price,
			BuyerName:  &buyer,
			DealDate:   date,
			Notes:      in.Notes,
		}
		if err := uc.dealRepo.Create(ctx, deal); err != nil {
			return err
		}

		status := entity.PropertyStatusSold
		if in.Type == entity.DealTypeRent {
			status = entity.PropertyStatusRented
		}
		if err := uc.propertyRepo.UpdateStatus(ctx, propertyID, status); err != nil {
			return err
		}

		commission := &entity.Commission{
			DealID:     deal.ID,
			SalesID:    p.PrimaryMarketingID,
			Percentage: uc.defaultCommissionPct,
			Amount:     entity.CalculateCommission(in.Price, uc.defaultCommissionPct),
			Status:     entity.CommissionStatusPending,
		}
		if err := uc.commissionRepo.Create(ctx, commission); err != nil {
			return err
		}

		result.Deal, result.Commission = deal, commission
		return uc.auditService.LogSync(ctx, &actor.ID, "recorded_deal", "deals", &deal.ID,
			map[string]string{"property_id": propertyID.String(), "type": in.Type}, meta.IP, meta.UserAgent)
	})
	if err != nil {
		return nil, mapTxError(err)
	}
	return &result, nil
}

// Cancel records that the owner no longer sells or rents the property. The
// property becomes archived (hidden from lists); no commission is created.
func (uc *DealUsecase) Cancel(ctx context.Context, actor Actor, propertyID uuid.UUID, in CancelInput, meta RequestMeta) (*entity.Deal, error) {
	if err := requireDealAdmin(actor); err != nil {
		return nil, err
	}
	var deal *entity.Deal
	err := uc.tx.Do(ctx, func(ctx context.Context) error {
		if err := uc.propertyRepo.LockByID(ctx, propertyID); err != nil {
			return err
		}
		p, err := uc.propertyRepo.FindByID(ctx, propertyID)
		if err != nil {
			return err
		}
		if err := recordable(p.Status); err != nil {
			return err
		}
		deal = &entity.Deal{
			PropertyID: propertyID,
			SalesID:    p.PrimaryMarketingID,
			RecordedBy: actor.ID,
			Outcome:    entity.DealOutcomeCancelled,
			DealDate:   time.Now(),
			Notes:      in.Notes,
		}
		if err := uc.dealRepo.Create(ctx, deal); err != nil {
			return err
		}
		if err := uc.propertyRepo.UpdateStatus(ctx, propertyID, entity.PropertyStatusArchived); err != nil {
			return err
		}
		return uc.auditService.LogSync(ctx, &actor.ID, "cancelled_property", "deals", &deal.ID,
			map[string]string{"property_id": propertyID.String()}, meta.IP, meta.UserAgent)
	})
	if err != nil {
		return nil, mapTxError(err)
	}
	return deal, nil
}

// Get returns a deal. Marketing sees only deals of their own listings; another
// marketing user's deal is reported as not found.
func (uc *DealUsecase) Get(ctx context.Context, actor Actor, id uuid.UUID) (*entity.Deal, error) {
	if !actor.IsSales() && !actor.IsAdmin() {
		return nil, errs.Forbidden("You do not have access to deals")
	}
	d, err := uc.dealRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if actor.IsSales() && d.SalesID != actor.ID {
		return nil, errs.NotFound("Deal")
	}
	return d, nil
}

// List returns deals scoped to the actor.
func (uc *DealUsecase) List(ctx context.Context, actor Actor, page, limit int) ([]entity.Deal, int64, error) {
	var salesID *uuid.UUID
	switch {
	case actor.IsSales():
		salesID = &actor.ID
	case actor.IsAdmin():
	default:
		return nil, 0, errs.Forbidden("You do not have access to deals")
	}
	return uc.dealRepo.List(ctx, page, limit, salesID)
}

func parseDealDate(s string) (time.Time, error) {
	if s == "" {
		return time.Now(), nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, errs.Validation("deal_date must be YYYY-MM-DD")
	}
	return t, nil
}

// mapTxError keeps application errors and hides everything else.
func mapTxError(err error) error {
	if _, ok := errs.IsAppError(err); ok {
		return err
	}
	return errs.Internal(err)
}
