package router

import "load-balancer/backend"

type AlgoIO interface {
	GetBackend() *backend.Backend
}

type Router struct {
	router map[string]AlgoIO
}

func NewRouter(path string, be AlgoIO) *Router {
	router := make(map[string]AlgoIO)
	router[path] = be
	return &Router{router: router}
}

func (r *Router) Route(path string) *backend.Backend {
	return r.router[path].GetBackend()
}
