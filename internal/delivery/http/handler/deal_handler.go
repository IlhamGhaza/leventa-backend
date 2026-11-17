package handler

import (
	"math"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/leventa-backend/internal/usecase"
	"github.com/leventa-backend/pkg/response"
	"github.com/leventa-backend/pkg/validator"
)

// DealHandler serves deal records and the Admin-only outcome endpoints.
type DealHandler struct {
	uc        *usecase.DealUsecase
	validator *validator.Validator
}

// NewDealHandler creates a deal handler.
func NewDealHandler(uc *usecase.DealUsecase, v *validator.Validator) *DealHandler {
	return &DealHandler{uc: uc, validator: v}
}

// RecordDeal records a deal on a property (Admin only).
func (h *DealHandler) RecordDeal(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	var in usecase.RecordDealInput
	if err := c.Bind().Body(&in); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(in); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	res, err := h.uc.RecordDeal(c.Context(), actorFrom(c), propertyID, in, metaFrom(c))
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusCreated, "Deal recorded", res)
}

// Cancel records a cancellation on a property (Admin only).
func (h *DealHandler) Cancel(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	var in usecase.CancelInput
	if len(c.Body()) > 0 {
		if err := c.Bind().Body(&in); err != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
		}
	}
	if errs := h.validator.Validate(in); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	deal, err := h.uc.Cancel(c.Context(), actorFrom(c), propertyID, in, metaFrom(c))
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusCreated, "Cancellation recorded", deal)
}

// GetByID returns a deal (Admin: any; Marketing: own listings only).
func (h *DealHandler) GetByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid deal ID", nil)
	}
	deal, err := h.uc.Get(c.Context(), actorFrom(c), id)
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Deal retrieved", deal)
}

// List returns deals scoped to the actor.
func (h *DealHandler) List(c fiber.Ctx) error {
	page, limit := paginationParams(c)
	deals, total, err := h.uc.List(c.Context(), actorFrom(c), page, limit)
	if err != nil {
		return err
	}
	totalPages := int64(math.Ceil(float64(total) / float64(limit)))
	return response.SuccessWithPagination(c, "Deals retrieved", deals, response.PaginationMeta{Page: page, Limit: limit, Total: total, TotalPages: totalPages})
}
