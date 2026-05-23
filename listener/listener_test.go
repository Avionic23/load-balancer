package listener

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

type stubProxy struct {
	called chan struct{}
	block  chan struct{}
}

func newStubProxy() *stubProxy {
	return &stubProxy{
		called: make(chan struct{}, 1),
		block:  make(chan struct{}),
	}
}

func (s *stubProxy) Handle(conn net.Conn) error {
	s.called <- struct{}{}
	<-s.block
	return nil
}

func startListener(t *testing.T, l *Listener) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go l.Listen(ln)
	return ln.Addr().String()
}

func TestListenAcceptsConnection(t *testing.T) {
	proxy := newStubProxy()
	l := NewListener(proxy)
	addr := startListener(t, l)
	defer l.GracefulShutdown(context.Background())

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	close(proxy.block)

	select {
	case <-proxy.called:
	case <-time.After(time.Second):
		t.Error("proxy Handle was not called")
	}
}

func TestConcurrentListenAcceptsConnection(t *testing.T) {
	count := 100
	proxy := &stubProxy{
		called: make(chan struct{}, count),
		block:  make(chan struct{}),
	}
	close(proxy.block)
	l := NewListener(proxy)
	addr := startListener(t, l)
	defer l.GracefulShutdown(context.Background())

	wg := sync.WaitGroup{}
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				return
			}
			conn.Close()
		}()
	}
	wg.Wait()

	for i := 0; i < 100; i++ {
		select {
		case <-proxy.called:
		case <-time.After(time.Second):
			t.Errorf("only %d of 100 connections were handled", i)
			return
		}
	}
}

func TestGracefulShutdownWaitsForConnections(t *testing.T) {
	proxy := newStubProxy()
	l := NewListener(proxy)
	addr := startListener(t, l)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	<-proxy.called

	done := make(chan struct{})
	go func() {
		l.GracefulShutdown(context.Background())
		close(done)
	}()

	select {
	case <-done:
		t.Error("GracefulShutdown returned before connection finished")
	case <-time.After(50 * time.Millisecond):
	}

	close(proxy.block)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("GracefulShutdown did not return after connection finished")
	}
}

func TestGracefulShutdownForcesCloseOnDeadline(t *testing.T) {
	proxy := newStubProxy()
	l := NewListener(proxy)
	addr := startListener(t, l)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	<-proxy.called

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		l.GracefulShutdown(ctx)
		close(done)
	}()

	<-ctx.Done()
	close(proxy.block)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Error("GracefulShutdown did not force close on deadline")
	}
}
