package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
)

type Backend struct {
	Address string
	Healthy bool
}

var backends = []Backend{
	{ Address: "localhost:9001", Healthy: true },
	{ Address: "localhost:9002", Healthy: true },
	{ Address: "localhost:9003", Healthy: true },
}

var backendsMu sync.RWMutex

var nextBackend atomic.Uint64

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	
	fmt.Println("Listening on :8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}	
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Client connected:", conn.RemoteAddr())
	
	backend, err := connectToBackend()
	if err != nil {
		fmt.Println("Error connecting to a server", err)
		return
	}

	defer backend.Close()

	if err != nil {
		
	}

	fmt.Println("Connected to backend:", backend.RemoteAddr())

	clientTCP := conn.(*net.TCPConn)
	backendTCP := backend.(*net.TCPConn)

	done := make(chan error, 2)

	go func() {
		err := copyData(backend, conn)
		backendTCP.CloseWrite()
		done <- err
	}()

	go func() {
		err := copyData(conn, backend)
		clientTCP.CloseWrite()
		done <- err
	}()

	err1 := <-done
	err2 := <-done

	fmt.Println("Closing Connection")

	if err1 != nil {
		fmt.Println("Proxy connection ended with error:", err1)
	}

	if err2 != nil {
		fmt.Println("Proxy connection ended with error:", err2)
	}
}

func copyData(dst net.Conn, src net.Conn) error {
	_, err := io.Copy(dst, src)
	return err
}

func chooseBackend() (Backend, error) {
	backendsMu.RLock()
	defer backendsMu.RUnlock()

	value := nextBackend.Add(1)
	start := int(value - 1) % len(backends)

	for i := 0; i < len(backends); i++ {
		index := (start + i) % len(backends)
		if backends[index].Healthy {
			return backends[index], nil
		}
	}

	return Backend{}, errors.New("No healthy servers available")
}

func markBackendUnhealthy(address string) {
	backendsMu.Lock()
	defer backendsMu.Unlock()

	for i := range backends {
		if backends[i].Address == address {
			backends[i].Healthy = false
			return
		}
	}
}

func connectToBackend() (net.Conn, error) {
	for {
		selected, err := chooseBackend()
		if err != nil {
			return nil, err
		}

		conn, err := net.Dial("tcp", selected.Address)
		if err == nil {
			return conn, nil
		}

		fmt.Println("Backend unavailable:", selected.Address)
		markBackendUnhealthy(selected.Address)
	}
}