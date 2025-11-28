package utils

import (
	"sync"
	"time"
)

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed    bool
	RetryAfter time.Duration
}

// requestRecord tracks individual request timestamps for an IP
type requestRecord struct {
	timestamps  []time.Time
	lockedUntil time.Time
}

// InMemoryRateLimiter provides thread-safe in-memory rate limiting
type InMemoryRateLimiter struct {
	records          sync.Map // map[string]*requestRecord (key: "ip:endpoint")
	maxAttempts      int
	windowDuration   time.Duration
	lockoutDuration  time.Duration
	cleanupInterval  time.Duration
	stopCleanup      chan struct{}
	cleanupWaitGroup sync.WaitGroup
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter with default settings
func NewInMemoryRateLimiter(maxAttempts int, windowDuration time.Duration, lockoutDuration time.Duration) *InMemoryRateLimiter {
	limiter := &InMemoryRateLimiter{
		maxAttempts:     maxAttempts,
		windowDuration:  windowDuration,
		lockoutDuration: lockoutDuration,
		cleanupInterval: 5 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}

	// Start background cleanup goroutine
	limiter.cleanupWaitGroup.Add(1)
	go limiter.cleanupOldRecords()

	return limiter
}

// Check verifies if a request from the given IP to the given endpoint should be allowed
func (r *InMemoryRateLimiter) Check(ip string, endpoint string) RateLimitResult {
	key := ip + ":" + endpoint
	now := time.Now()

	// Load or create record
	recordInterface, _ := r.records.LoadOrStore(key, &requestRecord{
		timestamps:  make([]time.Time, 0),
		lockedUntil: time.Time{},
	})
	record := recordInterface.(*requestRecord)

	// Check if currently locked out
	if !record.lockedUntil.IsZero() && now.Before(record.lockedUntil) {
		retryAfter := record.lockedUntil.Sub(now)
		return RateLimitResult{
			Allowed:    false,
			RetryAfter: retryAfter,
		}
	}

	// If lockout has expired, reset the record
	if !record.lockedUntil.IsZero() && now.After(record.lockedUntil) {
		record.timestamps = make([]time.Time, 0)
		record.lockedUntil = time.Time{}
	}

	// Remove timestamps outside the window
	cutoff := now.Add(-r.windowDuration)
	validTimestamps := make([]time.Time, 0)
	for _, ts := range record.timestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	record.timestamps = validTimestamps

	// Check if limit exceeded
	if len(record.timestamps) >= r.maxAttempts {
		// Lock out the IP
		record.lockedUntil = now.Add(r.lockoutDuration)
		r.records.Store(key, record)
		return RateLimitResult{
			Allowed:    false,
			RetryAfter: r.lockoutDuration,
		}
	}

	// Add current timestamp and allow request
	record.timestamps = append(record.timestamps, now)
	r.records.Store(key, record)

	return RateLimitResult{
		Allowed:    true,
		RetryAfter: 0,
	}
}

// cleanupOldRecords periodically removes old records to prevent memory leaks
func (r *InMemoryRateLimiter) cleanupOldRecords() {
	defer r.cleanupWaitGroup.Done()
	ticker := time.NewTicker(r.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			cutoff := now.Add(-r.windowDuration - r.lockoutDuration)

			r.records.Range(func(key, value interface{}) bool {
				record := value.(*requestRecord)

				// If no recent activity and not locked, remove the record
				if len(record.timestamps) == 0 ||
					(len(record.timestamps) > 0 && record.timestamps[len(record.timestamps)-1].Before(cutoff)) {
					if record.lockedUntil.IsZero() || record.lockedUntil.Before(now) {
						r.records.Delete(key)
					}
				}

				return true // continue iteration
			})
		case <-r.stopCleanup:
			return
		}
	}
}

// Stop gracefully stops the rate limiter's background cleanup
func (r *InMemoryRateLimiter) Stop() {
	close(r.stopCleanup)
	r.cleanupWaitGroup.Wait()
}
