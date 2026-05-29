package healthcheck

import (
	"context"
	"load-balancer/backend"
	"log"
	"net"
	"sync"
	"time"
)

type BackendPoolIO interface {
	GetPool() []*backend.Backend
}

type HealthChecker struct {
	bp BackendPoolIO
}

func NewHealthChecker(bp BackendPoolIO) *HealthChecker {
	if bp == nil {
		panic("backend pool cannot be nil")
	}
	return &HealthChecker{bp: bp}
}

func (hc *HealthChecker) Run(ctx context.Context, duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			hc.RunOnce()
		case <-ctx.Done():
			return
		}
	}
}

func (hc *HealthChecker) RunOnce() {
	bp := hc.bp.GetPool()
	wg := sync.WaitGroup{}
	for _, b := range bp {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.DialTimeout("tcp", b.GetUrl(), 10*time.Second)
			if err != nil {
				b.SetHealth(false)
				log.Printf("dial %s: %v", b.GetUrl(), err)
				return
			}
			b.SetHealth(true)
			conn.Close()
		}()
	}
	wg.Wait()
}
