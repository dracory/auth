package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/dracory/auth/types"
)

// EmailSendThrottle configures the per-email send quota enforced before
// any code or link is sent. The zero value applies the package defaults
// (DefaultEmailSendMaxSends per DefaultEmailSendWindowSeconds).
type EmailSendThrottle struct {
	// MaxSends is the number of sends allowed per email within the
	// window; <= 0 falls back to DefaultEmailSendMaxSends.
	MaxSends int
	// WindowSeconds is the sliding-window length in seconds; <= 0 falls
	// back to DefaultEmailSendWindowSeconds. The window restarts on each
	// successful send (RecordEmailSend rewrites the counter with a fresh
	// TTL), so MaxSends=1 + WindowSeconds=300 acts as a 5-minute cooldown.
	WindowSeconds int
}

// normalize applies the package defaults to a zero-value config.
func (c EmailSendThrottle) normalize() EmailSendThrottle {
	if c.MaxSends <= 0 {
		c.MaxSends = DefaultEmailSendMaxSends
	}
	if c.WindowSeconds <= 0 {
		c.WindowSeconds = DefaultEmailSendWindowSeconds
	}
	return c
}

// sendThrottleMu serializes the get-then-set on send counters, which is not
// atomic through the TemporaryKeyGet/Set function pair.
// Note: check (CanSendEmail) and act (RecordEmailSend) take this mutex
// separately — concurrent sends can all pass the check before any records,
// briefly exceeding the quota. Combined with the mutex being per-process,
// this throttle is best-effort, same posture as the IP rate limiter.
var sendThrottleMu sync.Mutex

// EmailSendsKey returns the temporary-key for the per-email send counter.
// The email is SHA-256 hashed (lowercased) so the store holds no PII —
// throttle keys can outlive the code keys they protect.
func EmailSendsKey(email string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(strings.TrimSpace(email))))
	return EmailSendKeyPrefix + hex.EncodeToString(sum[:])
}

// CanSendEmail reports whether another code may be sent to the email within
// the current throttle window. A store read error fails open (treated as 0
// sends) so a flaky store does not block legitimate sends; the send throttle
// stays best-effort like the IP rate limiter.
func CanSendEmail(get func(key string) (string, error), email string, config EmailSendThrottle) bool {
	sendThrottleMu.Lock()
	defer sendThrottleMu.Unlock()
	return emailSendCount(get, email) < config.normalize().MaxSends
}

// RecordEmailSend increments the per-email send counter. Call it ONLY after
// the email send (or enqueue) succeeds — failed sends must not consume the
// victim's quota, otherwise transient mailer errors would lock the real
// user out.
// Each call rewrites the key with a fresh TTL (config.WindowSeconds), so
// this is a sliding window: ~MaxSends sends per window sustained, not a
// fixed quota.
func RecordEmailSend(
	get func(key string) (string, error),
	set func(key string, value string, expiresSeconds int) error,
	email string,
	config EmailSendThrottle,
) error {
	sendThrottleMu.Lock()
	defer sendThrottleMu.Unlock()

	count := emailSendCount(get, email)
	if set == nil {
		return nil
	}
	return set(EmailSendsKey(email), fmt.Sprintf("%d", count+1), config.normalize().WindowSeconds)
}

// EmailSendThrottleFromAuth resolves the throttle config from an auth
// implementation via the optional GetEmailSendThrottleConfig interface —
// the same optional-hook pattern used for GetFuncEmailEnqueue, so
// AuthSharedInterface stays stable. Unconfigured implementations get the
// package defaults.
func EmailSendThrottleFromAuth(a types.AuthSharedInterface) EmailSendThrottle {
	if c, ok := a.(interface {
		GetEmailSendThrottleConfig() EmailSendThrottle
	}); ok {
		return c.GetEmailSendThrottleConfig()
	}
	return EmailSendThrottle{}
}

func emailSendCount(get func(key string) (string, error), email string) int {
	if get == nil {
		return 0
	}
	count := 0
	if s, err := get(EmailSendsKey(email)); err == nil {
		_, _ = fmt.Sscanf(s, "%d", &count)
	}
	return count
}
