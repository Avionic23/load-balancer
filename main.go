package main

import (
	"context"
	"fmt"
	"load-balancer/backend"
	"load-balancer/listener"
	"load-balancer/proxy"
	"load-balancer/router"
	"load-balancer/router/roundrobin"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	port := 8080
	host := "[::1]"
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", port, err)
	}
	b := backend.NewBackend("localhost:80")
	b1 := backend.NewBackend("localhost:8081")
	bp := backend.NewBackendPool([]*backend.Backend{b, b1})
	algo := roundrobin.NewRoundRobin(bp)
	r := router.NewRouter(host+fmt.Sprintf(":%d", port), algo)
	p := proxy.NewProxy(r)
	l := listener.NewListener(p)

	go l.Listen(ln)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	l.GracefulShutdown(ctx)
}
