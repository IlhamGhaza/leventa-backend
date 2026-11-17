package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/leventa-backend/internal/usecase"
	"github.com/leventa-backend/pkg/response"
	"github.com/leventa-backend/pkg/validator"
)

// AuthHandler handles authentication HTTP requests.
type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
	userUsecase *usecase.UserUsecase
	validator   *validator.Validator
}

// NewAuthHandler creates a new auth handler.
func NewAuthHandler(authUsecase *usecase.AuthUsecase, userUsecase *usecase.UserUsecase, validator *validator.Validator) *AuthHandler {
	return &AuthHandler{authUsecase: authUsecase, userUsecase: userUsecase, validator: validator}
}

// Login handles user login.
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var input usecase.LoginInput
	if err := c.Bind().Body(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if errs := h.validator.Validate(input); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}

	result, err := h.authUsecase.Login(c.Context(), input)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, "Login successful", result)
}

// Refresh handles token refresh.
func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := c.Bind().Body(&body); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if body.RefreshToken == "" {
		return response.Error(c, fiber.StatusBadRequest, "Refresh token is required", nil)
	}

	result, err := h.authUsecase.RefreshToken(c.Context(), body.RefreshToken)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, "Token refreshed", result)
}

// Logout handles user logout.
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	if err := h.authUsecase.Logout(c.Context(), userID); err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, "Logged out successfully", nil)
}

// Me returns the authenticated user's profile.
func (h *AuthHandler) Me(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uuid.UUID)

	user, err := h.authUsecase.GetMe(c.Context(), userID)
	if err != nil {
		return err
	}

	return response.Success(c, fiber.StatusOK, "User profile", user)
}

// UpdateMe changes the authenticated user's own WhatsApp number. The user is
// taken from the verified token; the body carries only whatsapp_number.
func (h *AuthHandler) UpdateMe(c fiber.Ctx) error {
	var input usecase.UpdateOwnProfileInput
	if err := c.Bind().Body(&input); err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if errs := h.validator.Validate(input); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	user, err := h.userUsecase.UpdateOwnWhatsApp(c.Context(), getUserID(c), *input.WhatsAppNumber, c.IP(), c.Get("User-Agent"))
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Profile updated", user)
}
