package middleware

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/service"
)

type fakeUsers struct{ u *entity.User }

func (f *fakeUsers) FindByID(_ context.Context, _ uuid.UUID) (*entity.User, error) {
	if f.u == nil {
		return nil, errors.New("not found")
	}
	return f.u, nil
}

func newTestApp(jwt *service.JWTService, users UserLookup) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler})
	app.Get("/me", Auth(jwt, users), func(c fiber.Ctx) error {
		return c.SendString(c.Locals("user_role").(string))
	})
	app.Get("/admin", Auth(jwt, users), RequireRole(entity.RoleAdmin), func(c fiber.Ctx) error { return c.SendString("ok") })
	return app
}

func call(t *testing.T, app *fiber.App, path, token string) (int, string) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func TestAuthUsesCurrentDatabaseState(t *testing.T) {
	jwt := service.NewJWTService("test-secret-test-secret-test-secret", 900, 3600)
	id := uuid.New()
	// Token was issued while the user was an admin.
	token, err := jwt.GenerateAccessToken(id, entity.RoleAdmin)
	if err != nil {
		t.Fatal(err)
	}
	users := &fakeUsers{u: &entity.User{ID: id, Role: entity.RoleAdmin, Status: entity.UserStatusActive}}
	app := newTestApp(jwt, users)

	if code, _ := call(t, app, "/admin", token); code != 200 {
		t.Fatalf("admin allowed: %d", code)
	}
	if code, _ := call(t, app, "/me", ""); code != 401 {
		t.Errorf("no token: %d, want 401", code)
	}
	if code, _ := call(t, app, "/me", "garbage"); code != 401 {
		t.Errorf("garbage token: %d, want 401", code)
	}

	// Role downgraded in the database: the old token must lose admin access immediately.
	users.u = &entity.User{ID: id, Role: entity.RoleSales, Status: entity.UserStatusActive}
	if code, _ := call(t, app, "/admin", token); code != 403 {
		t.Errorf("downgraded role still admin: %d, want 403", code)
	}
	if code, role := call(t, app, "/me", token); code != 200 || role != entity.RoleSales {
		t.Errorf("role should come from DB: %d %q", code, role)
	}

	// Deactivated or deleted users are rejected.
	users.u = &entity.User{ID: id, Role: entity.RoleAdmin, Status: entity.UserStatusInactive}
	if code, _ := call(t, app, "/me", token); code != 401 {
		t.Errorf("inactive user: %d, want 401", code)
	}
	users.u = nil
	if code, _ := call(t, app, "/me", token); code != 401 {
		t.Errorf("deleted user: %d, want 401", code)
	}
}
