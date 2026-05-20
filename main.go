package main

import (
	"fmt"
	"load-balancer/backend"
	"load-balancer/listener"
	"load-balancer/proxy"
	"load-balancer/router"
	"load-balancer/router/roundrobin"
)

func main() {
	port := 8080
	host := "[::1]"
	b := backend.NewBackend("localhost:80")
	b1 := backend.NewBackend("localhost:8081")
	bp := backend.NewBackendPool([]*backend.Backend{b, b1})
	algo := roundrobin.NewRoundRobin(bp)
	r := router.NewRouter(host+fmt.Sprintf(":%d", port), algo)
	p := proxy.NewProxy(r)
	l := listener.NewListener(p)
	l.Listen(int64(port))
}
