package types

// ObservabilityHooks allows applications to record metrics and tracing spans
// for authentication events using their preferred provider (Prometheus,
// Datadog, OpenTelemetry, etc.).
//
// All methods must be safe to call from multiple goroutines.
// Implementations should be non-blocking; if a method needs to do
// expensive work it should do so asynchronously.
//
// A nil ObservabilityHooks is valid and means no metrics are recorded.
type ObservabilityHooks interface {
	// RecordLoginAttempt is called after every login attempt (success or
	// failure). method is "password" or "passwordless". err is nil on
	// success.
	RecordLoginAttempt(method string, success bool, err error)

	// RecordRegistrationAttempt is called after every registration attempt.
	// err is nil on success.
	RecordRegistrationAttempt(success bool, err error)

	// RecordRateLimitHit is called when a request is denied due to rate
	// limiting.
	RecordRateLimitHit(endpoint string, ip string)

	// RecordSessionCreated is called when a new session/token is issued
	// for a user.
	RecordSessionCreated(userID string)

	// RecordPasswordReset is called after a password reset attempt.
	// err is nil on success.
	RecordPasswordReset(success bool, err error)
}

// Login method identifiers used by RecordLoginAttempt.
const (
	LoginMethodPassword     = "password"
	LoginMethodPasswordless = "passwordless"
)

// NoopObservabilityHooks is a no-op implementation of ObservabilityHooks.
// It is the default when no hooks are configured.
type NoopObservabilityHooks struct{}

func (NoopObservabilityHooks) RecordLoginAttempt(string, bool, error) {}
func (NoopObservabilityHooks) RecordRegistrationAttempt(bool, error)  {}
func (NoopObservabilityHooks) RecordRateLimitHit(string, string)      {}
func (NoopObservabilityHooks) RecordSessionCreated(string)            {}
func (NoopObservabilityHooks) RecordPasswordReset(bool, error)        {}
