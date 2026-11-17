package handler

import (
	"math"
	"strings"

	"github.com/leventa-backend/pkg/errs"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/internal/usecase"
	"github.com/leventa-backend/pkg/response"
	"github.com/leventa-backend/pkg/validator"
)

// --- User Handler (Admin) ---

type UserHandler struct {
	userUsecase *usecase.UserUsecase
	validator   *validator.Validator
}

func NewUserHandler(userUsecase *usecase.UserUsecase, validator *validator.Validator) *UserHandler {
	return &UserHandler{userUsecase: userUsecase, validator: validator}
}

func (h *UserHandler) List(c fiber.Ctx) error {
	page, limit := paginationParams(c)
	users, total, err := h.userUsecase.ListUsers(c.Context(), page, limit, c.Query("search"), c.Query("role"), c.Query("status"))
	if err != nil {
		return err
	}
	totalPages := int64(math.Ceil(float64(total) / float64(limit)))
	return response.SuccessWithPagination(c, "Users retrieved", users, response.PaginationMeta{Page: page, Limit: limit, Total: total, TotalPages: totalPages})
}

func (h *UserHandler) GetByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", nil)
	}
	user, err := h.userUsecase.GetUser(c.Context(), id)
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "User retrieved", user)
}

func (h *UserHandler) Create(c fiber.Ctx) error {
	var input usecase.CreateUserInput
	if err := c.Bind().Body(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(input); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	user, err := h.userUsecase.CreateUser(c.Context(), input, getUserID(c), c.IP(), c.Get("User-Agent"))
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusCreated, "User created", user)
}

func (h *UserHandler) Update(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", nil)
	}
	var input usecase.UpdateUserInput
	if err := c.Bind().Body(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(input); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	user, err := h.userUsecase.UpdateUser(c.Context(), id, input)
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "User updated", user)
}

func (h *UserHandler) UpdateStatus(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", nil)
	}
	var body struct {
		Status string `json:"status" validate:"required,oneof=active inactive"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	if err := h.userUsecase.UpdateStatus(c.Context(), id, body.Status, getUserID(c), c.IP(), c.Get("User-Agent")); err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "User status updated", nil)
}

func (h *UserHandler) UpdateRole(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", nil)
	}
	var body struct {
		Role string `json:"role" validate:"required,oneof=sales admin"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}

	// Rule 17: Sales cannot change their own role
	if getUserRole(c) == entity.RoleSales && id == getUserID(c) {
		return response.Error(c, fiber.StatusForbidden, "Sales cannot change their own role", nil)
	}

	if err := h.userUsecase.UpdateRole(c.Context(), id, body.Role, getUserID(c), c.IP(), c.Get("User-Agent")); err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "User role updated", nil)
}

func (h *UserHandler) Delete(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid user ID", nil)
	}
	if err := h.userUsecase.DeleteUser(c.Context(), id); err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "User deleted", nil)
}

// --- Commission Handler ---

type CommissionHandler struct {
	uc        *usecase.CommissionUsecase
	validator *validator.Validator
}

func NewCommissionHandler(uc *usecase.CommissionUsecase, validator *validator.Validator) *CommissionHandler {
	return &CommissionHandler{uc: uc, validator: validator}
}

func (h *CommissionHandler) GetByID(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid commission ID", nil)
	}
	commission, err := h.uc.Get(c.Context(), actorFrom(c), id)
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Commission retrieved", commission)
}

func (h *CommissionHandler) List(c fiber.Ctx) error {
	page, limit := paginationParams(c)
	commissions, total, err := h.uc.List(c.Context(), actorFrom(c), page, limit, c.Query("status"))
	if err != nil {
		return err
	}
	totalPages := int64(math.Ceil(float64(total) / float64(limit)))
	return response.SuccessWithPagination(c, "Commissions retrieved", commissions, response.PaginationMeta{Page: page, Limit: limit, Total: total, TotalPages: totalPages})
}

func (h *CommissionHandler) UpdateStatus(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid commission ID", nil)
	}
	var body struct {
		Status string `json:"status" validate:"required,oneof=paid cancelled"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	commission, err := h.uc.UpdateStatus(c.Context(), actorFrom(c), id, body.Status, metaFrom(c))
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Commission status updated", commission)
}

func (h *CommissionHandler) UpdatePercentage(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid commission ID", nil)
	}
	var body struct {
		Percentage *float64 `json:"percentage" validate:"required,gte=0,lte=100"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	commission, err := h.uc.UpdatePercentage(c.Context(), actorFrom(c), id, *body.Percentage, metaFrom(c))
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Commission percentage updated", commission)
}

