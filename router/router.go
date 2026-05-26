package router

import "load-balancer/backend"

type AlgoIO interface {
	GetBackend() (*backend.Backend, error)
}

type Router struct {
	router map[string]AlgoIO
}

func NewRouter(path string, be AlgoIO) *Router {
	if be == nil {
		panic("algorithm cannot be nil")
	}
	router := make(map[string]AlgoIO)
	router[path] = be
	return &Router{router: router}
}

func (r *Router) Route(path string) (*backend.Backend, error) {
	return r.router[path].GetBackend()
}
