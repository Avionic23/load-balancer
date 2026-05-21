package proxy

import (
	"fmt"
	"io"
	"load-balancer/backend"
	"net"
	"time"
)

type RouterIO interface {
	Route(string) *backend.Backend
}

type Proxy struct {
	router      RouterIO
	dialTimeout time.Duration
	connTimeout time.Duration
}

func NewProxy(rt RouterIO) *Proxy {
	return &Proxy{
		router:      rt,
		dialTimeout: 10 * time.Second,
		connTimeout: 20 * time.Second,
	}
}

func (p *Proxy) Handle(conn net.Conn) error {
	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	b := p.router.Route(localAddr)
	if b == nil {
		return fmt.Errorf("no available backend")
	}
	backendConn, err := net.DialTimeout("tcp", b.GetUrl(), p.dialTimeout)
	if err != nil {
		return err
	}
	defer backendConn.Close()
	err = conn.SetDeadline(time.Now().Add(p.connTimeout))
	if err != nil {
		return err
	}
	err = backendConn.SetDeadline(time.Now().Add(p.connTimeout))
	if err != nil {
		return err
	}
	ch := make(chan error, 2)
	go func() {
		_, localErr := io.Copy(backendConn, conn)
		ch <- localErr
	}()
	go func() {
		_, localErr := io.Copy(conn, backendConn)
		ch <- localErr
	}()
	for range 2 {
		if err = <-ch; err != nil {
			return err
		}
	}
	return nil
}
