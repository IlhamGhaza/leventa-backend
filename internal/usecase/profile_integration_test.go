package usecase_test

import (
	"testing"
	"time"

	"github.com/leventa-backend/internal/domain/entity"
	pgRepo "github.com/leventa-backend/internal/repository/postgres"
	"github.com/leventa-backend/internal/usecase"
)

func TestUpdateOwnWhatsApp(t *testing.T) {
	e := setup(t)
	uu := usecase.NewUserUsecase(pgRepo.NewUserRepository(e.db), pgRepo.NewRefreshTokenRepository(e.db), e.audit())
	me := e.user(t, entity.RoleSales, "active")
	other := e.user(t, entity.RoleSales, "active")

	// Lockout state must survive the profile update (single-column update).
	lockedUntil := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	e.db.Model(&entity.User{}).Where("id = ?", me.ID).
		Updates(map[string]any{"failed_login_attempts": 3, "locked_until": lockedUntil})

	u, err := uu.UpdateOwnWhatsApp(ctx(), me.ID, " 0812-3456-7890 ", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if u.WhatsAppNumber != "6281234567890" {
		t.Errorf("returned number = %q, want normalized 6281234567890", u.WhatsAppNumber)
	}

	var stored entity.User
	e.db.First(&stored, "id = ?", me.ID)
	if stored.WhatsAppNumber != "6281234567890" {
		t.Errorf("stored number = %q", stored.WhatsAppNumber)
	}
	if stored.Role != entity.RoleSales || stored.Status != "active" {
		t.Errorf("role/status changed: %s/%s", stored.Role, stored.Status)
	}
	if stored.FailedLoginAttempts != 3 || stored.LockedUntil == nil {
		t.Error("lockout state was overwritten by the profile update")
	}

	// Other users are untouched.
	var o entity.User
	e.db.First(&o, "id = ?", other.ID)
	if o.WhatsAppNumber != "6281200000099" {
		t.Errorf("other user's number changed: %q", o.WhatsAppNumber)
	}

	// Invalid number is a validation error and changes nothing.
	if _, err := uu.UpdateOwnWhatsApp(ctx(), me.ID, "12ab", "", ""); statusOf(err) != 422 {
		t.Errorf("invalid number: %d, want 422", statusOf(err))
	}

	// Empty value clears the number (contact then has no WhatsApp link).
	if _, err := uu.UpdateOwnWhatsApp(ctx(), me.ID, "", "", ""); err != nil {
		t.Fatal(err)
	}
	e.db.First(&stored, "id = ?", me.ID)
	if stored.WhatsAppNumber != "" {
		t.Errorf("number not cleared: %q", stored.WhatsAppNumber)
	}
}
