package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/leventa-backend/pkg/errs"
)

// RequireRole creates a role-based authorization middleware.
func RequireRole(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals("user_role").(string)
		if !ok || userRole == "" {
			return errs.ErrUnauthorized
		}

		for _, role := range roles {
			if userRole == role {
				return c.Next()
			}
		}

		return errs.Forbidden("You do not have permission to access this resource")
	}
}