// --- Audit Log Handler ---

type AuditLogHandler struct {
	auditLogRepo repository.AuditLogRepository
}

func NewAuditLogHandler(auditLogRepo repository.AuditLogRepository) *AuditLogHandler {
	return &AuditLogHandler{auditLogRepo: auditLogRepo}
}

func (h *AuditLogHandler) List(c fiber.Ctx) error {
	page, limit := paginationParams(c)
	var userID *uuid.UUID
	if uid := c.Query("user_id"); uid != "" {
		if id, err := uuid.Parse(uid); err == nil {
			userID = &id
		}
	}
	logs, total, err := h.auditLogRepo.List(c.Context(), page, limit, c.Query("module"), c.Query("action"), userID)
	if err != nil {
		return err
	}
	totalPages := int64(math.Ceil(float64(total) / float64(limit)))
	return response.SuccessWithPagination(c, "Audit logs retrieved", logs, response.PaginationMeta{Page: page, Limit: limit, Total: total, TotalPages: totalPages})
}

// --- Master Data Handler ---

type MasterDataHandler struct {
	masterDataRepo repository.MasterDataRepository
	validator      *validator.Validator
}

func NewMasterDataHandler(masterDataRepo repository.MasterDataRepository, validator *validator.Validator) *MasterDataHandler {
	return &MasterDataHandler{masterDataRepo: masterDataRepo, validator: validator}
}

func (h *MasterDataHandler) ListProvinces(c fiber.Ctx) error {
	provinces, err := h.masterDataRepo.ListProvinces(c.Context())
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Provinces retrieved", provinces)
}

func (h *MasterDataHandler) ListCities(c fiber.Ctx) error {
	var provinceID *uuid.UUID
	if pid := c.Query("province_id"); pid != "" {
		if id, err := uuid.Parse(pid); err == nil {
			provinceID = &id
		}
	}
	cities, err := h.masterDataRepo.ListCities(c.Context(), provinceID)
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Cities retrieved", cities)
}

func (h *MasterDataHandler) CreateProvince(c fiber.Ctx) error {
	var body struct {
		Name string `json:"name" validate:"required,min=2,max=255"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	province := &entity.Province{Name: body.Name}
	if err := h.masterDataRepo.CreateProvince(c.Context(), province); err != nil {
		return err
	}
	return response.Success(c, fiber.StatusCreated, "Province created", province)
}

func (h *MasterDataHandler) CreateCity(c fiber.Ctx) error {
	var body struct {
		ProvinceID uuid.UUID `json:"province_id" validate:"required"`
		Name       string    `json:"name" validate:"required,min=2,max=255"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	city := &entity.City{ProvinceID: body.ProvinceID, Name: body.Name}
	if err := h.masterDataRepo.CreateCity(c.Context(), city); err != nil {
		return err
	}
	return response.Success(c, fiber.StatusCreated, "City created", city)
}

type masterBody struct {
	Name string `json:"name" validate:"required,min=2,max=255"`
}

func (h *MasterDataHandler) UpdateProvince(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid province ID", nil)
	}
	var body masterBody
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	province, err := h.masterDataRepo.FindProvinceByID(c.Context(), id)
	if err != nil {
		return err
	}
	province.Name = body.Name
	province.Cities = nil
	if err := h.masterDataRepo.UpdateProvince(c.Context(), province); err != nil {
		return duplicateOrInternal(err)
	}
	return response.Success(c, fiber.StatusOK, "Province updated", province)
}

func (h *MasterDataHandler) DeleteProvince(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid province ID", nil)
	}
	if _, err := h.masterDataRepo.FindProvinceByID(c.Context(), id); err != nil {
		return err
	}
	if err := h.masterDataRepo.DeleteProvince(c.Context(), id); err != nil {
		return referencedOrInternal(err)
	}
	return response.Success(c, fiber.StatusOK, "Province deleted", nil)
}

func (h *MasterDataHandler) UpdateCity(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid city ID", nil)
	}
	var body masterBody
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	city, err := h.masterDataRepo.FindCityByID(c.Context(), id)
	if err != nil {
		return err
	}
	city.Name = body.Name
	city.Province = nil
	if err := h.masterDataRepo.UpdateCity(c.Context(), city); err != nil {
		return duplicateOrInternal(err)
	}
	return response.Success(c, fiber.StatusOK, "City updated", city)
}

