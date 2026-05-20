package router

type BackendIO interface {
	IsHealthy() string
}

type BackendPoolIO interface {
	Lock()
	Unlock()
	GetPool()
}

type RoundRobin struct {
}
