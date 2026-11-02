package types

// UserCreationPolicy selects when the user record is created relative to
// registration for verified-email flows (OTP, magic link, AuthKnight).
type UserCreationPolicy string

const (
	// CreateUserOnAuth is the default: the user record is created at
	// verification time (UserFindByEmailOrCreate), a session is issued, and
	// the RegistrationComplete gate sends incomplete profiles to the
	// registration-completion form. Verified-but-incomplete rows may exist.
	CreateUserOnAuth UserCreationPolicy = "on_auth"

	// CreateUserOnRegister defers the user record until the registration
	// form submits: verification issues a short-lived verified-email token,
	// the register form renders the email read-only, and a single insert
	// creates a verified + complete user. No rows exist before the form
	// submits.
	//
	// Fail-open note: when the policy cannot be enforced (no UserExists
	// hook configured, existence check error, or token store missing) the
	// flows fall back to session-first behavior — availability over
	// strictness — and log a Warn. Hosts whose threat model requires
	// fail-closed should ensure GetFuncUserExists (or the otp
	// find-by-email fallback) is configured and monitor the warnings.
	CreateUserOnRegister UserCreationPolicy = "on_register"
)
