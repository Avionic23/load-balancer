package leastconnections

import (
	"load-balancer/backend"
	"sync"
	"testing"
)

type stubPool struct {
	backends []*backend.Backend
}

func (s *stubPool) GetPool() []*backend.Backend {
	return s.backends
}

func newHealthy(activeConns int64) *backend.Backend {
	b := backend.NewBackend(backend.BackendOptions{Url: "127.0.0.1:0"})
	b.SetHealth(true)
	for range activeConns {
		b.IncrActiveConns()
	}
	return b
}

func newUnhealthy(activeConns int64) *backend.Backend {
	b := backend.NewBackend(backend.BackendOptions{Url: "127.0.0.1:0"})
	for range activeConns {
		b.IncrActiveConns()
	}
	return b
}

func TestPicksLeastConnectedBackend(t *testing.T) {
	busy := newHealthy(5)
	idle := newHealthy(1)
	lc := NewLeastConnections(&stubPool{backends: []*backend.Backend{busy, idle}})

	got, err := lc.GetBackend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != idle {
		t.Error("expected backend with fewest connections")
	}
}

func TestSkipsUnhealthyBackend(t *testing.T) {
	unhealthy := newUnhealthy(0)
	healthy := newHealthy(5)
	lc := NewLeastConnections(&stubPool{backends: []*backend.Backend{unhealthy, healthy}})

	got, err := lc.GetBackend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != healthy {
		t.Error("expected healthy backend even with higher conn count")
	}
}

func TestReturnsErrNoBackendsWhenAllUnhealthy(t *testing.T) {
	lc := NewLeastConnections(&stubPool{backends: []*backend.Backend{
		newUnhealthy(0),
		newUnhealthy(2),
	}})

	_, err := lc.GetBackend()
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestReturnsErrNoBackendsOnEmptyPool(t *testing.T) {
	lc := NewLeastConnections(&stubPool{backends: []*backend.Backend{}})

	_, err := lc.GetBackend()
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestPicksOnlyHealthyBackendRegardlessOfConnCount(t *testing.T) {
	healthy := newHealthy(100)
	lc := NewLeastConnections(&stubPool{backends: []*backend.Backend{
		newUnhealthy(0),
		newUnhealthy(0),
		healthy,
	}})

	got, err := lc.GetBackend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != healthy {
		t.Error("expected the only healthy backend")
	}
}

func TestEqualConnectionsReturnsHealthyBackend(t *testing.T) {
	a := newHealthy(3)
	b := newHealthy(3)
	lc := NewLeastConnections(&stubPool{backends: []*backend.Backend{a, b}})

	got, err := lc.GetBackend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.IsHealthy() {
		t.Error("expected a healthy backend")
	}
}

func TestConcurrentSafety(t *testing.T) {
	backends := []*backend.Backend{
		newHealthy(0),
		newHealthy(1),
		newHealthy(2),
	}
	lc := NewLeastConnections(&stubPool{backends: backends})

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lc.GetBackend()
		}()
	}
	wg.Wait()
}
