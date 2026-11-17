package middleware

import (
	"errors"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/leventa-backend/pkg/errs"
	"github.com/leventa-backend/pkg/logger"
)

// A 500 must reach the log with its real cause and the request id, while the
// client only sees the generic message.
func TestInternalErrorIsLoggedWithCauseButHiddenFromClient(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "app.log")
	logger.Init("production", logFile)
	t.Cleanup(logger.Close)

	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Use(RequestID())
	app.Use(Logger())
	app.Get("/boom", func(c fiber.Ctx) error {
		return errs.Internal(errors.New("pq: connection refused (db down)"))
	})

	resp, err := app.Test(httptest.NewRequest("GET", "/boom", nil))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 500 {
		t.Fatalf("status %d, want 500", resp.StatusCode)
	}
	if strings.Contains(string(body), "db down") || strings.Contains(string(body), "connection") {
		t.Errorf("cause leaked to the client: %s", body)
	}
	if !strings.Contains(string(body), "Internal server error") {
		t.Errorf("unexpected body: %s", body)
	}

	raw, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatal(err)
	}
	logs := string(raw)
	if !strings.Contains(logs, "db down") {
		t.Errorf("cause missing from the log: %s", logs)
	}
	if !strings.Contains(logs, `"status":500`) {
		t.Errorf("request line should carry status 500: %s", logs)
	}
	if !strings.Contains(logs, `"request_id":"`) || strings.Contains(logs, `"request_id":""`) {
		t.Errorf("request id missing from the log: %s", logs)
	}
}
