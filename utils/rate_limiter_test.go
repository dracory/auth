package utils

import (
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
