package entity

import (
	"testing"
)

func TestUserSetPassword(t *testing.T) {
	user := &User{}
	err := user.SetPassword("password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.PasswordHash == "" {
		t.Fatal("expected password hash to be set")
	}
	if user.PasswordHash == "password123" {
		t.Fatal("password should be hashed, not stored as plain text")
	}
}

func TestUserCheckPassword(t *testing.T) {
	user := &User{}
	_ = user.SetPassword("correctpassword")

	if !user.CheckPassword("correctpassword") {
		t.Fatal("expected CheckPassword to return true for correct password")
	}
	if user.CheckPassword("wrongpassword") {
		t.Fatal("expected CheckPassword to return false for wrong password")
	}
}

func TestUserIsAdmin(t *testing.T) {
	user := &User{Role: RoleAdmin}
	if !user.IsAdmin() {
		t.Fatal("expected IsAdmin to return true for admin role")
	}

	user.Role = RoleSales
	if user.IsAdmin() {
		t.Fatal("expected IsAdmin to return false for sales role")
	}
}

func TestUserIsSales(t *testing.T) {
	user := &User{Role: RoleSales}
	if !user.IsSales() {
		t.Fatal("expected IsSales to return true for sales role")
	}
}

func TestUserIsActive(t *testing.T) {
	user := &User{Status: UserStatusActive}
	if !user.IsActive() {
		t.Fatal("expected IsActive to return true for active status")
	}

	user.Status = UserStatusInactive
	if user.IsActive() {
		t.Fatal("expected IsActive to return false for inactive status")
	}
}

func TestRefreshTokenIsExpired(t *testing.T) {
	token := &RefreshToken{}
	// Zero time means expired
	if !token.IsExpired() {
		t.Fatal("expected zero-time token to be expired")
	}
}

func TestCalculateCommission(t *testing.T) {
	tests := []struct {
		name       string
		dealPrice  float64
		percentage float64
		expected   float64
	}{
		{"standard 2.5%", 1000000000, 2.5, 25000000},
		{"custom 5%", 500000000, 5.0, 25000000},
		{"zero price", 0, 2.5, 0},
		{"zero percentage", 1000000000, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateCommission(tt.dealPrice, tt.percentage)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestValidPropertyTypes(t *testing.T) {
	types := ValidPropertyTypes()
	if len(types) != 4 {
		t.Fatalf("expected 4 property types, got %d", len(types))
	}
}

func TestValidPropertyStatuses(t *testing.T) {
	statuses := ValidPropertyStatuses()
	if len(statuses) != 7 {
		t.Fatalf("expected 7 property statuses, got %d", len(statuses))
	}
}
