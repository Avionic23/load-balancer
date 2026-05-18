package main

import (
	"fmt"
	backend2 "load-balancer/backend"
	listener2 "load-balancer/listener"
	proxy2 "load-balancer/proxy"
	router2 "load-balancer/router"
)

func main() {
	port := 8080
	host := "[::1]"
	backend := backend2.NewBackend("localhost:80")
	router := router2.NewRouter(host+fmt.Sprintf(":%d", port), backend)
	proxy := proxy2.NewProxy(router)
	listener := listener2.NewListener(proxy)
	listener.Listen(int64(port))
}
