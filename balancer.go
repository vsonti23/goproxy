package main

import (
	"errors"
	"fmt"
	"sync/atomic"
)

var nextBackend atomic.Uint64

func chooseLeastConnectionsBackend() (Backend, error) {
	backendsMu.Lock()
	defer backendsMu.Unlock()

	var candidateIndices []int
	minConnections := 0
	found := false

	for i := range backends {
		if !backends[i].Healthy {
			continue
		}

		if !found || backends[i].ActiveConnections < minConnections  {
			found = true
			candidateIndices = []int{i}
			minConnections = backends[i].ActiveConnections
			continue
		}

		if backends[i].ActiveConnections == minConnections {
			candidateIndices = append(candidateIndices, i)
		}
	}

	if !found {
		return Backend{}, errors.New("No healthy servers available")
	}

	value := nextBackend.Add(1)
	candidateIndex := int(value - 1) % len(candidateIndices)
	index := candidateIndices[candidateIndex]
	backends[index].ActiveConnections++
	fmt.Printf(
		"%s active connections: %d\n",
		backends[index].Address,
		backends[index].ActiveConnections,
	)

	return backends[index], nil
}