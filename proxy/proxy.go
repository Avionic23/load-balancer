package proxy

import (
	"fmt"
	"io"
	"load-balancer/backend"
	"net"
)

type BackendIO interface {
	GetUrl() string
}

type RouterIO interface {
	Route(string) *backend.Backend
}

type Proxy struct {
	router RouterIO
}

func NewProxy(rt RouterIO) *Proxy {
	return &Proxy{router: rt}
}

func (p *Proxy) Handle(conn net.Conn) error {
	localAddr := conn.LocalAddr().String()
	b := p.router.Route(localAddr)
	if b == nil {
		return fmt.Errorf("no available backend")
	}
	fmt.Println(b.GetUrl())
	backendConn, err := net.Dial("tcp", b.GetUrl())
	if err != nil {
		return err
	}
	defer backendConn.Close()
	go io.Copy(backendConn, conn)
	io.Copy(conn, backendConn)
	return nil
}
