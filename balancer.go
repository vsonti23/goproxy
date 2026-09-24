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