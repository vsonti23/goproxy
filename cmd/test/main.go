package main

import (
	"io"
	"log/slog"
	"net"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}

	conn, err := listener.Accept()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	data, err := io.ReadAll(conn)
	if err != nil {
		panic(err)
	}

	slog.Warn("Got EOF after receiving", "data", data)

	// We received FIN, but our write direction is still open.
	slog.Info("Waiting 5 seconds...")
	time.Sleep(5 * time.Second)

	slog.Info("Sending response after EOF")
	conn.Write([]byte("backend response\n"))
}
