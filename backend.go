package main

import "log/slog"

type Backend struct {
	Address           string
	Healthy           bool
	ActiveConnections int
}

func (b *Balancer) setBackendHealth(address string, healthy bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := range b.backends {
		if b.backends[i].Address == address {
			b.backends[i].Healthy = healthy
			return
		}
	}
}

func (b *Balancer) getBackendAddresses() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	addresses := make([]string, 0, len(b.backends))

	for _, backend := range b.backends {
		addresses = append(addresses, backend.Address)
	}

	return addresses
}

func (b *Balancer) decrementConnections(address string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for i := range b.backends {
		if b.backends[i].Address == address {
			b.backends[i].ActiveConnections--
			slog.Info(
				"active connections changed",
				"backend", address,
				"active_connections", b.backends[i].ActiveConnections,
			)
			return
		}
	}
}
