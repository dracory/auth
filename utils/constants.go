package utils

// Temporary-key prefixes — keep every key namespace here so the flat
// keyspace shared by the temp store cannot collide silently.
const (
	// LoginCodeKeyPrefix prefixes the temporary keys under which passwordless
	// login codes are stored. The nonce — not the code — is the storage key so
	// that guessing a code does not reveal or consume it.
	LoginCodeKeyPrefix = "logincode:"

	// RegisterCodeKeyPrefix prefixes the temporary keys under which registration
	// code payloads are stored.
	RegisterCodeKeyPrefix = "registercode:"

	// LoginCodeAttemptsKeyPrefix prefixes the per-nonce failed-attempt counters
	// for passwordless login codes.
	LoginCodeAttemptsKeyPrefix = "logincode_attempts:"

	// RegisterCodeAttemptsKeyPrefix prefixes the per-nonce failed-attempt
	// counters for registration codes.
	RegisterCodeAttemptsKeyPrefix = "registercode_attempts:"

	// PasswordResetTokenKeyPrefix namespaces password reset tokens in the
	// temporary key store, keeping them out of the flat keyspace shared with
	// other temp keys.
	PasswordResetTokenKeyPrefix = "pwdreset:"

	// MagicLinkKeyPrefix prefixes the temporary keys under which magic-link
	// payloads are stored. The key IS the token — it is 128+ bits of entropy,
	// single-use, and expires with the payload.
	MagicLinkKeyPrefix = "magiclink:token:"

	// DeliveryKeyPrefix prefixes the temporary keys under which queued email
	// deliveries are stored. Queue payloads contain only the opaque key — the
	// secret (code or link) lives solely in the TTL-bound temporary-key store.
	DeliveryKeyPrefix = "delivery:"

	// EmailSendKeyPrefix prefixes the per-email send-throttle counters.
	EmailSendKeyPrefix = "code_sent:"
)

// Quotas and TTLs — the tunable limits behind the verification and
// send-throttle flows.
const (
	// MagicLinkExpiresSeconds is the default magic-link TTL (15 minutes).
	MagicLinkExpiresSeconds = 900

	// MaxCodeAttempts is the number of verification attempts allowed per nonce
	// before the code is invalidated and a new one must be requested.
	MaxCodeAttempts = 5

	// DefaultEmailSendMaxSends is the default number of code-send requests
	// allowed per email address within one throttle window.
	DefaultEmailSendMaxSends = 3

	// DefaultEmailSendWindowSeconds is the default length of the per-email
	// send-throttle sliding window (one hour).
	DefaultEmailSendWindowSeconds = 3600

	// MaxEmailSends is kept for backward compatibility; new code should use
	// DefaultEmailSendMaxSends.
	MaxEmailSends = DefaultEmailSendMaxSends
)

// Remember-me markers.
const (
	// RememberSessionMarker is the session value marking remember-me sessions
	// so a remember key cannot act as a normal auth credential and vice versa.
	RememberSessionMarker = "remember"

	// DefaultRememberCookieName is the cookie holding the remember token.
	DefaultRememberCookieName = "remember_token"
)
