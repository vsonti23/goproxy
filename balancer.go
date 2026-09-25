package main

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

type Balancer struct {
	backends    []Backend
	mu          sync.RWMutex
	nextBackend atomic.Uint64
}

func newBalancer(addresses []string) *Balancer {
	backends := make([]Backend, 0, len(addresses))

	for _, address := range addresses {
		backends = append(backends, Backend{
			Address: address,
			Healthy: true,
		})
	}

	return &Balancer{
		backends: backends,
	}
}

func (b *Balancer) chooseLeastConnectionsBackend() (Backend, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	var candidateIndices []int
	minConnections := 0
	found := false

	for i := range b.backends {
		if !b.backends[i].Healthy {
			continue
		}

		if !found || b.backends[i].ActiveConnections < minConnections {
			found = true
			candidateIndices = []int{i}
			minConnections = b.backends[i].ActiveConnections
			continue
		}

		if b.backends[i].ActiveConnections == minConnections {
			candidateIndices = append(candidateIndices, i)
		}
	}

	if !found {
		return Backend{}, errors.New("No healthy servers available")
	}

	value := b.nextBackend.Add(1)
	candidateIndex := int(value-1) % len(candidateIndices)
	index := candidateIndices[candidateIndex]
	b.backends[index].ActiveConnections++
	fmt.Printf(
		"%s active connections: %d\n",
		b.backends[index].Address,
		b.backends[index].ActiveConnections,
	)

	return b.backends[index], nil
}
