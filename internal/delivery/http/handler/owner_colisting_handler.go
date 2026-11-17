package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/leventa-backend/internal/usecase"
	"github.com/leventa-backend/pkg/response"
	"github.com/leventa-backend/pkg/validator"
)

// OwnerHandler serves owner PII through the owning property only.
type OwnerHandler struct {
	ownerUsecase *usecase.OwnerUsecase
	validator    *validator.Validator
}

// NewOwnerHandler creates a new owner handler.
func NewOwnerHandler(ownerUsecase *usecase.OwnerUsecase, validator *validator.Validator) *OwnerHandler {
	return &OwnerHandler{ownerUsecase: ownerUsecase, validator: validator}
}

// Get returns the owner of a property (primary marketing only).
func (h *OwnerHandler) Get(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	owner, err := h.ownerUsecase.GetByProperty(c.Context(), actorFrom(c), propertyID)
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Owner retrieved", owner)
}

// Update updates the owner of a property (primary marketing only).
func (h *OwnerHandler) Update(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	var input usecase.UpdateOwnerInput
	if err := c.Bind().Body(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(input); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	owner, err := h.ownerUsecase.UpdateByProperty(c.Context(), actorFrom(c), propertyID, input, metaFrom(c))
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Owner updated", owner)
}

// CoListingHandler handles co-listing management.
type CoListingHandler struct {
	uc        *usecase.CoListingUsecase
	validator *validator.Validator
}

// NewCoListingHandler creates a new co-listing handler.
func NewCoListingHandler(uc *usecase.CoListingUsecase, validator *validator.Validator) *CoListingHandler {
	return &CoListingHandler{uc: uc, validator: validator}
}

// List returns the marketing team of a property.
func (h *CoListingHandler) List(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	team, err := h.uc.List(c.Context(), actorFrom(c), propertyID)
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Marketing team retrieved", team)
}

type marketingBody struct {
	MarketingID uuid.UUID `json:"marketing_id" validate:"required"`
}

// Add assigns a co-listing marketing user (admin only).
func (h *CoListingHandler) Add(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	var body marketingBody
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	team, err := h.uc.Add(c.Context(), actorFrom(c), propertyID, body.MarketingID, metaFrom(c))
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusCreated, "Co-listing marketing added", team)
}

// Remove removes a co-listing marketing user (admin only).
func (h *CoListingHandler) Remove(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	marketingID, err := uuid.Parse(c.Params("marketingId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid marketing ID", nil)
	}
	if err := h.uc.Remove(c.Context(), actorFrom(c), propertyID, marketingID, metaFrom(c)); err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Co-listing marketing removed", nil)
}

// Replace replaces a co-listing marketing user with another (admin only).
func (h *CoListingHandler) Replace(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	oldID, err := uuid.Parse(c.Params("marketingId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid marketing ID", nil)
	}
	var body marketingBody
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	team, err := h.uc.Replace(c.Context(), actorFrom(c), propertyID, oldID, body.MarketingID, metaFrom(c))
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Co-listing marketing replaced", team)
}

// Transfer moves primary ownership to another marketing user (admin only).
func (h *CoListingHandler) Transfer(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	var body marketingBody
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	team, err := h.uc.TransferPrimary(c.Context(), actorFrom(c), propertyID, body.MarketingID, metaFrom(c))
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Primary ownership transferred", team)
}
