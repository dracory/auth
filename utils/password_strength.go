package utils

import (
	"errors"
	"strconv"
	"strings"
	"unicode"

	authtypes "github.com/dracory/auth/types"
)

// ValidatePasswordStrength validates the provided password against the
// supplied PasswordStrengthConfig. If cfg is nil, no checks are applied.
func ValidatePasswordStrength(password string, cfg *authtypes.PasswordStrengthConfig) error {
	if cfg == nil {
		return nil
	}

	if cfg.MinLength > 0 && len(password) < cfg.MinLength {
		return errors.New("password must be at least " + strconv.Itoa(cfg.MinLength) + " characters long")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if cfg.RequireUppercase && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if cfg.RequireLowercase && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if cfg.RequireDigit && !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if cfg.RequireSpecial && !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	if cfg.ForbidCommonWords {
		lower := strings.ToLower(password)
		common := []string{"password", "123456", "123456789", "qwerty", "admin", "letmein"}
		for _, w := range common {
			if lower == w {
				return errors.New("password is too common")
			}
		}
	}

	return nil
}
