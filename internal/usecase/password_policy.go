package usecase

import (
	"strings"
	"unicode"

	"github.com/leventa-backend/pkg/errs"
)

// Password length bounds. bcrypt only uses the first 72 bytes, so longer input is rejected
// instead of being silently truncated.
const (
	MinPasswordLength = 8
	MaxPasswordBytes  = 72
)

var commonPasswords = map[string]bool{
	"password": true, "password1": true, "password123": true, "12345678": true, "123456789": true,
	"qwerty123": true, "11111111": true, "iloveyou1": true, "admin123": false,
}

// ValidatePasswordStrength enforces length (8..72 bytes), at least one letter and one digit,
// and rejects a few extremely common passwords.
func ValidatePasswordStrength(pw string) error {
	if len([]rune(pw)) < MinPasswordLength {
		return errs.Validation("Password must be at least 8 characters")
	}
	if len(pw) > MaxPasswordBytes {
		return errs.Validation("Password must be at most 72 bytes")
	}
	var letter, digit bool
	for _, r := range pw {
		switch {
		case unicode.IsLetter(r):
			letter = true
		case unicode.IsDigit(r):
			digit = true
		}
	}
	if !letter || !digit {
		return errs.Validation("Password must contain at least one letter and one digit")
	}
	if commonPasswords[strings.ToLower(pw)] {
		return errs.Validation("Password is too common")
	}
	return nil
}

// normalizeEmail lower-cases and trims an email so that case variants cannot create duplicate accounts.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
