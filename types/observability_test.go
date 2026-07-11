package types

import (
	"sync"
	"testing"
)

// mockObservabilityHooks is a test implementation that records all calls
// for later assertion in tests.
type mockObservabilityHooks struct {
	mu sync.Mutex

	LoginAttempts        []loginAttempt
	RegistrationAttempts []registrationAttempt
	RateLimitHits        []rateLimitHit
	SessionsCreated      []string
	PasswordResets       []passwordReset
}

type loginAttempt struct {
	Method  string
	Success bool
	Err     error
}

type registrationAttempt struct {
	Success bool
	Err     error
}

type rateLimitHit struct {
	Endpoint string
	IP       string
}

type passwordReset struct {
	Success bool
	Err     error
}

func (m *mockObservabilityHooks) RecordLoginAttempt(method string, success bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LoginAttempts = append(m.LoginAttempts, loginAttempt{method, success, err})
}

func (m *mockObservabilityHooks) RecordRegistrationAttempt(success bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RegistrationAttempts = append(m.RegistrationAttempts, registrationAttempt{success, err})
}

func (m *mockObservabilityHooks) RecordRateLimitHit(endpoint string, ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.RateLimitHits = append(m.RateLimitHits, rateLimitHit{endpoint, ip})
}

func (m *mockObservabilityHooks) RecordSessionCreated(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SessionsCreated = append(m.SessionsCreated, userID)
}

func (m *mockObservabilityHooks) RecordPasswordReset(success bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PasswordResets = append(m.PasswordResets, passwordReset{success, err})
}

func TestNoopObservabilityHooks(t *testing.T) {
	hooks := NoopObservabilityHooks{}

	// Verify all methods can be called without panicking
	hooks.RecordLoginAttempt(LoginMethodPassword, true, nil)
	hooks.RecordRegistrationAttempt(true, nil)
	hooks.RecordRateLimitHit("/login", "127.0.0.1")
	hooks.RecordSessionCreated("user-123")
	hooks.RecordPasswordReset(true, nil)
}

func TestMockObservabilityHooks(t *testing.T) {
	m := &mockObservabilityHooks{}

	m.RecordLoginAttempt(LoginMethodPassword, false, nil)
	m.RecordLoginAttempt(LoginMethodPasswordless, true, nil)
	m.RecordRegistrationAttempt(true, nil)
	m.RecordRateLimitHit("/login", "1.2.3.4")
	m.RecordSessionCreated("user-abc")
	m.RecordPasswordReset(false, nil)

	if len(m.LoginAttempts) != 2 {
		t.Fatalf("expected 2 login attempts, got %d", len(m.LoginAttempts))
	}
	if m.LoginAttempts[0].Method != LoginMethodPassword || m.LoginAttempts[0].Success != false {
		t.Errorf("unexpected first login attempt: %+v", m.LoginAttempts[0])
	}
	if m.LoginAttempts[1].Method != LoginMethodPasswordless || m.LoginAttempts[1].Success != true {
		t.Errorf("unexpected second login attempt: %+v", m.LoginAttempts[1])
	}

	if len(m.RegistrationAttempts) != 1 || !m.RegistrationAttempts[0].Success {
		t.Fatalf("expected 1 successful registration, got %+v", m.RegistrationAttempts)
	}

	if len(m.RateLimitHits) != 1 || m.RateLimitHits[0].Endpoint != "/login" || m.RateLimitHits[0].IP != "1.2.3.4" {
		t.Fatalf("unexpected rate limit hits: %+v", m.RateLimitHits)
	}

	if len(m.SessionsCreated) != 1 || m.SessionsCreated[0] != "user-abc" {
		t.Fatalf("unexpected sessions: %+v", m.SessionsCreated)
	}

	if len(m.PasswordResets) != 1 || m.PasswordResets[0].Success != false {
		t.Fatalf("unexpected password resets: %+v", m.PasswordResets)
	}
}

func TestMockObservabilityHooks_ConcurrentSafety(t *testing.T) {
	m := &mockObservabilityHooks{}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			m.RecordLoginAttempt(LoginMethodPassword, true, nil)
			m.RecordSessionCreated("user-concurrent")
			m.RecordRateLimitHit("/api/login", "10.0.0.1")
		}()
	}
	wg.Wait()

	if len(m.LoginAttempts) != goroutines {
		t.Errorf("expected %d login attempts, got %d", goroutines, len(m.LoginAttempts))
	}
	if len(m.SessionsCreated) != goroutines {
		t.Errorf("expected %d sessions, got %d", goroutines, len(m.SessionsCreated))
	}
	if len(m.RateLimitHits) != goroutines {
		t.Errorf("expected %d rate limit hits, got %d", goroutines, len(m.RateLimitHits))
	}
}
