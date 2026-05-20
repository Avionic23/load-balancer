package backend

import "sync"

type Backend struct {
	url     string
	healthy bool
}

func NewBackend(url string) *Backend {
	return &Backend{url: url}
}

func (b *Backend) GetUrl() string {
	return b.url
}

func (b *Backend) IsHealthy() bool {
	return b.healthy
}

type BackendPool struct {
	bs []*Backend
	mu sync.RWMutex
}

func NewBackendPool(bs []*Backend) *BackendPool {
	return &BackendPool{bs: bs}
}

func (bp *BackendPool) GetPool() []*Backend {
	bp.mu.RLock()
	defer bp.mu.RUnlock()
	bs := make([]*Backend, len(bp.bs))
	copy(bs, bp.bs)
	return bs
}
