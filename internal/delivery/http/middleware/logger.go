package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/leventa-backend/pkg/errs"
	"github.com/leventa-backend/pkg/logger"
)

// localString reads a request-scoped value as a string ("" when absent).
func localString(c fiber.Ctx, key string) string {
	switch v := c.Locals(key).(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	}
	return ""
}

// Logger logs each HTTP request with method, path, status, latency, request id
// and (when authenticated) the user id and role. The status reflects errors
// returned by handlers, which Fiber's error handler only writes later.
func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		status := c.Response().StatusCode()
		if err != nil {
			if appErr, ok := errs.IsAppError(err); ok {
				status = appErr.Code
			} else if fe, ok := err.(*fiber.Error); ok {
				status = fe.Code
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		event := logger.Info()
		if status >= 500 {
			event = logger.Error()
		} else if status >= 400 {
			event = logger.Warn()
		}

		event.
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Dur("latency", time.Since(start)).
			Str("ip", c.IP()).
			Str("request_id", localString(c, "request_id")).
			Str("user_id", localString(c, "user_id")).
			Str("role", localString(c, "user_role")).
			Msg("HTTP Request")

		return err
	}
}
