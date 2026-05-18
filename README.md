# Load Balancer

A TCP load balancer written in Go as a system design learning project.

## Architecture

The project is composed of four layered packages:

```
Listener → Proxy → Router → Backend
```

- **listener** — accepts incoming TCP connections and dispatches each to the proxy in a goroutine
- **proxy** — forwards traffic between the client and the selected backend using bidirectional byte copying
- **router** — maps a local address to a backend; the extension point for load balancing strategies
- **backend** — represents a single upstream server

## Getting Started

### Prerequisites

- Go 1.24+

### Run

```bash
go run main.go
```

The load balancer listens on port `8080` by default and forwards traffic to `localhost:80`.

### Build

```bash
go build -o load-balancer .
./load-balancer
```

## Project Structure

```
.
├── main.go          # entry point — wires all components together
├── backend/
│   └── backend.go   # upstream server representation
├── router/
│   └── router.go    # routing logic (address → backend)
├── proxy/
│   └── proxy.go     # TCP proxy (bidirectional copy)
└── listener/
    └── listener.go  # TCP listener
```

## Extending

To add a load balancing algorithm (round-robin, least-connections, etc.), update `router/router.go` — it implements the `RouterIO` interface consumed by the proxy:

```go
type RouterIO interface {
    Route(address string) BackendIO
}
```

Multiple backends can be registered and the `Route` method can select among them using any strategy.
