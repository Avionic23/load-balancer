package leastconnections

import (
	"errors"
	"load-balancer/backend"
	"math"
)

var ErrNoBackends = errors.New("no backends available")

type BackendPool interface {
	GetPool() []*backend.Backend
}

type LeastConnections struct {
	bp BackendPool
}

func NewLeastConnections(bp BackendPool) *LeastConnections {
	if bp == nil {
		panic("backend pool cannot be nil")
	}
	return &LeastConnections{bp: bp}
}

func (ls *LeastConnections) GetBackend() (*backend.Backend, error) {
	bp := ls.bp.GetPool()
	var best *backend.Backend
	bestConns := int64(math.MaxInt64)
	for _, b := range bp {
		if !b.IsHealthy() || b.IsOpen() {
			continue
		}
		if activeConns := b.GetActiveConns(); activeConns < bestConns {
			bestConns = activeConns
			best = b
		}
	}
	if best == nil {
		return nil, ErrNoBackends
	}
	return best, nil
}
