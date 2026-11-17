package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the JWT token claims.
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations.
type JWTService struct {
	secret            string
	accessExpiration  time.Duration
	refreshExpiration time.Duration
}

// NewJWTService creates a new JWT service.
func NewJWTService(secret string, accessExp, refreshExp int) *JWTService {
	return &JWTService{
		secret:            secret,
		accessExpiration:  time.Duration(accessExp) * time.Second,
		refreshExpiration: time.Duration(refreshExp) * time.Second,
	}
}

// GenerateAccessToken creates a signed JWT access token.
func (s *JWTService) GenerateAccessToken(userID uuid.UUID, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "leventa-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

// ValidateAccessToken validates a JWT access token and returns the claims.
func (s *JWTService) ValidateAccessToken(tokenStr string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}

// GenerateRefreshToken creates a random refresh token and returns the raw token and its SHA-256 hash.
func (s *JWTService) GenerateRefreshToken() (rawToken string, tokenHash string, expiresAt time.Time, err error) {
	bytes := make([]byte, 32)
	if _, err = rand.Read(bytes); err != nil {
		return "", "", time.Time{}, err
	}

	rawToken = hex.EncodeToString(bytes)
	tokenHash = s.HashToken(rawToken)
	expiresAt = time.Now().Add(s.refreshExpiration)

	return rawToken, tokenHash, expiresAt, nil
}

// HashToken hashes a token using SHA-256.
func (s *JWTService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// AccessExpiration returns the access token expiration in seconds.
func (s *JWTService) AccessExpiration() int {
	return int(s.accessExpiration.Seconds())
}
