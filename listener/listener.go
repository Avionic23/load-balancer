package listener

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

type ProxyIO interface {
	Handle(net.Conn) error
}

type Listener struct {
	proxy       ProxyIO
	ln          net.Listener
	activeConns map[net.Conn]struct{}
	mu          sync.Mutex
	wg          sync.WaitGroup
}

func NewListener(px ProxyIO) *Listener {
	if px == nil {
		panic("proxy cannot be nil")
	}
	return &Listener{
		proxy:       px,
		activeConns: make(map[net.Conn]struct{}),
	}
}

func (l *Listener) Listen(ln net.Listener) {
	l.ln = ln
	defer ln.Close()
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("accept error: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		l.mu.Lock()
		l.activeConns[conn] = struct{}{}
		l.mu.Unlock()
		l.wg.Add(1)
		go func() {
			defer l.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("panic handling connection %s: %v",
						conn.RemoteAddr(), r)
				}
			}()
			defer conn.Close()
			defer func() {
				l.mu.Lock()
				delete(l.activeConns, conn)
				l.mu.Unlock()
			}()
			err := l.proxy.Handle(conn)
			if err != nil {
				log.Printf("connection %s: %v", conn.RemoteAddr(), err)
			}
		}()
	}
}

func (l *Listener) GracefulShutdown(ctx context.Context) {
	l.ln.Close()

	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return
	case <-ctx.Done():
		l.mu.Lock()
		for conn := range l.activeConns {
			conn.Close()
		}
		l.mu.Unlock()
		<-done
	}
}
