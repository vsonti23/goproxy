package main

import (
	"net"
	"time"
)

type HealthChecker struct {
	balancer    *Balancer
	interval    time.Duration
	dialTimeout time.Duration
}

func (hc *HealthChecker) check(address string) bool {
	conn, err := net.DialTimeout(
		"tcp",
		address,
		hc.dialTimeout,
	)
	if err != nil {
		return false
	}

	defer conn.Close()
	return true
}

func (hc *HealthChecker) run() {
	for {
		addresses := hc.balancer.getBackendAddresses()
		for _, address := range addresses {
			healthy := hc.check(address)
			hc.balancer.setBackendHealth(address, healthy)
		}

		time.Sleep(hc.interval)
	}
}
