package backend

type Backend struct {
	url string
}

func NewBackend(url string) *Backend {
	return &Backend{url: url}
}

func (b *Backend) GetUrl() string {
	return b.url
}
