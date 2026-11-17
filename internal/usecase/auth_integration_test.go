package usecase_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	pgRepo "github.com/leventa-backend/internal/repository/postgres"
	"github.com/leventa-backend/internal/service"
	"github.com/leventa-backend/internal/usecase"
)

func TestAuthFlow(t *testing.T) {
	e := setup(t)
	uc := usecase.NewAuthUsecase(pgRepo.NewUserRepository(e.db), pgRepo.NewRefreshTokenRepository(e.db),
		service.NewJWTService("test-secret-test-secret-test-secret", 3600, 3600), usecase.AuthPolicy{MaxLoginAttempts: 3, LockoutDuration: time.Second, MaxActiveSessions: 2})

	email := uuid.NewString() + "@test.local"
	if _, err := e.createAccount(t, "Reg User", email, "secret123"); err != nil {
		t.Fatal(err)
	}
	reg, err := uc.Login(ctx(), usecase.LoginInput{Email: email, Password: "secret123"})
	if err != nil {
		t.Fatal(err)
	}
	if reg.User.Role != entity.RoleSales {
		t.Fatalf("created role = %s, want sales", reg.User.Role)
	}
	if strings.Contains(mustJSON(t, reg), "password") {
		t.Error("response exposes password data")
	}
	// Refresh tokens are stored hashed.
	var raw int64
	e.db.Model(&entity.RefreshToken{}).Where("token_hash = ?", reg.RefreshToken).Count(&raw)
	if raw != 0 {
		t.Error("raw refresh token stored in database")
	}

	if _, err := uc.Login(ctx(), usecase.LoginInput{Email: email, Password: "wrong"}); statusOf(err) != 401 {
		t.Errorf("wrong password: %d, want 401", statusOf(err))
	}
	login, err := uc.Login(ctx(), usecase.LoginInput{Email: email, Password: "secret123"})
	if err != nil {
		t.Fatal(err)
	}

	// Rotation: new pair, old token dead; reusing it kills the whole session family.
	rot, err := uc.RefreshToken(ctx(), login.RefreshToken)
	if err != nil || rot.RefreshToken == login.RefreshToken {
		t.Fatalf("rotation: %v", err)
	}
	if _, err := uc.RefreshToken(ctx(), login.RefreshToken); statusOf(err) != 401 {
		t.Errorf("reuse of rotated token: %d, want 401", statusOf(err))
	}
	if _, err := uc.RefreshToken(ctx(), rot.RefreshToken); statusOf(err) != 401 {
		t.Errorf("token issued before theft detection should be revoked: %d", statusOf(err))
	}

	// Logout revokes refresh tokens.
	l2, _ := uc.Login(ctx(), usecase.LoginInput{Email: email, Password: "secret123"})
	if err := uc.Logout(ctx(), l2.User.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.RefreshToken(ctx(), l2.RefreshToken); statusOf(err) != 401 {
		t.Errorf("refresh after logout: %d, want 401", statusOf(err))
	}

	// Inactive accounts cannot log in.
	e.db.Model(&entity.User{}).Where("email = ?", email).Update("status", "inactive")
	if _, err := uc.Login(ctx(), usecase.LoginInput{Email: email, Password: "secret123"}); statusOf(err) != 403 {
		t.Errorf("inactive login: %d, want 403", statusOf(err))
	}
}

func newAuthUC(e *env, attempts int, lock time.Duration, sessions int) *usecase.AuthUsecase {
	return usecase.NewAuthUsecase(pgRepo.NewUserRepository(e.db), pgRepo.NewRefreshTokenRepository(e.db),
		service.NewJWTService("test-secret-test-secret-test-secret", 3600, 3600),
		usecase.AuthPolicy{MaxLoginAttempts: attempts, LockoutDuration: lock, MaxActiveSessions: sessions})
}

func TestPasswordPolicyAndEmailNormalization(t *testing.T) {
	e := setup(t)
	uc := newAuthUC(e, 3, time.Second, 5)
	email := strings.ToUpper(uuid.NewString()) + "@Test.Local"
	for _, pw := range []string{"short1", "onlyletters", "12345678", "password1", strings.Repeat("a1", 40)} {
		if _, err := e.createAccount(t, "Weak User", email, pw); statusOf(err) != 422 {
			t.Errorf("password %q accepted (status %d)", pw, statusOf(err))
		}
	}
	if _, err := e.createAccount(t, "Good User", email, "goodpass1"); err != nil {
		t.Fatal(err)
	}
	// Same email in another case is the same account.
	if _, err := e.createAccount(t, "Dup User", strings.ToLower(email), "goodpass1"); statusOf(err) != 409 {
		t.Errorf("case-variant duplicate: %d, want 409", statusOf(err))
	}
	if _, err := uc.Login(ctx(), usecase.LoginInput{Email: strings.ToLower(email), Password: "goodpass1"}); err != nil {
		t.Errorf("login with lower-cased email: %v", err)
	}
}

func TestLoginLockout(t *testing.T) {
	e := setup(t)
	uc := newAuthUC(e, 3, 1500*time.Millisecond, 5)
	email := uuid.NewString() + "@test.local"
	if _, err := e.createAccount(t, "Lock User", email, "goodpass1"); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Login(ctx(), usecase.LoginInput{Email: "nobody-" + email, Password: "whatever1"}); statusOf(err) != 401 {
		t.Errorf("unknown email: %d, want 401", statusOf(err))
	}
	for i := 0; i < 3; i++ {
		if _, err := uc.Login(ctx(), usecase.LoginInput{Email: email, Password: "wrongpass1"}); statusOf(err) != 401 {
			t.Fatalf("attempt %d: %d, want 401", i, statusOf(err))
		}
	}
	// Locked: even the correct password is refused.
	if _, err := uc.Login(ctx(), usecase.LoginInput{Email: email, Password: "goodpass1"}); statusOf(err) != 429 {
		t.Errorf("locked account: %d, want 429", statusOf(err))
	}
	time.Sleep(1700 * time.Millisecond)
	if _, err := uc.Login(ctx(), usecase.LoginInput{Email: email, Password: "goodpass1"}); err != nil {
		t.Errorf("after lock expiry: %v", err)
	}
	// A successful login resets the counter: two more failures must not lock.
	for i := 0; i < 2; i++ {
		_, _ = uc.Login(ctx(), usecase.LoginInput{Email: email, Password: "wrongpass1"})
	}
	if _, err := uc.Login(ctx(), usecase.LoginInput{Email: email, Password: "goodpass1"}); err != nil {
		t.Errorf("counter not reset by success: %v", err)
	}
}

func TestSessionCapAndDeactivationRevokesTokens(t *testing.T) {
	e := setup(t)
	uc := newAuthUC(e, 5, time.Minute, 2)
	email := uuid.NewString() + "@test.local"
	if _, err := e.createAccount(t, "Session User", email, "goodpass1"); err != nil {
		t.Fatal(err)
	}
	var first, last *usecase.AuthResponse
	for i := 0; i < 3; i++ {
		r, err := uc.Login(ctx(), usecase.LoginInput{Email: email, Password: "goodpass1"})
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = r
		}
		last = r
		time.Sleep(20 * time.Millisecond)
	}
	var active int64
	e.db.Model(&entity.RefreshToken{}).Where("user_id = ? AND revoked_at IS NULL", last.User.ID).Count(&active)
	if active > 2 {
		t.Errorf("active sessions = %d, want <= 2", active)
	}
	if _, err := uc.RefreshToken(ctx(), first.RefreshToken); statusOf(err) != 401 {
		t.Errorf("oldest session should be revoked: %d", statusOf(err))
	}

	// Admin deactivation revokes the refresh tokens.
	admin := e.user(t, entity.RoleAdmin, "active")
	uu := usecase.NewUserUsecase(pgRepo.NewUserRepository(e.db), pgRepo.NewRefreshTokenRepository(e.db), e.audit())
	r, _ := uc.Login(ctx(), usecase.LoginInput{Email: email, Password: "goodpass1"})
	if err := uu.UpdateStatus(ctx(), r.User.ID, entity.UserStatusInactive, admin.ID, "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.RefreshToken(ctx(), r.RefreshToken); statusOf(err) != 401 {
		t.Errorf("refresh after deactivation: %d, want 401", statusOf(err))
	}
	// Self role change is refused.
	if err := uu.UpdateRole(ctx(), admin.ID, entity.RoleSales, admin.ID, "", ""); statusOf(err) != 403 {
		t.Errorf("self role change: %d, want 403", statusOf(err))
	}
}

// createAccount creates a marketing account the way Admin does (there is no self-registration).
func (e *env) createAccount(t *testing.T, name, email, password string) (*entity.User, error) {
	t.Helper()
	admin := e.user(t, entity.RoleAdmin, "active")
	uu := usecase.NewUserUsecase(pgRepo.NewUserRepository(e.db), pgRepo.NewRefreshTokenRepository(e.db), e.audit())
	return uu.CreateUser(ctx(), usecase.CreateUserInput{Name: name, Email: email, Password: password, Role: entity.RoleSales}, admin.ID, "", "")
}
