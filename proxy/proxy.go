package proxy

import (
	"fmt"
	"io"
	"load-balancer/backend"
	"net"
	"time"
)

type RouterIO interface {
	Route(string) (*backend.Backend, error)
}

type Proxy struct {
	router      RouterIO
	dialTimeout time.Duration
	connTimeout time.Duration
}

type ProxyOptions struct {
	DialTimeout time.Duration
	ConnTimeout time.Duration
}

func NewProxy(rt RouterIO, opts ProxyOptions) *Proxy {
	if rt == nil {
		panic("router cannot be nil")
	}
	return &Proxy{
		router:      rt,
		dialTimeout: opts.DialTimeout,
		connTimeout: opts.ConnTimeout,
	}
}

func (p *Proxy) Handle(conn net.Conn) error {
	localAddr := conn.LocalAddr().String()
	b, err := p.router.Route(localAddr)
	if err != nil {
		return fmt.Errorf("routing: %w", err)
	}

	backendConn, err := net.DialTimeout("tcp", b.GetUrl(), p.dialTimeout)
	if err != nil {
		return fmt.Errorf("dial %s: %w", b.GetUrl(), err)
	}

	b.IncrActiveConns()
	defer b.DecrActiveConns()
	defer backendConn.Close()

	err = conn.SetDeadline(time.Now().Add(p.connTimeout))
	if err != nil {
		return fmt.Errorf("set deadline %s: %w", conn.RemoteAddr(), err)
	}
	err = backendConn.SetDeadline(time.Now().Add(p.connTimeout))
	if err != nil {
		return fmt.Errorf("set deadline %s: %w", backendConn.RemoteAddr(), err)
	}

	ch := make(chan error, 2)
	go func() {
		_, localErr := io.Copy(backendConn, conn)
		if localErr != nil {
			localErr = fmt.Errorf("copy from %s to %s: %w", conn.RemoteAddr(), backendConn.RemoteAddr(), localErr)
		}
		ch <- localErr
	}()
	go func() {
		_, localErr := io.Copy(conn, backendConn)
		if localErr != nil {
			localErr = fmt.Errorf("copy from %s to %s: %w", backendConn.RemoteAddr(), conn.RemoteAddr(), localErr)
		}
		ch <- localErr
	}()

	for range 2 {
		if err = <-ch; err != nil {
			return err
		}
	}

	return nil
}
