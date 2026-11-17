package service

import (
	"testing"

	"github.com/google/uuid"
)

func TestGenerateAccessToken(t *testing.T) {
	svc := NewJWTService("test-secret-key-12345", 3600, 604800)
	userID := uuid.New()

	token, err := svc.GenerateAccessToken(userID, "admin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestValidateAccessToken(t *testing.T) {
	svc := NewJWTService("test-secret-key-12345", 3600, 604800)
	userID := uuid.New()
	role := "sales"

	token, err := svc.GenerateAccessToken(userID, role)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected user_id %s, got %s", userID, claims.UserID)
	}
	if claims.Role != role {
		t.Errorf("expected role %s, got %s", role, claims.Role)
	}
}

func TestValidateAccessTokenInvalidSecret(t *testing.T) {
	svc1 := NewJWTService("secret-1", 3600, 604800)
	svc2 := NewJWTService("secret-2", 3600, 604800)

	token, _ := svc1.GenerateAccessToken(uuid.New(), "admin")

	_, err := svc2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for token signed with different secret")
	}
}

func TestValidateAccessTokenInvalidToken(t *testing.T) {
	svc := NewJWTService("test-secret", 3600, 604800)

	_, err := svc.ValidateAccessToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	svc := NewJWTService("test-secret", 3600, 604800)

	rawToken, tokenHash, expiresAt, err := svc.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rawToken == "" {
		t.Fatal("expected non-empty raw token")
	}
	if tokenHash == "" {
		t.Fatal("expected non-empty token hash")
	}
	if expiresAt.IsZero() {
		t.Fatal("expected non-zero expiration")
	}

	// Verify hash matches
	if svc.HashToken(rawToken) != tokenHash {
		t.Fatal("hash of raw token should match returned hash")
	}
}

func TestHashTokenConsistency(t *testing.T) {
	svc := NewJWTService("test-secret", 3600, 604800)

	token := "my-refresh-token"
	hash1 := svc.HashToken(token)
	hash2 := svc.HashToken(token)

	if hash1 != hash2 {
		t.Fatal("same token should produce same hash")
	}

	differentHash := svc.HashToken("different-token")
	if hash1 == differentHash {
		t.Fatal("different tokens should produce different hashes")
	}
}

func TestAccessExpiration(t *testing.T) {
	svc := NewJWTService("test-secret", 3600, 604800)
	if svc.AccessExpiration() != 3600 {
		t.Errorf("expected 3600, got %d", svc.AccessExpiration())
	}
}
