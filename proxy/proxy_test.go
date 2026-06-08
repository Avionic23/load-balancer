package proxy

import (
	"errors"
	"io"
	"load-balancer/backend"
	"net"
	"testing"
	"time"
)

var errNoBackend = errors.New("no backend available")

type stubRouter struct {
	b   *backend.Backend
	err error
}

func (s *stubRouter) Route(_ string) (*backend.Backend, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.b, nil
}

func defaultOpts() ProxyOptions {
	return ProxyOptions{
		DialTimeout: 10 * time.Second,
		ConnTimeout: 30 * time.Second,
	}
}

func startEchoServer(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		io.Copy(conn, conn)
	}()
	return ln.Addr().String()
}

func TestHandleForwardsData(t *testing.T) {
	addr := startEchoServer(t)
	b := backend.NewBackend(backend.BackendOptions{Url: addr})
	p := NewProxy(&stubRouter{b: b}, defaultOpts())

	clientConn, proxyConn := net.Pipe()
	defer clientConn.Close()
	go p.Handle(proxyConn)

	if _, err := clientConn.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, 5)
	if _, err := io.ReadFull(clientConn, buf); err != nil {
		t.Fatal(err)
	}
	if string(buf) != "hello" {
		t.Errorf("got %q, want %q", string(buf), "hello")
	}
}

func TestHandleRoutingError(t *testing.T) {
	p := NewProxy(&stubRouter{err: errNoBackend}, defaultOpts())
	_, proxyConn := net.Pipe()

	err := p.Handle(proxyConn)
	if !errors.Is(err, errNoBackend) {
		t.Errorf("expected errNoBackend, got %v", err)
	}
}

func TestHandleDeadlineExceeded(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		select {}
	}()

	b := backend.NewBackend(backend.BackendOptions{Url: ln.Addr().String()})
	p := NewProxy(&stubRouter{b: b}, ProxyOptions{
		DialTimeout: 10 * time.Second,
		ConnTimeout: 5 * time.Millisecond,
	})

	clientConn, proxyConn := net.Pipe()
	defer clientConn.Close()

	err = p.Handle(proxyConn)
	proxyConn.Close()

	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Errorf("expected timeout error, got %v", err)
	}

	_, writeErr := clientConn.Write([]byte("x"))
	if writeErr == nil {
		t.Error("expected clientConn write to fail after proxyConn closed")
	}
}

func TestHandleBackendUnreachable(t *testing.T) {
	b := backend.NewBackend(backend.BackendOptions{Url: "127.0.0.1:1"})
	p := NewProxy(&stubRouter{b: b}, defaultOpts())
	_, proxyConn := net.Pipe()

	err := p.Handle(proxyConn)
	if err == nil {
		t.Error("expected error when backend is unreachable")
	}
}

func TestActiveConnsIncrementedDuringConnection(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	connEstablished := make(chan struct{})
	unblock := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		close(connEstablished)
		<-unblock
	}()

	b := backend.NewBackend(backend.BackendOptions{Url: ln.Addr().String()})
	p := NewProxy(&stubRouter{b: b}, defaultOpts())

	clientConn, proxyConn := net.Pipe()
	defer clientConn.Close()

	done := make(chan struct{})
	go func() {
		p.Handle(proxyConn)
		close(done)
	}()

	<-connEstablished
	deadline := time.After(time.Second)
	for b.GetActiveConns() != 1 {
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for active conn increment, got %d", b.GetActiveConns())
		default:
		}
	}

	close(unblock)
	clientConn.Close()
	<-done

	if b.GetActiveConns() != 0 {
		t.Errorf("expected 0 active conns after Handle returned, got %d", b.GetActiveConns())
	}
}

func TestActiveConnsNotIncrementedOnFailedDial(t *testing.T) {
	b := backend.NewBackend(backend.BackendOptions{Url: "127.0.0.1:1"})
	p := NewProxy(&stubRouter{b: b}, defaultOpts())
	_, proxyConn := net.Pipe()

	p.Handle(proxyConn)

	if b.GetActiveConns() != 0 {
		t.Errorf("expected 0 active conns after failed dial, got %d", b.GetActiveConns())
	}
}
