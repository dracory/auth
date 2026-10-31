package utils

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dracory/auth/types"
)

func TestDummyPasswordCompare(t *testing.T) {
	DummyPasswordCompare("any-password")
	// Second call reuses the lazily-computed hash.
	DummyPasswordCompare("another")
}

func TestIsSafeReturnPath(t *testing.T) {
	cases := map[string]bool{
		"/dashboard":        true,
		"/a/b?c=d":          true,
		"":                  false,
		"dashboard":         false,
		"//evil.com":        false,
		"/\\evil.com":       false,
		"/ok\r\nSet-Cookie": false,
	}
	for path, want := range cases {
		if got := IsSafeReturnPath(path); got != want {
			t.Errorf("IsSafeReturnPath(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestIsSafeBackURL(t *testing.T) {
	host := "example.com"
	cases := map[string]bool{
		"":                              false,
		"/auth/login":                   true,
		"//evil.com":                    false,
		"https://example.com/done":      true,
		"https://EXAMPLE.com/done":      true,
		"https://evil.com/done":         false,
		"https://example.com.evil.com/": false,
		":://not-a-url":                 false,
	}
	for u, want := range cases {
		if got := IsSafeBackURL(u, host); got != want {
			t.Errorf("IsSafeBackURL(%q) = %v, want %v", u, got, want)
		}
	}
}

func TestIsValidNonce(t *testing.T) {
	if IsValidNonce("not-hex") {
		t.Error("expected invalid nonce")
	}
	if IsValidNonce("abcdef") {
		t.Error("expected invalid nonce for wrong length")
	}
	if !IsValidNonce(strings.Repeat("a1", 16)) {
		t.Error("expected valid 32-char hex nonce")
	}
}

func TestInvalidateTempKey(t *testing.T) {
	deleted := ""
	set := ""

	deleteFn := func(key string) error { deleted = key; return nil }
	setFn := func(key string, value string, expiresSeconds int) error {
		set = key
		if value != "" || expiresSeconds != 0 {
			t.Errorf("expected empty value with 0 expiry, got %q %d", value, expiresSeconds)
		}
		return nil
	}

	if err := InvalidateTempKey(deleteFn, setFn, "k1"); err != nil || deleted != "k1" {
		t.Fatalf("expected delete path, got deleted=%q err=%v", deleted, err)
	}
	if err := InvalidateTempKey(nil, setFn, "k2"); err != nil || set != "k2" {
		t.Fatalf("expected set fallback, got set=%q err=%v", set, err)
	}
	if err := InvalidateTempKey(nil, nil, "k3"); err != nil {
		t.Fatalf("expected nil no-op, got %v", err)
	}

	wantErr := errors.New("boom")
	if err := InvalidateTempKey(func(key string) error { return wantErr }, nil, "k4"); err != wantErr {
		t.Fatalf("expected delete error propagation, got %v", err)
	}
}

func TestRememberMeCheckbox(t *testing.T) {
	if got := RememberMeCheckbox(false); got != "" {
		t.Fatalf("expected empty when disabled, got %q", got)
	}
	if got := RememberMeCheckbox(true); !strings.Contains(got, `name="remember"`) {
		t.Fatalf("expected remember checkbox html, got %q", got)
	}
}

func TestRememberCookieSetAndRemove(t *testing.T) {
	rec := httptest.NewRecorder()
	RememberCookieSet(rec, "remember", "key-1", 30, true)

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one cookie, got %d", len(cookies))
	}
	c := cookies[0]
	if c.Name != "remember" || c.Value != "key-1" || !c.HttpOnly || !c.Secure || c.MaxAge != 30*24*3600 {
		t.Fatalf("unexpected remember cookie %+v", c)
	}

	rec = httptest.NewRecorder()
	RememberCookieRemove(rec, "remember")
	c = rec.Result().Cookies()[0]
	if c.MaxAge != -1 || c.Value != "" {
		t.Fatalf("expected expired cookie, got %+v", c)
	}
}

func TestEmailSendsKey(t *testing.T) {
	k1 := EmailSendsKey("User@Example.com")
	k2 := EmailSendsKey("user@example.com")
	if k1 != k2 {
		t.Fatal("expected email case-insensitive key")
	}
	if !strings.HasPrefix(k1, EmailSendKeyPrefix) || len(k1) != len(EmailSendKeyPrefix)+64 {
		t.Fatalf("unexpected key %q", k1)
	}
}

func TestCanSendEmailAndRecordEmailSend(t *testing.T) {
	store := map[string]string{}
	get := func(key string) (string, error) { return store[key], nil }
	set := func(key string, value string, expiresSeconds int) error {
		store[key] = value
		return nil
	}
	cfg := EmailSendThrottle{MaxSends: 2, WindowSeconds: 60}

	if !CanSendEmail(get, "a@b.com", cfg) {
		t.Fatal("expected first send allowed")
	}
	if err := RecordEmailSend(get, set, "a@b.com", cfg); err != nil {
		t.Fatal(err)
	}
	if err := RecordEmailSend(get, set, "a@b.com", cfg); err != nil {
		t.Fatal(err)
	}
	if CanSendEmail(get, "a@b.com", cfg) {
		t.Fatal("expected third send throttled")
	}
	// A different email is unaffected.
	if !CanSendEmail(get, "c@d.com", cfg) {
		t.Fatal("expected other email allowed")
	}
}

func TestEmailSendThrottleEdgeCases(t *testing.T) {
	cfg := EmailSendThrottle{MaxSends: 1, WindowSeconds: 60}

	// nil get fails open.
	if !CanSendEmail(nil, "a@b.com", cfg) {
		t.Fatal("expected fail-open on nil get")
	}
	// get error fails open.
	errGet := func(key string) (string, error) { return "", errors.New("store down") }
	if !CanSendEmail(errGet, "a@b.com", cfg) {
		t.Fatal("expected fail-open on get error")
	}
	// unparsable stored value counts as zero.
	badGet := func(key string) (string, error) { return "not-a-number", nil }
	if !CanSendEmail(badGet, "a@b.com", cfg) {
		t.Fatal("expected fail-open on unparsable count")
	}
	// nil set is a no-op.
	if err := RecordEmailSend(badGet, nil, "a@b.com", cfg); err != nil {
		t.Fatal(err)
	}
	// set error propagates.
	wantErr := errors.New("set failed")
	errSet := func(key string, value string, expiresSeconds int) error { return wantErr }
	if err := RecordEmailSend(nil, errSet, "a@b.com", cfg); err != wantErr {
		t.Fatalf("expected set error, got %v", err)
	}
	// zero-value config normalizes to defaults.
	n := EmailSendThrottle{}.normalize()
	if n.MaxSends != DefaultEmailSendMaxSends || n.WindowSeconds != DefaultEmailSendWindowSeconds {
		t.Fatalf("unexpected defaults %+v", n)
	}
}

type throttleAuth struct {
	types.AuthSharedInterface
	cfg EmailSendThrottle
}

func (a throttleAuth) GetEmailSendThrottleConfig() EmailSendThrottle { return a.cfg }

func TestEmailSendThrottleFromAuth(t *testing.T) {
	cfg := EmailSendThrottle{MaxSends: 7, WindowSeconds: 9}
	if got := EmailSendThrottleFromAuth(throttleAuth{cfg: cfg}); got != cfg {
		t.Fatalf("expected configured throttle %+v, got %+v", cfg, got)
	}
	if got := EmailSendThrottleFromAuth(throttleAuth{}); got != (EmailSendThrottle{}) {
		t.Fatalf("expected zero config without interface support, got %+v", got)
	}
}

func TestMemSessionAccessorsAndStore(t *testing.T) {
	s := NewMemSession()
	s.SetID("id-1")
	s.SetKey("key-1")
	s.SetUserID("u1")

	if s.GetID() != "id-1" || s.GetKey() != "key-1" {
		t.Fatalf("unexpected session %+v", s)
	}
	if !s.GetExpiresAt().IsZero() {
		t.Fatal("expected zero expiry")
	}

	store := NewMemSessionStore()
	if _, ok := store.NewSession().(*MemSession); !ok {
		t.Fatal("expected *MemSession from store")
	}

	if err := store.SessionCreate(context.Background(), s); err != nil {
		t.Fatal(err)
	}

	// SessionUpdate shortens the session's expiry in place.
	s.SetExpiresAt(s.GetExpiresAt().Add(0)) // no-op touch
	if err := store.SessionUpdate(context.Background(), s); err != nil {
		t.Fatal(err)
	}

	// SessionUpdate errors.
	if err := store.SessionUpdate(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil session")
	}
	if err := store.SessionUpdate(context.Background(), otherSession{}); err == nil {
		t.Fatal("expected error for foreign session type")
	}
	if err := store.SessionCreate(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil session create")
	}
}

type otherSession struct{ types.SessionInterface }
