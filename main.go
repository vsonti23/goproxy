package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	config, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	balancer := newBalancer(config.BackendAddresses)

	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	fmt.Printf("Listening on %s\n", config.ListenAddress)

	healthChecker := HealthChecker{
		balancer:    balancer,
		dialTimeout: config.DialTimeout,
		interval:    config.HealthInterval,
	}

	go healthChecker.run()

	proxy := Proxy{
		balancer:    balancer,
		dialTimeout: config.DialTimeout,
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go proxy.handleConnection(conn)
	}
}
