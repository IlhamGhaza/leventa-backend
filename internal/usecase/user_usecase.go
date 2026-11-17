package usecase

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/internal/service"
	"github.com/leventa-backend/pkg/errs"
)

// UserUsecase handles user management business logic (admin only).
type UserUsecase struct {
	userRepo     repository.UserRepository
	tokenRepo    repository.RefreshTokenRepository
	auditService *service.AuditService
}

// NewUserUsecase creates a new user usecase.
func NewUserUsecase(userRepo repository.UserRepository, tokenRepo repository.RefreshTokenRepository, auditService *service.AuditService) *UserUsecase {
	return &UserUsecase{userRepo: userRepo, tokenRepo: tokenRepo, auditService: auditService}
}

// CreateUserInput holds the data for admin user creation.
type CreateUserInput struct {
	Name           string `json:"name" validate:"required,min=2,max=255"`
	Email          string `json:"email" validate:"required,email"`
	Phone          string `json:"phone" validate:"omitempty,min=8,max=20"`
	WhatsAppNumber string `json:"whatsapp_number" validate:"omitempty,min=8,max=20"`
	Password       string `json:"password" validate:"required,min=8,max=72"`
	Role           string `json:"role" validate:"required,oneof=sales admin"`
}

// UpdateUserInput holds the data for admin user update.
type UpdateUserInput struct {
	Name           string `json:"name" validate:"required,min=2,max=255"`
	Email          string `json:"email" validate:"required,email"`
	Phone          string `json:"phone" validate:"omitempty,min=8,max=20"`
	WhatsAppNumber string `json:"whatsapp_number" validate:"omitempty,min=8,max=20"`
}

// CreateUser creates a new user (admin only, Rule 4).
func (uc *UserUsecase) CreateUser(ctx context.Context, input CreateUserInput, adminID uuid.UUID, ip, userAgent string) (*entity.User, error) {
	input.Email = normalizeEmail(input.Email)
	if err := ValidatePasswordStrength(input.Password); err != nil {
		return nil, err
	}
	existing, _ := uc.userRepo.FindByEmail(ctx, input.Email)
	if existing != nil {
		return nil, errs.Conflict("Email already registered")
	}

	wa, err := normalizeWA(input.WhatsAppNumber)
	if err != nil {
		return nil, err
	}
	user := &entity.User{
		Name:           input.Name,
		Email:          input.Email,
		Phone:          input.Phone,
		WhatsAppNumber: wa,
		Role:           input.Role,
		Status:         entity.UserStatusActive,
	}
	if err := user.SetPassword(input.Password); err != nil {
		return nil, errs.Internal(err)
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, errs.Internal(err)
	}

	uc.auditService.Log(&adminID, "created_user", "users", &user.ID,
		map[string]string{"role": input.Role}, ip, userAgent)

	return user, nil
}

// UpdateUser updates a user (admin only).
func (uc *UserUsecase) UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*entity.User, error) {
	input.Email = normalizeEmail(input.Email)
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check email uniqueness if changed
	if user.Email != input.Email {
		existing, _ := uc.userRepo.FindByEmail(ctx, input.Email)
		if existing != nil {
			return nil, errs.Conflict("Email already registered")
		}
	}

	user.Name = input.Name
	user.Email = input.Email
	user.Phone = input.Phone
	wa, err := normalizeWA(input.WhatsAppNumber)
	if err != nil {
		return nil, err
	}
	user.WhatsAppNumber = wa

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, errs.Internal(err)
	}

	return user, nil
}

// UpdateRole updates a user's role (admin only, Rule 14).
func (uc *UserUsecase) UpdateRole(ctx context.Context, id uuid.UUID, role string, adminID uuid.UUID, ip, userAgent string) error {
	if id == adminID {
		return errs.Forbidden("You cannot change your own role")
	}
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	oldRole := user.Role
	if err := uc.userRepo.UpdateRole(ctx, id, role); err != nil {
		return errs.Internal(err)
	}
	// Force a fresh login so the new role takes effect everywhere.
	_ = uc.tokenRepo.RevokeAllByUserID(ctx, id)

	uc.auditService.Log(&adminID, "changed_role", "users", &id,
		map[string]string{"old_role": oldRole, "new_role": role, "email": user.Email}, ip, userAgent)

	return nil
}

// UpdateStatus updates a user's status (admin only).
func (uc *UserUsecase) UpdateStatus(ctx context.Context, id uuid.UUID, status string, adminID uuid.UUID, ip, userAgent string) error {
	_, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.userRepo.UpdateStatus(ctx, id, status); err != nil {
		return errs.Internal(err)
	}
	if status != entity.UserStatusActive {
		_ = uc.tokenRepo.RevokeAllByUserID(ctx, id)
	}

	uc.auditService.Log(&adminID, "changed_status", "users", &id,
		map[string]string{"new_status": status}, ip, userAgent)

	return nil
}

// DeleteUser soft-deletes a user (admin only).
func (uc *UserUsecase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := uc.userRepo.Delete(ctx, id); err != nil {
		return err
	}
	_ = uc.tokenRepo.RevokeAllByUserID(ctx, id)
	return nil
}

// GetUser returns a user by ID.
func (uc *UserUsecase) GetUser(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return uc.userRepo.FindByID(ctx, id)
}

// ListUsers returns a paginated list of users.
func (uc *UserUsecase) ListUsers(ctx context.Context, page, limit int, search, role, status string) ([]entity.User, int64, error) {
	return uc.userRepo.List(ctx, page, limit, search, role, status)
}

// normalizeWA validates and normalizes a WhatsApp number; empty stays empty.
// UpdateOwnProfileInput is the self-service profile change. Only the
// WhatsApp number can be changed; an empty value clears it.
type UpdateOwnProfileInput struct {
	WhatsAppNumber *string `json:"whatsapp_number" validate:"required"`
}

// UpdateOwnWhatsApp lets an authenticated user change their own WhatsApp
// number (the public Marketing contact, D-012/D-023). The user ID always comes
// from the verified token, never from the request body.
func (uc *UserUsecase) UpdateOwnWhatsApp(ctx context.Context, userID uuid.UUID, raw string, ip, userAgent string) (*entity.User, error) {
	wa, err := normalizeWA(strings.TrimSpace(raw))
	if err != nil {
		return nil, err
	}
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.WhatsAppNumber == wa {
		return user, nil
	}
	if err := uc.userRepo.UpdateWhatsAppNumber(ctx, userID, wa); err != nil {
		return nil, errs.Internal(err)
	}
	uc.auditService.Log(&userID, "updated_own_whatsapp", "users", &userID, nil, ip, userAgent)
	user.WhatsAppNumber = wa
	return user, nil
}

func normalizeWA(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	n, ok := NormalizeWhatsApp(raw)
	if !ok {
		return "", errs.Validation("whatsapp_number is not a valid number")
	}
	return n, nil
}
