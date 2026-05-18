package router

type BackendIO interface {
	GetUrl() string
}

type Router struct {
	router map[string]BackendIO
}

func NewRouter(path string, be BackendIO) *Router {
	router := make(map[string]BackendIO)
	router[path] = be
	return &Router{router: router}
}

func (r *Router) Route(path string) BackendIO {
	return r.router[path]
}
