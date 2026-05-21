package proxy

import (
	"errors"
	"io"
	"load-balancer/backend"
	"net"
	"testing"
	"time"
)

type stubRouter struct {
	b *backend.Backend
}

func (s *stubRouter) Route(_ string) *backend.Backend {
	return s.b
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
	p := NewProxy(&stubRouter{b: backend.NewBackend(addr)})

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

func TestHandleNoBackend(t *testing.T) {
	p := NewProxy(&stubRouter{b: nil})
	_, proxyConn := net.Pipe()

	err := p.Handle(proxyConn)
	if err == nil {
		t.Error("expected error when no backend available")
	}
}

func TestHandleDeadlineExceeded(t *testing.T) {
	// backend accepts but never sends anything — deadline fires first
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
		// block forever — never read or write
		select {}
	}()

	p := &Proxy{
		router:      &stubRouter{b: backend.NewBackend(ln.Addr().String())},
		dialTimeout: 10 * time.Second,
		connTimeout: 5 * time.Millisecond,
	}

	clientConn, proxyConn := net.Pipe()
	defer clientConn.Close()

	err = p.Handle(proxyConn)

	var netErr net.Error
	if !errors.As(err, &netErr) || !netErr.Timeout() {
		t.Errorf("expected timeout error, got %v", err)
	}

	// proxyConn was closed by defer conn.Close() inside Handle
	// so writing to clientConn should now fail
	_, writeErr := clientConn.Write([]byte("x"))
	if writeErr == nil {
		t.Error("expected clientConn write to fail after Handle returned")
	}
}

func TestHandleBackendUnreachable(t *testing.T) {
	p := NewProxy(&stubRouter{b: backend.NewBackend("127.0.0.1:1")})
	_, proxyConn := net.Pipe()

	err := p.Handle(proxyConn)
	if err == nil {
		t.Error("expected error when backend is unreachable")
	}
}
