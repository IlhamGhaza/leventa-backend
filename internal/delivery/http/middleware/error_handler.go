package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/leventa-backend/pkg/errs"
	"github.com/leventa-backend/pkg/logger"
)

// ErrorHandler is a custom Fiber error handler that maps domain errors to HTTP responses.
// Server-side failures (5xx) are logged with their underlying cause and the
// request id; clients only get a generic message.
func ErrorHandler(c fiber.Ctx, err error) error {
	// Check for AppError
	if appErr, ok := errs.IsAppError(err); ok {
		if appErr.Code >= fiber.StatusInternalServerError {
			logFailure(c, appErr.Cause, appErr.Message)
		}
		return c.Status(appErr.Code).JSON(fiber.Map{
			"success": false,
			"message": appErr.Message,
		})
	}

	// Check for Fiber error
	if fiberErr, ok := err.(*fiber.Error); ok {
		if fiberErr.Code >= fiber.StatusInternalServerError {
			logFailure(c, fiberErr, fiberErr.Message)
		}
		return c.Status(fiberErr.Code).JSON(fiber.Map{
			"success": false,
			"message": fiberErr.Message,
		})
	}

	// Unknown error: log it but don't expose internals
	logFailure(c, err, "Unhandled error")

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"message": "Internal server error",
	})
}

func logFailure(c fiber.Ctx, cause error, msg string) {
	e := logger.Error().
		Str("request_id", localString(c, "request_id")).
		Str("method", c.Method()).
		Str("path", c.Path()).
		Str("user_id", localString(c, "user_id")).
		Str("role", localString(c, "user_role"))
	if cause != nil {
		e = e.Err(cause)
	}
	e.Msg(msg)
}