func (h *MasterDataHandler) DeleteCity(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid city ID", nil)
	}
	if _, err := h.masterDataRepo.FindCityByID(c.Context(), id); err != nil {
		return err
	}
	if err := h.masterDataRepo.DeleteCity(c.Context(), id); err != nil {
		return referencedOrInternal(err)
	}
	return response.Success(c, fiber.StatusOK, "City deleted", nil)
}

func duplicateOrInternal(err error) error {
	if strings.Contains(err.Error(), "23505") || strings.Contains(err.Error(), "duplicate key") {
		return errs.Conflict("Name already exists")
	}
	return errs.Internal(err)
}

func referencedOrInternal(err error) error {
	if strings.Contains(err.Error(), "23503") || strings.Contains(err.Error(), "foreign key") {
		return errs.Conflict("Record is still referenced and cannot be deleted")
	}
	return errs.Internal(err)
}

// --- Districts ---

func (h *MasterDataHandler) ListDistricts(c fiber.Ctx) error {
	var cityID *uuid.UUID
	if cid := c.Query("city_id"); cid != "" {
		id, err := uuid.Parse(cid)
		if err != nil {
			return response.Error(c, fiber.StatusBadRequest, "Invalid city_id", nil)
		}
		cityID = &id
	}
	list, err := h.masterDataRepo.ListDistricts(c.Context(), cityID)
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Districts retrieved", list)
}

func (h *MasterDataHandler) CreateDistrict(c fiber.Ctx) error {
	var body struct {
		CityID uuid.UUID `json:"city_id" validate:"required"`
		Name   string    `json:"name" validate:"required,min=2,max=255"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	if _, err := h.masterDataRepo.FindCityByID(c.Context(), body.CityID); err != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "city_id not found", nil)
	}
	d := &entity.District{CityID: body.CityID, Name: body.Name}
	if err := h.masterDataRepo.CreateDistrict(c.Context(), d); err != nil {
		return duplicateOrInternal(err)
	}
	return response.Success(c, fiber.StatusCreated, "District created", d)
}

func (h *MasterDataHandler) UpdateDistrict(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid district ID", nil)
	}
	var body masterBody
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	d, err := h.masterDataRepo.FindDistrictByID(c.Context(), id)
	if err != nil {
		return err
	}
	d.Name = body.Name
	if err := h.masterDataRepo.UpdateDistrict(c.Context(), d); err != nil {
		return duplicateOrInternal(err)
	}
	return response.Success(c, fiber.StatusOK, "District updated", d)
}

func (h *MasterDataHandler) DeleteDistrict(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid district ID", nil)
	}
	if _, err := h.masterDataRepo.FindDistrictByID(c.Context(), id); err != nil {
		return err
	}
	if err := h.masterDataRepo.DeleteDistrict(c.Context(), id); err != nil {
		return referencedOrInternal(err)
	}
	return response.Success(c, fiber.StatusOK, "District deleted", nil)
}

// --- Facilities ---

type facilityBody struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

func (h *MasterDataHandler) ListFacilities(c fiber.Ctx) error {
	list, err := h.masterDataRepo.ListFacilities(c.Context())
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Facilities retrieved", list)
}

func (h *MasterDataHandler) CreateFacility(c fiber.Ctx) error {
	var body facilityBody
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	f := &entity.Facility{Name: body.Name}
	if err := h.masterDataRepo.CreateFacility(c.Context(), f); err != nil {
		return duplicateOrInternal(err)
	}
	return response.Success(c, fiber.StatusCreated, "Facility created", f)
}

func (h *MasterDataHandler) UpdateFacility(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid facility ID", nil)
	}
	var body facilityBody
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(body); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	f, err := h.masterDataRepo.FindFacilityByID(c.Context(), id)
	if err != nil {
		return err
	}
	f.Name = body.Name
	if err := h.masterDataRepo.UpdateFacility(c.Context(), f); err != nil {
		return duplicateOrInternal(err)
	}
	return response.Success(c, fiber.StatusOK, "Facility updated", f)
}

func (h *MasterDataHandler) DeleteFacility(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid facility ID", nil)
	}
	if _, err := h.masterDataRepo.FindFacilityByID(c.Context(), id); err != nil {
		return err
	}
	if err := h.masterDataRepo.DeleteFacility(c.Context(), id); err != nil {
		return referencedOrInternal(err)
	}
	return response.Success(c, fiber.StatusOK, "Facility deleted", nil)
}
