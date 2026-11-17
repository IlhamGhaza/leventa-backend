package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
)

// CORS configures Cross-Origin Resource Sharing with an explicit origin
// allowlist (comma separated). An empty list disables cross-origin access:
// Fiber v3 treats an empty AllowOrigins as "allow all", so deny explicitly.
func CORS(allowedOrigins string) fiber.Handler {
	cfg := cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
	}
	var origins []string
	for _, o := range strings.Split(allowedOrigins, ",") {
		if o = strings.TrimSpace(o); o != "" {
			origins = append(origins, o)
		}
	}
	if len(origins) == 0 {
		cfg.AllowOriginsFunc = func(string) bool { return false }
	} else {
		cfg.AllowOrigins = origins
	}
	return cors.New(cfg)
}

// SecurityHeaders sets standard security response headers. In production the app sits
// behind a TLS-terminating proxy, so HSTS is added explicitly.
func SecurityHeaders(production bool) fiber.Handler {
	h := helmet.New()
	return func(c fiber.Ctx) error {
		if production {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		return h(c)
	}
}
