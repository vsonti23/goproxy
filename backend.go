package main

import (
	"fmt"
	"sync"
)

type Backend struct {
	Address string
	Healthy bool
	ActiveConnections int
}

var backends = []Backend{
	{ Address: "localhost:9001", Healthy: true, ActiveConnections: 0 },
	{ Address: "localhost:9002", Healthy: true, ActiveConnections: 0 },
	{ Address: "localhost:9003", Healthy: true, ActiveConnections: 0 },
}

var backendsMu sync.RWMutex

func setBackendHealth(address string, healthy bool) {
	backendsMu.Lock()
	defer backendsMu.Unlock()

	for i := range backends {
		if backends[i].Address == address {
			backends[i].Healthy = healthy
			return
		}
	}
}

func getBackendAddresses() []string {
	backendsMu.RLock()
	defer backendsMu.RUnlock()

	addresses := make([]string, 0, len(backends))
	
	for _, backend := range backends {
		addresses = append(addresses, backend.Address)
	}

	return addresses
}

func decrementConnections(address string) {
	backendsMu.Lock()
	defer backendsMu.Unlock()

	for i := range backends {
		if backends[i].Address == address {
			backends[i].ActiveConnections--
			fmt.Printf(
				"%s active connections: %d\n",
				address,
				backends[i].ActiveConnections,
			)
			return
		}
	}
}