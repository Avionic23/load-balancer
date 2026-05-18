package listener

import (
	"fmt"
	"net"
)

type ProxyIO interface {
	Handle(net.Conn) error
}

type Listener struct {
	proxy ProxyIO
}

func NewListener(px ProxyIO) *Listener {
	return &Listener{proxy: px}
}

func (l *Listener) Listen(port int64) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go func() {
			defer conn.Close()
			err = l.proxy.Handle(conn)
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
}
