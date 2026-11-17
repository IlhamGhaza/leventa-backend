package usecase

import (
	"time"

	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/internal/service"
	"github.com/leventa-backend/pkg/errs"
)

// AuthUsecase handles authentication business logic.
type AuthUsecase struct {
	policy     AuthPolicy
	userRepo   repository.UserRepository
	tokenRepo  repository.RefreshTokenRepository
	jwtService *service.JWTService
}

// NewAuthUsecase creates a new auth usecase.
// AuthPolicy configures brute-force protection and session limits.
type AuthPolicy struct {
	MaxLoginAttempts  int
	LockoutDuration   time.Duration
	MaxActiveSessions int
}

// dummyHash is compared against when the email is unknown (timing equalization).
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("timing-equalization-dummy"), bcrypt.DefaultCost)

func NewAuthUsecase(
	userRepo repository.UserRepository,
	tokenRepo repository.RefreshTokenRepository,
	jwtService *service.JWTService,
	policy AuthPolicy,
) *AuthUsecase {
	if policy.MaxLoginAttempts <= 0 {
		policy.MaxLoginAttempts = 5
	}
	if policy.LockoutDuration <= 0 {
		policy.LockoutDuration = 15 * time.Minute
	}
	if policy.MaxActiveSessions <= 0 {
		policy.MaxActiveSessions = 5
	}
	return &AuthUsecase{
		policy:     policy,
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtService: jwtService,
	}
}

// LoginInput holds the login request data.
type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse holds the authentication response data.
type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	User         entity.User `json:"user"`
}

// Login authenticates a user by email and password.
func (uc *AuthUsecase) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	invalid := errs.New(401, "Invalid email or password")

	user, err := uc.userRepo.FindByEmail(ctx, normalizeEmail(input.Email))
	if err != nil {
		// Burn the same CPU time as a real check so response time does not reveal unknown emails.
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(input.Password))
		return nil, invalid
	}

	if user.LockedUntil != nil {
		if time.Now().Before(*user.LockedUntil) {
			return nil, errs.New(429, "Too many failed login attempts. Try again later.")
		}
		// Lock expired: start counting from zero again.
		if err := uc.userRepo.ResetLoginFailures(ctx, user.ID); err != nil {
			return nil, errs.Internal(err)
		}
		user.FailedLoginAttempts, user.LockedUntil = 0, nil
	}

	if !user.CheckPassword(input.Password) {
		if err := uc.userRepo.RegisterLoginFailure(ctx, user.ID, uc.policy.MaxLoginAttempts, uc.policy.LockoutDuration); err != nil {
			return nil, errs.Internal(err)
		}
		return nil, invalid
	}

	if user.FailedLoginAttempts > 0 {
		_ = uc.userRepo.ResetLoginFailures(ctx, user.ID)
	}

	if !user.IsActive() {
		return nil, errs.Forbidden("Account is inactive")
	}

	return uc.generateAuthResponse(ctx, user)
}

// RefreshToken refreshes an access token using a refresh token.
func (uc *AuthUsecase) RefreshToken(ctx context.Context, rawToken string) (*AuthResponse, error) {
	tokenHash := uc.jwtService.HashToken(rawToken)

	storedToken, err := uc.tokenRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errs.ErrUnauthorized
	}

	if storedToken.IsRevoked() {
		// Reuse of a rotated/revoked token suggests theft: end every session of the user.
		_ = uc.tokenRepo.RevokeAllByUserID(ctx, storedToken.UserID)
		return nil, errs.ErrUnauthorized
	}
	if storedToken.IsExpired() {
		return nil, errs.ErrUnauthorized
	}

	// Revoke the old token (rotation); fails when a concurrent request already used it.
	if err := uc.tokenRepo.Revoke(ctx, storedToken.ID); err != nil {
		if _, ok := errs.IsAppError(err); ok {
			return nil, err
		}
		return nil, errs.Internal(err)
	}

	// Get user
	user, err := uc.userRepo.FindByID(ctx, storedToken.UserID)
	if err != nil {
		return nil, errs.ErrUnauthorized
	}

	if !user.IsActive() {
		return nil, errs.Forbidden("Account is inactive")
	}

	return uc.generateAuthResponse(ctx, user)
}

// Logout revokes all refresh tokens for a user.
func (uc *AuthUsecase) Logout(ctx context.Context, userID uuid.UUID) error {
	return uc.tokenRepo.RevokeAllByUserID(ctx, userID)
}

// GetMe returns the authenticated user's profile.
func (uc *AuthUsecase) GetMe(ctx context.Context, userID uuid.UUID) (*entity.User, error) {
	return uc.userRepo.FindByID(ctx, userID)
}

func (uc *AuthUsecase) generateAuthResponse(ctx context.Context, user *entity.User) (*AuthResponse, error) {
	accessToken, err := uc.jwtService.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, errs.Internal(err)
	}

	rawToken, tokenHash, expiresAt, err := uc.jwtService.GenerateRefreshToken()
	if err != nil {
		return nil, errs.Internal(err)
	}

	refreshToken := &entity.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}
	if err := uc.tokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, errs.Internal(err)
	}
	// Cap concurrent sessions per user; the oldest ones are revoked.
	_ = uc.tokenRepo.PruneActive(ctx, user.ID, uc.policy.MaxActiveSessions)

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: rawToken,
		ExpiresIn:    uc.jwtService.AccessExpiration(),
		User:         *user,
	}, nil
}
