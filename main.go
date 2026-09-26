package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	balancer := newBalancer(config.BackendAddresses)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-ctx.Done()
		fmt.Println("\nShutting down...")
		listener.Close()
	}()

	fmt.Printf("Listening on %s\n", config.ListenAddress)

	healthChecker := HealthChecker{
		balancer:    balancer,
		dialTimeout: config.DialTimeout,
		interval:    config.HealthInterval,
	}

	go healthChecker.run(ctx)

	proxy := Proxy{
		balancer:    balancer,
		dialTimeout: config.DialTimeout,
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				break
			}

			log.Printf("accept error: %v", err)
			continue
		}
		proxy.serveConnection(conn)
	}

	if proxy.waitWithTimeout(10 * time.Second) {
		fmt.Println("All connections drained")
	} else {
		fmt.Println("Shutdown timed out")
	}
}
