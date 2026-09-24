package main

import (
	"net"
	"time"
)

func checkBackendHealth(address string) bool {
	conn, err := net.DialTimeout(
		"tcp",
		address,
		2 * time.Second,
	)
	if err != nil {
		return false
	}

	defer conn.Close()
	return true
}

func runHealthChecks() {
	for {
		addresses := getBackendAddresses()
		for _, address := range addresses {
			healthy := checkBackendHealth(address)
			setBackendHealth(address, healthy)
		}

		time.Sleep(5 * time.Second)
	}
}
