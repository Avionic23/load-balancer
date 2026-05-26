package backend

import (
	"sync"
	"sync/atomic"
)

type Backend struct {
	url     string
	healthy atomic.Bool
}

func NewBackend(url string) *Backend {
	return &Backend{url: url}
}

func (b *Backend) GetUrl() string {
	return b.url
}

func (b *Backend) IsHealthy() bool {
	return b.healthy.Load()
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
