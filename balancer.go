package main

import (
	"errors"
	"sync/atomic"
)

var nextBackend atomic.Uint64

func chooseLeastConnectionsBackend() (Backend, error) {
	backendsMu.RLock()
	defer backendsMu.RUnlock()

	var candidates []Backend
	minConnections := 0
	found := false

	for _, backend := range backends {
		if !backend.Healthy {
			continue
		}

		if !found || backend.ActiveConnections < minConnections  {
			found = true
			candidates = []Backend{backend}
			minConnections = backend.ActiveConnections
			continue
		}

		if backend.ActiveConnections == minConnections {
			candidates = append(candidates, backend)
		}
	}

	if !found {
		return Backend{}, errors.New("No healthy servers available")
	}

	value := nextBackend.Add(1)
	index := int(value - 1) % len(candidates)

	return candidates[index], nil
}