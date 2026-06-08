package main

import (
	"context"
	"fmt"
	"load-balancer/backend"
	"load-balancer/config"
	"load-balancer/healthcheck"
	"load-balancer/listener"
	"load-balancer/proxy"
	"load-balancer/router"
	"load-balancer/router/leastconnections"
	"load-balancer/router/roundrobin"
	"load-balancer/router/weightedroundrobin"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	conf := config.Load("config/config_test.yaml")

	host := "[::1]"
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", conf.Port, err)
	}

	backends := make([]*backend.Backend, 0, len(conf.Backends))
	for _, b := range conf.Backends {
		backends = append(backends, backend.NewBackend(backend.BackendOptions{
			Url:    b.Url,
			Weight: b.Weight,
		}))
	}
	bp := backend.NewBackendPool(backends)

	var algo router.AlgoIO
	switch conf.Algorithm {
	case "round-robin":
		algo = roundrobin.NewRoundRobin(bp)
	case "weighted-round-robin":
		algo = weightedroundrobin.NewWeightedRoundRobin(bp)
	case "least_connections":
		algo = leastconnections.NewLeastConnections(bp)
	default:
		panic("unknown algorithm")
	}

	r := router.NewRouter(host+fmt.Sprintf(":%d", conf.Port), algo)
	p := proxy.NewProxy(r, proxy.ProxyOptions{
		DialTimeout: conf.Timeouts.Dial,
		ConnTimeout: conf.Timeouts.Connect,
	})
	l := listener.NewListener(p)

	hc := healthcheck.NewHealthChecker(bp)
	hc.RunOnce()

	ctx, cancel := context.WithTimeout(context.Background(), conf.Timeouts.Shutdown)
	defer cancel()

	go hc.Run(ctx, conf.Timeouts.HealthCheck)
	go l.Listen(ln)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	<-sigCh

	l.GracefulShutdown(ctx)
}
