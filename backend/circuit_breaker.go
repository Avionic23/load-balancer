package backend

import (
	"sync"
	"time"
)

type CircuitState int

const (
	Closed CircuitState = iota
	Open
	HalfOpen
)

type CircuitBreaker struct {
	state             CircuitState
	window            []byte
	windowSize        int
	successes         int
	failures          int
	failureThreshold  int
	halfOpenThreshold int
	cooldownTimeout   time.Duration
	cooldownStartedAt time.Time
	mu                sync.Mutex
}

type CircuitBreakerOptions struct {
	WindowSize        int
	FailureThreshold  int
	HalfOpenThreshold int
	CooldownTimeout   time.Duration
}

func NewCircuitBreaker(opts CircuitBreakerOptions) CircuitBreaker {
	return CircuitBreaker{
		state:             Closed,
		window:            make([]byte, 0, opts.WindowSize),
		windowSize:        opts.WindowSize,
		failureThreshold:  opts.FailureThreshold,
		halfOpenThreshold: opts.HalfOpenThreshold,
		cooldownTimeout:   opts.CooldownTimeout,
	}
}

func (b *CircuitBreaker) IsOpen() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state != Open {
		return false
	}
	if time.Now().After(b.cooldownStartedAt.Add(b.cooldownTimeout)) {
		b.state = HalfOpen
		return false
	}
	return true
}

func (b *CircuitBreaker) RecordFailure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == Open {
		return
	}
	b.failures++
	if b.state == Closed {
		if len(b.window) >= b.windowSize {
			if b.window[0] == 'S' {
				b.successes--
			}
			if b.window[0] == 'F' {
				b.failures--
			}
			copy(b.window, b.window[1:])
			b.window[len(b.window)-1] = 'F'
		} else {
			b.window = append(b.window, 'F')
		}
	}
	if b.state == HalfOpen || b.failures*100/b.windowSize >= b.failureThreshold {
		b.state = Open
		b.cooldownStartedAt = time.Now()
		b.successes, b.failures = 0, 0
		b.window = make([]byte, 0, b.windowSize)
	}
}

func (b *CircuitBreaker) RecordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == Open {
		return
	}
	b.successes++
	if b.state == Closed {
		if len(b.window) >= b.windowSize {
			if b.window[0] == 'S' {
				b.successes--
			}
			if b.window[0] == 'F' {
				b.failures--
			}
			copy(b.window, b.window[1:])
			b.window[len(b.window)-1] = 'S'
		} else {
			b.window = append(b.window, 'S')
		}
	}
	if b.state == HalfOpen {
		if b.successes >= b.halfOpenThreshold {
			b.state = Closed
			b.successes = 0
		}
	}
}
