package main

import (
	"errors"
	"sync/atomic"
)

var nextBackend atomic.Uint64


func chooseBackend() (Backend, error) {
	backendsMu.RLock()
	defer backendsMu.RUnlock()

	healthyBackends := make([]Backend, 0, len(backends))

	for _, backend := range backends {
		if backend.Healthy {
			healthyBackends = append(healthyBackends, backend)
		}
	}

	if len(healthyBackends) == 0 {
		return Backend{}, errors.New("No healthy servers available")
	}

	value := nextBackend.Add(1)
	index := int(value - 1) % len(healthyBackends)

	return healthyBackends[index], nil
}

func chooseLeastConnectionsBackend() (Backend, error) {
	backendsMu.RLock()
	defer backendsMu.Unlock()

	var selected Backend
	found := false

	for _, backend := range backends {
		if !backend.Healthy {
			continue
		}

		if !found {
			found = true
			selected = backend
			continue
		}

		if selected.ActiveConnections > backend.ActiveConnections {
			selected = backend
		}
	}

	if !found {
		return Backend{}, errors.New("No healthy servers available")
	}

	return selected, nil
}