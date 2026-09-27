package main

import (
	"sync"
	"testing"
)

func TestChooseLeastConnectionsBackend(t *testing.T) {
	balancer := newBalancer([]string{
		"localhost:9001",
		"localhost:9002",
		"localhost:9003",
	})

	backend, err := balancer.chooseLeastConnectionsBackend()
	if err != nil {
		t.Fatal(err)
	}

	if backend.ActiveConnections != 1 {
		t.Fatalf(
			"expected selected backend to have 1 active connection, got %d",
			backend.ActiveConnections,
		)
	}
}

func TestChooseLeastConnectionsBackendConcurrent(t *testing.T) {
	balancer := newBalancer([]string{
		"localhost:9001",
		"localhost:9002",
		"localhost:9003",
	})

	var wg sync.WaitGroup

	for range 100 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := balancer.chooseLeastConnectionsBackend()
			if err != nil {
				t.Errorf("failed to choose backend: %v", err)
			}
		}()
	}

	wg.Wait()

	backends := balancer.getBackends()

	if len(backends) == 0 {
		t.Fatal("expected backends")
	}

	minConnections := backends[0].ActiveConnections
	maxConnections := backends[0].ActiveConnections
	totalConnections := 0

	for _, backend := range backends {
		totalConnections += backend.ActiveConnections

		if backend.ActiveConnections < minConnections {
			minConnections = backend.ActiveConnections
		}

		if backend.ActiveConnections > maxConnections {
			maxConnections = backend.ActiveConnections
		}
	}

	if totalConnections != 100 {
		t.Fatalf(
			"expected 100 active connections, got %d",
			totalConnections,
		)
	}

	if maxConnections-minConnections > 1 {
		t.Fatalf(
			"expected connections to be evenly distributed, min=%d max=%d",
			minConnections,
			maxConnections,
		)
	}
}

func TestUnhealthyBackendIsNotSelected(t *testing.T) {
	balancer := newBalancer([]string{
		"localhost:9001",
		"localhost:9002",
		"localhost:9003",
	})

	balancer.setBackendHealth("localhost:9002", false)

	for range 100 {
		backend, err := balancer.chooseLeastConnectionsBackend()
		if err != nil {
			t.Fatal(err)
		}

		if backend.Address == "localhost:9002" {
			t.Fatal("unhealthy backend was selected")
		}
	}
}
