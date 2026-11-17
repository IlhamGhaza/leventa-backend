package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/leventa-backend/pkg/logger"
)

func TestRateLimiterBlocksBurstPerIP(t *testing.T) {
	logger.Init("production", "")
	app := fiber.New()
	app.Use(RequestID())
	app.Get("/x", RateLimiter(3, 60, "test"), func(c fiber.Ctx) error { return c.SendString("ok") })

	for i := 1; i <= 3; i++ {
		resp, err := app.Test(httptest.NewRequest("GET", "/x", nil))
		if err != nil || resp.StatusCode != 200 {
			t.Fatalf("request %d: %v %v", i, err, resp.StatusCode)
		}
	}
	resp, err := app.Test(httptest.NewRequest("GET", "/x", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("4th request: %d, want 429", resp.StatusCode)
	}
	if resp.Header.Get("Retry-After") != "60" {
		t.Errorf("Retry-After = %q, want 60", resp.Header.Get("Retry-After"))
	}
}
