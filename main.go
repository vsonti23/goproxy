package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(
		slog.NewTextHandler(os.Stdout, nil),
	)

	slog.SetDefault(logger)

	config, err := parseConfig()
	if err != nil {
		slog.Error("failed with error", "err", err)
		os.Exit(1)
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
		slog.Error("failed with error", "err", err)
		os.Exit(1)
	}

	go func() {
		<-ctx.Done()
		slog.Info("Shutting down...")
		listener.Close()
	}()

	slog.Info("Listening on", "address", config.ListenAddress)

	healthChecker := HealthChecker{
		balancer:    balancer,
		dialTimeout: config.DialTimeout,
		interval:    config.HealthInterval,
	}

	go healthChecker.run(ctx)

	proxy := Proxy{
		balancer:    balancer,
		dialTimeout: config.DialTimeout,
		connections: make(map[net.Conn]struct{}),
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				break
			}

			slog.Warn("accept error", "err", err)
			continue
		}
		proxy.serveConnection(conn)
	}

	if !proxy.waitWithTimeout(10 * time.Second) {
		slog.Info("Grace period expired, closing active connections")

		proxy.closeConnections()
		proxy.wait()
	}

	slog.Info("Shutdown complete")
}
