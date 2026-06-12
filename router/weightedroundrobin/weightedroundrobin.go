package weightedroundrobin

import (
	"errors"
	"load-balancer/backend"
	"sync"
)

var ErrNoBackends = errors.New("no backends available")

type BackendPoolIO interface {
	GetPool() []*backend.Backend
}

type WeightedRoundRobin struct {
	bs     []*backend.Backend
	index  int
	weight int
	mu     sync.Mutex
}

func NewWeightedRoundRobin(bp BackendPoolIO) *WeightedRoundRobin {
	if bp == nil {
		panic("backend pool cannot be nil")
	}
	copiedBp := bp.GetPool()
	bs := make([]*backend.Backend, 0, len(copiedBp))
	for _, b := range copiedBp {
		if b.GetWeight() > 0 {
			bs = append(bs, b)
		}
	}
	if len(bs) < 1 {
		panic("total backend pool weight cannot be zero")
	}
	return &WeightedRoundRobin{
		bs:     bs,
		weight: bs[0].GetWeight(),
	}
}

func (wrr *WeightedRoundRobin) GetBackend() (*backend.Backend, error) {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()
	for range len(wrr.bs) {
		b := wrr.bs[wrr.index]
		if b.IsHealthy() && !b.IsOpen() {
			wrr.weight--
			if wrr.weight <= 0 {
				wrr.index = (wrr.index + 1) % len(wrr.bs)
				wrr.weight = wrr.bs[wrr.index].GetWeight()
			}
			return b, nil
		}
		wrr.index = (wrr.index + 1) % len(wrr.bs)
		wrr.weight = wrr.bs[wrr.index].GetWeight()
	}
	return nil, ErrNoBackends
}
