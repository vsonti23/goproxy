package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
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

	go runHealthChecks()

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

	healthyBackends := make([]Backend, 0, len(backends))

	for _, backend := range backends {
		if backend.Healthy {
			healthyBackends = append(healthyBackends, backend)
		}
	}

	if len(healthyBackends) == 0 {
		return Backend{}, errors.New("No healthy servers available")
	}

	value := nextBackend.Add(1)
	index := int(value - 1) % len(healthyBackends)

	return healthyBackends[index], nil
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
		setBackendHealth(selected.Address, false)
	}
}

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

func setBackendHealth(address string, healthy bool) {
	backendsMu.Lock()
	defer backendsMu.Unlock()

	for i := range backends {
		if backends[i].Address == address {
			backends[i].Healthy = healthy
			return
		}
	}
}

func getBackendAddresses() []string {
	backendsMu.RLock()
	defer backendsMu.RUnlock()

	addresses := make([]string, 0, len(backends))
	
	for _, backend := range backends {
		addresses = append(addresses, backend.Address)
	}

	return addresses
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