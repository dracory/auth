package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
)

// EmailSendKeyPrefix prefixes the per-email send-throttle counters.
const EmailSendKeyPrefix = "code_sent:"

// MaxEmailSends is the number of code-send requests allowed per email
// address within one TTL window.
const MaxEmailSends = 3

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
// the current TTL window. A store read error fails open (treated as 0 sends)
// so a flaky store does not block legitimate sends; the send throttle stays
// best-effort like the IP rate limiter.
func CanSendEmail(get func(key string) (string, error), email string) bool {
	sendThrottleMu.Lock()
	defer sendThrottleMu.Unlock()
	return emailSendCount(get, email) < MaxEmailSends
}

// RecordEmailSend increments the per-email send counter. Call it ONLY after
// the email send (or enqueue) succeeds — failed sends must not consume the
// victim's quota, otherwise transient mailer errors would lock the real
// user out.
// Each call rewrites the key with a fresh TTL, so this is a sliding window:
// ~MaxEmailSends sends per `ttlSeconds` sustained, not a fixed quota.
func RecordEmailSend(
	get func(key string) (string, error),
	set func(key string, value string, expiresSeconds int) error,
	email string,
	ttlSeconds int,
) error {
	sendThrottleMu.Lock()
	defer sendThrottleMu.Unlock()

	if ttlSeconds <= 0 {
		ttlSeconds = 3600
	}

	count := emailSendCount(get, email)
	if set == nil {
		return nil
	}
	return set(EmailSendsKey(email), fmt.Sprintf("%d", count+1), ttlSeconds)
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
