package backend

import (
	"sync"
	"testing"
	"time"
)

func newTestCB() CircuitBreaker {
	return CircuitBreaker{
		state:             Closed,
		window:            make([]byte, 0, 4),
		windowSize:        4,
		failureThreshold:  50,
		halfOpenThreshold: 2,
		cooldownTimeout:   time.Second,
	}
}

func TestClosedToOpenOnFailureThreshold(t *testing.T) {
	cb := newTestCB()

	// 2 successes + 2 failures = 50% failure rate in window of 4
	cb.RecordSuccess()
	cb.RecordSuccess()
	cb.RecordFailure()
	cb.RecordFailure()

	if !cb.IsOpen() {
		t.Error("expected circuit to be open after failure threshold exceeded")
	}
}

func TestStaysClosedBelowThreshold(t *testing.T) {
	cb := newTestCB()

	cb.RecordSuccess()
	cb.RecordSuccess()
	cb.RecordSuccess()
	cb.RecordFailure()

	if cb.IsOpen() {
		t.Error("expected circuit to stay closed below failure threshold")
	}
}

func TestOpenToHalfOpenAfterCooldown(t *testing.T) {
	cb := newTestCB()

	cb.RecordSuccess()
	cb.RecordSuccess()
	cb.RecordFailure()
	cb.RecordFailure()

	// manually expire the cooldown
	cb.mu.Lock()
	cb.cooldownStartedAt = time.Now().Add(-2 * time.Second)
	cb.mu.Unlock()

	if cb.IsOpen() {
		t.Error("expected circuit to transition to half-open after cooldown")
	}
	cb.mu.Lock()
	state := cb.state
	cb.mu.Unlock()
	if state != HalfOpen {
		t.Errorf("expected HalfOpen state, got %v", state)
	}
}

func TestHalfOpenToClosedOnSuccesses(t *testing.T) {
	cb := newTestCB()
	cb.mu.Lock()
	cb.state = HalfOpen
	cb.mu.Unlock()

	cb.RecordSuccess()
	cb.RecordSuccess()

	if cb.IsOpen() {
		t.Error("expected circuit to close after half-open threshold successes")
	}
	cb.mu.Lock()
	state := cb.state
	cb.mu.Unlock()
	if state != Closed {
		t.Errorf("expected Closed state, got %v", state)
	}
}

func TestHalfOpenToOpenOnFailure(t *testing.T) {
	cb := newTestCB()
	cb.mu.Lock()
	cb.state = HalfOpen
	cb.mu.Unlock()

	cb.RecordFailure()

	if !cb.IsOpen() {
		t.Error("expected circuit to reopen on failure in half-open state")
	}
}

func TestRecordsIgnoredWhenOpen(t *testing.T) {
	cb := newTestCB()
	cb.mu.Lock()
	cb.state = Open
	cb.cooldownStartedAt = time.Now()
	cb.mu.Unlock()

	cb.RecordFailure()
	cb.RecordSuccess()

	cb.mu.Lock()
	failures := cb.failures
	successes := cb.successes
	cb.mu.Unlock()

	if failures != 0 || successes != 0 {
		t.Errorf("expected no records in open state, got failures=%d successes=%d", failures, successes)
	}
}

func TestConcurrentSafety(t *testing.T) {
	cb := newTestCB()
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(3)
		go func() { defer wg.Done(); cb.RecordFailure() }()
		go func() { defer wg.Done(); cb.RecordSuccess() }()
		go func() { defer wg.Done(); cb.IsOpen() }()
	}
	wg.Wait()
}
