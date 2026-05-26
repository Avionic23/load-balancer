package roundrobin

import (
	"errors"
	"load-balancer/backend"
	"sync"
)

var ErrNoBackends = errors.New("no backends available")

type BackendPoolIO interface {
	GetPool() []*backend.Backend
}

type RoundRobin struct {
	bp    BackendPoolIO
	index int
	mu    sync.Mutex
}

func NewRoundRobin(bs BackendPoolIO) *RoundRobin {
	if bs == nil {
		panic("backend pool cannot be nil")
	}
	return &RoundRobin{bp: bs}
}

func (rr *RoundRobin) GetBackend() (*backend.Backend, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	bp := rr.bp.GetPool()
	for range len(bp) {
		b := bp[rr.index]
		rr.index = (rr.index + 1) % len(bp)
		if b.IsHealthy() {
			return b, nil
		}
	}
	return nil, ErrNoBackends
}
