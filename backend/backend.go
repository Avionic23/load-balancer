package backend

import (
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	url         string
	healthy     atomic.Bool
	weight      int
	activeConns atomic.Int64
	cb          CircuitBreaker
}

type BackendOptions struct {
	Url                      string
	Weight                   int
	CircuitWindowSize        int
	CircuitFailureThreshold  int
	CircuitHalfOpenThreshold int
	CircuitCooldownTimeout   time.Duration
}

func NewBackend(opts BackendOptions) *Backend {
	if opts.CircuitWindowSize == 0 {
		opts.CircuitWindowSize = 10
	}
	if opts.CircuitFailureThreshold == 0 {
		opts.CircuitFailureThreshold = 50
	}
	if opts.CircuitHalfOpenThreshold == 0 {
		opts.CircuitHalfOpenThreshold = 3
	}
	if opts.CircuitCooldownTimeout == 0 {
		opts.CircuitCooldownTimeout = 30 * time.Second
	}
	return &Backend{
		url:    opts.Url,
		weight: opts.Weight,
		cb: NewCircuitBreaker(CircuitBreakerOptions{
			WindowSize:        opts.CircuitWindowSize,
			FailureThreshold:  opts.CircuitFailureThreshold,
			HalfOpenThreshold: opts.CircuitHalfOpenThreshold,
			CooldownTimeout:   opts.CircuitCooldownTimeout,
		}),
	}
}

func (b *Backend) GetUrl() string {
	return b.url
}

func (b *Backend) GetWeight() int {
	return b.weight
}

func (b *Backend) IsHealthy() bool {
	return b.healthy.Load()
}

func (b *Backend) SetHealth(isHealthy bool) {
	b.healthy.Store(isHealthy)
}

func (b *Backend) IncrActiveConns() {
	b.activeConns.Add(1)
}

func (b *Backend) DecrActiveConns() {
	b.activeConns.Add(-1)
}

func (b *Backend) GetActiveConns() int64 {
	return b.activeConns.Load()
}

func (b *Backend) IsOpen() bool {
	return b.cb.IsOpen()
}

func (b *Backend) RecordSuccess() {
	b.cb.RecordSuccess()
}

func (b *Backend) RecordFailure() {
	b.cb.RecordFailure()
}

type BackendPool struct {
	bs []*Backend
	mu sync.RWMutex
}

func NewBackendPool(bs []*Backend) *BackendPool {
	if len(bs) < 1 {
		panic("backend pool is empty")
	}
	return &BackendPool{bs: bs}
}

func (bp *BackendPool) GetPool() []*Backend {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	bs := make([]*Backend, len(bp.bs))
	copy(bs, bp.bs)
	return bs
}
