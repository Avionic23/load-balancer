package weightedroundrobin

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

func newHealthy(weight int) *backend.Backend {
	b := backend.NewBackend(backend.BackendOptions{Url: "127.0.0.1:0", Weight: weight})
	b.SetHealth(true)
	return b
}

func newUnhealthy(weight int) *backend.Backend {
	return backend.NewBackend(backend.BackendOptions{Url: "127.0.0.1:0", Weight: weight})
}

func TestWeightDistribution(t *testing.T) {
	a := newHealthy(3)
	b := newHealthy(1)
	wrr := NewWeightedRoundRobin(&stubPool{backends: []*backend.Backend{a, b}})

	want := []*backend.Backend{a, a, a, b, a, a, a, b}
	for i, expected := range want {
		got, err := wrr.GetBackend()
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if got != expected {
			t.Errorf("call %d: got wrong backend", i)
		}
	}
}

func TestSkipsUnhealthyBackend(t *testing.T) {
	healthy := newHealthy(2)
	unhealthy := newUnhealthy(3)
	wrr := NewWeightedRoundRobin(&stubPool{backends: []*backend.Backend{unhealthy, healthy}})

	for i := range 4 {
		got, err := wrr.GetBackend()
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if got != healthy {
			t.Errorf("call %d: expected healthy backend", i)
		}
	}
}

func TestReturnsErrNoBackendsWhenAllUnhealthy(t *testing.T) {
	wrr := NewWeightedRoundRobin(&stubPool{backends: []*backend.Backend{
		newUnhealthy(2),
		newUnhealthy(1),
	}})

	_, err := wrr.GetBackend()
	if err != ErrNoBackends {
		t.Errorf("expected ErrNoBackends, got %v", err)
	}
}

func TestWrapAround(t *testing.T) {
	a := newHealthy(1)
	b := newHealthy(1)
	wrr := NewWeightedRoundRobin(&stubPool{backends: []*backend.Backend{a, b}})

	got0, _ := wrr.GetBackend()
	got1, _ := wrr.GetBackend()
	got2, _ := wrr.GetBackend()

	if got0 != a {
		t.Error("call 0: expected a")
	}
	if got1 != b {
		t.Error("call 1: expected b")
	}
	if got2 != a {
		t.Error("call 2: expected a (wrap-around)")
	}
}

func TestZeroWeightBackendsAreExcluded(t *testing.T) {
	zero := newHealthy(0)
	nonzero := newHealthy(2)
	wrr := NewWeightedRoundRobin(&stubPool{backends: []*backend.Backend{zero, nonzero}})

	for i := range 4 {
		got, err := wrr.GetBackend()
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if got != nonzero {
			t.Errorf("call %d: zero-weight backend should be excluded", i)
		}
	}
}

func TestNewWeightedRoundRobinPanicsOnAllZeroWeights(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when all backends have zero weight")
		}
	}()
	NewWeightedRoundRobin(&stubPool{backends: []*backend.Backend{newHealthy(0)}})
}

func TestConcurrentSafety(t *testing.T) {
	backends := make([]*backend.Backend, 3)
	for i := range backends {
		backends[i] = newHealthy(i + 1)
	}
	wrr := NewWeightedRoundRobin(&stubPool{backends: backends})

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wrr.GetBackend()
		}()
	}
	wg.Wait()
}
