package middleware

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/service"
	"github.com/leventa-backend/pkg/errs"
)

// UserLookup loads the current account state. The role and status in the database,
// not the ones inside the token, decide access, so deactivation and role changes
// take effect immediately.
type UserLookup interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}

func authenticate(c fiber.Ctx, jwtService *service.JWTService, users UserLookup) bool {
	authHeader := c.Get("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return false
	}
	claims, err := jwtService.ValidateAccessToken(parts[1])
	if err != nil {
		return false
	}
	user, err := users.FindByID(c.Context(), claims.UserID)
	if err != nil || !user.IsActive() {
		return false
	}
	c.Locals("user_id", user.ID)
	c.Locals("user_role", user.Role)
	return true
}

// Auth requires a valid access token of an active account.
func Auth(jwtService *service.JWTService, users UserLookup) fiber.Handler {
	return func(c fiber.Ctx) error {
		if !authenticate(c, jwtService, users) {
			return errs.ErrUnauthorized
		}
		return c.Next()
	}
}

// OptionalAuth sets the actor when a valid token is present and never blocks the request.
func OptionalAuth(jwtService *service.JWTService, users UserLookup) fiber.Handler {
	return func(c fiber.Ctx) error {
		authenticate(c, jwtService, users)
		return c.Next()
	}
}
