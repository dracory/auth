package types

// PasswordStrengthConfig defines configurable rules for password strength.
type PasswordStrengthConfig struct {
	MinLength         int
	RequireUppercase  bool
	RequireLowercase  bool
	RequireDigit      bool
	RequireSpecial    bool
	ForbidCommonWords bool
}
