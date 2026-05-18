package proxy

import (
	"fmt"
	"io"
	"load-balancer/router"
	"net"
)

type RouterIO interface {
	Route(string) router.BackendIO
}

type Proxy struct {
	router RouterIO
}

func NewProxy(rt RouterIO) *Proxy {
	return &Proxy{router: rt}
}

func (p *Proxy) Handle(conn net.Conn) error {
	localAddr := conn.LocalAddr().String()
	backend := p.router.Route(localAddr)
	if backend == nil {
		return fmt.Errorf("no available backend")
	}
	backendConn, err := net.Dial("tcp", backend.GetUrl())
	if err != nil {
		return err
	}
	defer backendConn.Close()
	go io.Copy(backendConn, conn)
	io.Copy(conn, backendConn)
	return nil
}
