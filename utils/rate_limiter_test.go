package utils

import (
	"sync"
	"testing"
	"time"
)

func TestInMemoryRateLimiter_AllowsUpToMaxAttemptsThenBlocks(t *testing.T) {
	limiter := NewInMemoryRateLimiter(3, time.Second, 2*time.Second)
	defer limiter.Stop()

	ip := "127.0.0.1"
	endpoint := "login"

	for i := 0; i < 3; i++ {
		res := limiter.Check(ip, endpoint)
		if !res.Allowed {
			t.Fatalf("expected attempt %d to be allowed", i+1)
		}
	}

	res := limiter.Check(ip, endpoint)
	if res.Allowed {
		t.Fatalf("expected request to be blocked after exceeding max attempts")
	}
	if res.RetryAfter <= 0 {
		t.Fatalf("expected positive RetryAfter on block, got %v", res.RetryAfter)
	}
}

func TestInMemoryRateLimiter_LockoutExpires(t *testing.T) {
	lockout := 100 * time.Millisecond
	limiter := NewInMemoryRateLimiter(1, time.Second, lockout)
	defer limiter.Stop()

	ip := "127.0.0.1"
	endpoint := "login"

	res := limiter.Check(ip, endpoint)
	if !res.Allowed {
		t.Fatalf("expected first attempt to be allowed")
	}

	res = limiter.Check(ip, endpoint)
	if res.Allowed {
		t.Fatalf("expected second attempt to trigger lockout")
	}

	time.Sleep(lockout + 50*time.Millisecond)

	res = limiter.Check(ip, endpoint)
	if !res.Allowed {
		t.Fatalf("expected request to be allowed after lockout expires")
	}
	if res.RetryAfter != 0 {
		t.Fatalf("expected RetryAfter to be zero after lockout, got %v", res.RetryAfter)
	}
}

func TestInMemoryRateLimiter_WindowResetsAfterDuration(t *testing.T) {
	window := 50 * time.Millisecond
	limiter := NewInMemoryRateLimiter(2, window, time.Second)
	defer limiter.Stop()

	ip := "127.0.0.1"
	endpoint := "login"

	for i := 0; i < 2; i++ {
		res := limiter.Check(ip, endpoint)
		if !res.Allowed {
			t.Fatalf("expected attempt %d to be allowed", i+1)
		}
	}

	time.Sleep(window + 50*time.Millisecond)

	for i := 0; i < 2; i++ {
		res := limiter.Check(ip, endpoint)
		if !res.Allowed {
			t.Fatalf("expected attempt %d after window reset to be allowed", i+1)
		}
	}
}

func TestInMemoryRateLimiter_SeparatesKeysByIpAndEndpoint(t *testing.T) {
	limiter := NewInMemoryRateLimiter(1, time.Second, time.Second)
	defer limiter.Stop()

	ip1 := "127.0.0.1"
	ip2 := "10.0.0.1"
	endpoint := "login"

	res := limiter.Check(ip1, endpoint)
	if !res.Allowed {
		t.Fatalf("expected first attempt from ip1 to be allowed")
	}

	res = limiter.Check(ip1, endpoint)
	if res.Allowed {
		t.Fatalf("expected second attempt from ip1 to be blocked")
	}

	res = limiter.Check(ip2, endpoint)
	if !res.Allowed {
		t.Fatalf("expected attempt from different IP to be allowed")
	}
}

// TestInMemoryRateLimiter_ConcurrentRaceCondition proves the data race in
// InMemoryRateLimiter.Check by hammering the same ip+endpoint key from many
// goroutines simultaneously. Run with: go test -race ./utils/...
func TestInMemoryRateLimiter_ConcurrentRaceCondition(t *testing.T) {
	limiter := NewInMemoryRateLimiter(1000, 10*time.Second, 10*time.Second)
	defer limiter.Stop()

	ip := "127.0.0.1"
	endpoint := "login"

	const goroutines = 100
	const iterationsPerGoroutine = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterationsPerGoroutine; j++ {
				limiter.Check(ip, endpoint)
			}
		}()
	}
	wg.Wait()
}

// TestInMemoryRateLimiter_ConcurrentRaceCondition_Bypass proves the data race
// by demonstrating its observable security effect: rate limit bypass.
// Without synchronization, multiple goroutines read len(record.timestamps) <
// maxAttempts simultaneously, all pass the check, and all append — so more
// requests are allowed than the limit permits.
//
// With maxAttempts=1, only ONE goroutine should ever get Allowed=true.
// If more than one does, the race is proven.
// Note: this test may not trigger on every run without -race; the race
// detector on CI (see .github/workflows/test.yml) provides definitive proof.
func TestInMemoryRateLimiter_ConcurrentRaceCondition_Bypass(t *testing.T) {
	limiter := NewInMemoryRateLimiter(1, 1*time.Hour, 1*time.Hour)
	defer limiter.Stop()

	ip := "127.0.0.1"
	endpoint := "login"

	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	allowedCount := int64(0)
	var countMu sync.Mutex

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			result := limiter.Check(ip, endpoint)
			if result.Allowed {
				countMu.Lock()
				allowedCount++
				countMu.Unlock()
			}
		}()
	}
	wg.Wait()

	if allowedCount > 1 {
		t.Fatalf("RATE LIMIT BYPASS: maxAttempts=1 but %d goroutines were allowed (expected at most 1). "+
			"This proves the data race: concurrent goroutines read len(record.timestamps) < maxAttempts "+
			"simultaneously before any of them appends.", allowedCount)
	}

	t.Logf("allowedCount=%d (race may not manifest on every run; use -race for definitive detection)", allowedCount)
}
