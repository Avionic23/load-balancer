package roundrobin

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

func newPool(urls ...string) *stubPool {
	bs := make([]*backend.Backend, len(urls))
	for i, u := range urls {
		bs[i] = backend.NewBackend(u)
	}
	return &stubPool{backends: bs}
}

func TestRoundRobinCyclesInOrder(t *testing.T) {
	rr := NewRoundRobin(newPool("a", "b", "c"))

	want := []string{"a", "b", "c", "a", "b"}
	for _, expected := range want {
		got := rr.GetBackend().GetUrl()
		if got != expected {
			t.Errorf("got %q, want %q", got, expected)
		}
	}
}

func TestRoundRobinWrapsAround(t *testing.T) {
	pool := newPool("a", "b")
	rr := NewRoundRobin(pool)

	for i := 0; i < len(pool.backends); i++ {
		rr.GetBackend()
	}

	got := rr.GetBackend().GetUrl()
	if got != "a" {
		t.Errorf("expected wrap-around to first backend, got %q", got)
	}
}

func TestRoundRobinConcurrentSafety(t *testing.T) {
	rr := NewRoundRobin(newPool("a", "b", "c"))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rr.GetBackend()
		}()
	}
	wg.Wait()
}
