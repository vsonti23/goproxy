package main

import (
	"net"
	"time"
)

func checkBackendHealth(address string, timeout time.Duration) bool {
	conn, err := net.DialTimeout(
		"tcp",
		address,
		timeout,
	)
	if err != nil {
		return false
	}

	defer conn.Close()
	return true
}

func runHealthChecks(balancer *Balancer, config *Config) {
	for {
		addresses := balancer.getBackendAddresses()
		for _, address := range addresses {
			healthy := checkBackendHealth(address, config.DialTimeout)
			balancer.setBackendHealth(address, healthy)
		}

		time.Sleep(config.HealthInterval)
	}
}
