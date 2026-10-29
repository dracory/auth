package utils

import (
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// dummyPasswordHash is a precomputed bcrypt hash compared against on
// unknown-email logins so the response time matches a real wrong-password
// compare (anti-enumeration). Computed once lazily — bcrypt cost is only
// paid on the first unknown-email login, not at startup.
var (
	dummyPasswordHash     []byte
	dummyPasswordHashOnce sync.Once
)

// DummyPasswordCompare runs a bcrypt compare of the given password against
// a fixed dummy hash. Callers discard the result; the call exists purely to
// spend the same CPU time as a real password verification so unknown-email
// logins are not timing-distinguishable from wrong-password logins.
func DummyPasswordCompare(password string) {
	dummyPasswordHashOnce.Do(func() {
		h, err := bcrypt.GenerateFromPassword([]byte("dracory-auth-dummy-password"), bcrypt.DefaultCost)
		if err != nil {
			// bcrypt only errors on an overlong password (>72 bytes) or a
			// broken rand source; the constant below keeps the compare
			// working even then.
			h = []byte("$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy")
		}
		dummyPasswordHash = h
	})
	_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
}
