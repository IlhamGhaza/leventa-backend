package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// RequestIDHeader is the header carrying the request correlation ID.
const RequestIDHeader = "X-Request-ID"

// RequestID adds a unique request ID to every request.
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		requestID := c.Get(RequestIDHeader)
		// Never trust arbitrary client input in logs: accept only short, plain IDs.
		if requestID == "" || len(requestID) > 64 || strings.ContainsAny(requestID, "\r\n\t ") {
			requestID = uuid.New().String()
		}
		c.Set(RequestIDHeader, requestID)
		c.Locals("request_id", requestID)
		return c.Next()
	}
}
