package response

import (
	"github.com/gofiber/fiber/v3"
)

// PaginationMeta holds pagination metadata.
type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
}

// Success sends a successful JSON response.
func Success(c fiber.Ctx, statusCode int, message string, data interface{}) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// SuccessWithPagination sends a successful JSON response with pagination metadata.
func SuccessWithPagination(c fiber.Ctx, message string, data interface{}, meta PaginationMeta) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    meta,
	})
}

// Error sends an error JSON response.
func Error(c fiber.Ctx, statusCode int, message string, errors interface{}) error {
	resp := fiber.Map{
		"success": false,
		"message": message,
	}
	if errors != nil {
		resp["errors"] = errors
	}
	return c.Status(statusCode).JSON(resp)
}
