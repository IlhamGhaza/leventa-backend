package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/leventa-backend/pkg/logger"
)

// RateLimiter limits requests per client IP within a sliding window (in-memory,
// per process; no Redis). Behind a reverse proxy set TRUSTED_PROXIES so the
// real client IP is used. name identifies the limiter in the log.
func RateLimiter(max int, window int, name ...string) fiber.Handler {
	label := "default"
	if len(name) > 0 {
		label = name[0]
	}
	return limiter.New(limiter.Config{
		Max:               max,
		Expiration:        time.Duration(window) * time.Second,
		LimiterMiddleware: limiter.SlidingWindow{},
		KeyGenerator: func(c fiber.Ctx) string {
			return label + ":" + c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			logger.Warn().
				Str("limiter", label).
				Str("ip", c.IP()).
				Str("method", c.Method()).
				Str("path", c.Path()).
				Str("request_id", localString(c, "request_id")).
				Msg("Rate limit reached")
			c.Set(fiber.HeaderRetryAfter, strconv.Itoa(window))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"success": false,
				"message": "Too many requests. Please try again later.",
			})
		},
	})
}
