package handler

import (
	"math"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/usecase"
	"github.com/leventa-backend/pkg/response"
	"github.com/leventa-backend/pkg/validator"
)

// PropertyHandler handles property HTTP requests.
type PropertyHandler struct {
	propertyUsecase *usecase.PropertyUsecase
	validator       *validator.Validator
}

// NewPropertyHandler creates a new property handler.
func NewPropertyHandler(propertyUsecase *usecase.PropertyUsecase, validator *validator.Validator) *PropertyHandler {
	return &PropertyHandler{propertyUsecase: propertyUsecase, validator: validator}
}

// queryFloat parses an optional float query parameter. ok=false means malformed.
func queryFloat(c fiber.Ctx, key string) (*float64, bool) {
	raw := c.Query(key)
	if raw == "" {
		return nil, true
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 {
		return nil, false
	}
	return &v, true
}

func queryInt(c fiber.Ctx, key string) (*int, bool) {
	raw := c.Query(key)
	if raw == "" {
		return nil, true
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v < 0 {
		return nil, false
	}
	return &v, true
}

func queryUUID(c fiber.Ctx, key string) (*uuid.UUID, bool) {
	raw := c.Query(key)
	if raw == "" {
		return nil, true
	}
	v, err := uuid.Parse(raw)
	if err != nil {
		return nil, false
	}
	return &v, true
}

// List returns a paginated, role-scoped list of properties.
func (h *PropertyHandler) List(c fiber.Ctx) error {
	page, limit := paginationParams(c)
	in := usecase.ListPropertiesInput{
		Search:          c.Query("search"),
		Type:            c.Query("type"),
		TransactionType: c.Query("transaction_type"),
		Status:          c.Query("status"),
		Scope:           c.Query("scope"),
		History:         c.Query("history") == "true",
		IncludeHistory:  c.Query("include_history") == "true",
		Sort:            c.Query("sort"),
		Page:            page,
		Limit:           limit,
	}

	var ok [13]bool
	in.ProvinceID, ok[0] = queryUUID(c, "province_id")
	in.CityID, ok[1] = queryUUID(c, "city_id")
	in.MinPrice, ok[2] = queryFloat(c, "min_price")
	in.MaxPrice, ok[3] = queryFloat(c, "max_price")
	in.MinLandArea, ok[4] = queryFloat(c, "min_land_area")
	in.MaxLandArea, ok[5] = queryFloat(c, "max_land_area")
	in.MinBuildingArea, ok[6] = queryFloat(c, "min_building_area")
	in.MaxBuildingArea, ok[7] = queryFloat(c, "max_building_area")
	in.Bedrooms, ok[8] = queryInt(c, "bedrooms")
	in.Bathrooms, ok[9] = queryInt(c, "bathrooms")
	in.DistrictID, ok[10] = queryUUID(c, "district_id")
	ok[11] = true
	if raw := c.Query("facility_ids"); raw != "" {
		for _, part := range strings.Split(raw, ",") {
			id, err := uuid.Parse(strings.TrimSpace(part))
			if err != nil {
				ok[12] = false
				break
			}
			in.FacilityIDs = append(in.FacilityIDs, id)
			ok[12] = true
		}
	} else {
		ok[12] = true
	}
	for _, good := range ok {
		if !good {
			return response.Error(c, fiber.StatusBadRequest, "Invalid query parameter", nil)
		}
	}

	properties, total, err := h.propertyUsecase.List(c.Context(), actorFrom(c), in)
	if err != nil {
		return err
	}
	totalPages := int64(math.Ceil(float64(total) / float64(limit)))
	return response.SuccessWithPagination(c, "Properties retrieved", properties, response.PaginationMeta{
		Page: page, Limit: limit, Total: total, TotalPages: totalPages,
	})
}

// GetByID returns the role-appropriate projection of a property.
func (h *PropertyHandler) GetByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	property, err := h.propertyUsecase.GetByID(c.Context(), actorFrom(c), id)
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Property retrieved", property)
}

// Create creates a listing owned by the authenticated marketing user.
func (h *PropertyHandler) Create(c fiber.Ctx) error {
	var input usecase.CreatePropertyInput
	if err := c.Bind().Body(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(input); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	property, err := h.propertyUsecase.Create(c.Context(), actorFrom(c), input, metaFrom(c))
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusCreated, "Property created successfully", property)
}

// Update partially updates a property (primary marketing only).
func (h *PropertyHandler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	var input usecase.UpdatePropertyInput
	if err := c.Bind().Body(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(input); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	property, err := h.propertyUsecase.Update(c.Context(), actorFrom(c), id, input, metaFrom(c))
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Property updated successfully", property)
}

// Delete deletes or archives a property (primary marketing only).
func (h *PropertyHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	res, err := h.propertyUsecase.Delete(c.Context(), actorFrom(c), id, metaFrom(c))
	if err != nil {
		return err
	}
	if res.Archived {
		return response.Success(c, fiber.StatusOK, "Property is part of a deal and was archived instead of deleted", nil)
	}
	return response.Success(c, fiber.StatusOK, "Property deleted successfully", nil)
}

// UpdateStatus updates a property's status (primary marketing only).
func (h *PropertyHandler) UpdateStatus(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	var body struct {
		Status string `json:"status" validate:"required,oneof=draft available reserved maintenance archived"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	if err := h.propertyUsecase.UpdateStatus(c.Context(), actorFrom(c), id, body.Status, metaFrom(c)); err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Property status updated", nil)
}

// UploadImage uploads an image for a property (primary marketing only).
func (h *PropertyHandler) UploadImage(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	file, err := c.FormFile("image")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Image file is required", nil)
	}
	f, err := file.Open()
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Failed to read file", nil)
	}
	defer f.Close()

	image, err := h.propertyUsecase.UploadImage(c.Context(), actorFrom(c), id, f, file)
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusCreated, "Image uploaded successfully", image)
}

// DeleteImage deletes a property image (primary marketing only).
func (h *PropertyHandler) DeleteImage(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	imageID, err := uuid.Parse(c.Params("imageId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid image ID", nil)
	}
	if err := h.propertyUsecase.DeleteImage(c.Context(), actorFrom(c), propertyID, imageID); err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Image deleted successfully", nil)
}

// SetThumbnail marks an image as the property's cover photo.
func (h *PropertyHandler) SetThumbnail(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	imageID, err := uuid.Parse(c.Params("imageId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid image ID", nil)
	}
	if err := h.propertyUsecase.SetThumbnail(c.Context(), actorFrom(c), propertyID, imageID); err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Thumbnail updated successfully", nil)
}

// --- Helpers used across handlers ---

func paginationParams(c fiber.Ctx) (int, int) {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > usecase.MaxPageSize {
		limit = 20
	}
	return page, limit
}

func getUserID(c fiber.Ctx) uuid.UUID {
	id, _ := c.Locals("user_id").(uuid.UUID)
	return id
}

func getUserRole(c fiber.Ctx) string {
	role, _ := c.Locals("user_role").(string)
	return role
}

func isAdmin(c fiber.Ctx) bool {
	return getUserRole(c) == entity.RoleAdmin
}

// actorFrom builds the actor from the verified token claims set by the auth
// middleware. It never reads identity from the request body or query.
func actorFrom(c fiber.Ctx) usecase.Actor {
	return usecase.Actor{ID: getUserID(c), Role: getUserRole(c)}
}

func metaFrom(c fiber.Ctx) usecase.RequestMeta {
	return usecase.RequestMeta{IP: c.IP(), UserAgent: c.Get("User-Agent")}
}
