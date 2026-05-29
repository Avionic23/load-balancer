package healthcheck

import (
	"context"
	"load-balancer/backend"
	"net"
	"testing"
	"time"
)

type stubPool struct {
	backends []*backend.Backend
}

func (s *stubPool) GetPool() []*backend.Backend {
	return s.backends
}

func startTCPServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()
	return ln.Addr().String()
}

func TestRunOnceMarksBackendHealthy(t *testing.T) {
	addr := startTCPServer(t)
	b := backend.NewBackend(backend.BackendOptions{Url: addr})
	hc := NewHealthChecker(&stubPool{backends: []*backend.Backend{b}})

	hc.RunOnce()

	if !b.IsHealthy() {
		t.Error("expected backend to be marked healthy")
	}
}

func TestRunOnceMarksBackendUnhealthy(t *testing.T) {
	b := backend.NewBackend(backend.BackendOptions{Url: "127.0.0.1:1"})
	hc := NewHealthChecker(&stubPool{backends: []*backend.Backend{b}})

	hc.RunOnce()

	if b.IsHealthy() {
		t.Error("expected backend to be marked unhealthy")
	}
}

func TestRunOnceChecksAllBackends(t *testing.T) {
	healthyAddr := startTCPServer(t)
	healthy := backend.NewBackend(backend.BackendOptions{Url: healthyAddr})
	unhealthy := backend.NewBackend(backend.BackendOptions{Url: "127.0.0.1:1"})

	hc := NewHealthChecker(&stubPool{backends: []*backend.Backend{healthy, unhealthy}})
	hc.RunOnce()

	if !healthy.IsHealthy() {
		t.Error("expected healthy backend to be marked healthy")
	}
	if unhealthy.IsHealthy() {
		t.Error("expected unreachable backend to be marked unhealthy")
	}
}

func TestRunStopsOnContextCancellation(t *testing.T) {
	b := backend.NewBackend(backend.BackendOptions{Url: "127.0.0.1:1"})
	hc := NewHealthChecker(&stubPool{backends: []*backend.Backend{b}})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		hc.Run(ctx, time.Hour)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("Run did not stop after context cancellation")
	}
}
