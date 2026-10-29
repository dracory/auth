package utils_test

import (
	"testing"
	"time"

	"github.com/dracory/auth/utils"
)

func TestInMemoryRateLimiter_PerEndpointLimit(t *testing.T) {
	l := utils.NewInMemoryRateLimiter(5, time.Minute, time.Minute)
	l.SetEndpointLimit("register", 10)

	// login limited at 5 (default), register allowed up to 10.
	for i := 0; i < 5; i++ {
		if !l.Check("1.1.1.1", "login").Allowed {
			t.Fatalf("login attempt %d should be allowed", i+1)
		}
	}
	if l.Check("1.1.1.1", "login").Allowed {
		t.Fatal("6th login attempt should be limited")
	}
	for i := 0; i < 10; i++ {
		if !l.Check("1.1.1.1", "register").Allowed {
			t.Fatalf("register attempt %d should be allowed", i+1)
		}
	}
	if l.Check("1.1.1.1", "register").Allowed {
		t.Fatal("11th register attempt should be limited")
	}
	// Different IP unaffected.
	if !l.Check("2.2.2.2", "login").Allowed {
		t.Fatal("different IP should not be cross-limited")
	}
}
