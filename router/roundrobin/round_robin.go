package roundrobin

import (
	"load-balancer/backend"
	"sync"
)

type BackendPoolIO interface {
	GetPool() []*backend.Backend
}

type RoundRobin struct {
	bp    BackendPoolIO
	index int
	mu    sync.Mutex
}

func NewRoundRobin(bs BackendPoolIO) *RoundRobin {
	return &RoundRobin{bp: bs}
}

func (rr *RoundRobin) GetBackend() *backend.Backend {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	bp := rr.bp.GetPool()
	b := bp[rr.index]
	rr.index = (rr.index + 1) % len(bp)
	return b
}
