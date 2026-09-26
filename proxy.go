package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Proxy struct {
	balancer    *Balancer
	dialTimeout time.Duration
	wg          sync.WaitGroup
	mu          sync.Mutex
	connections map[net.Conn]struct{}
}

func (p *Proxy) serveConnection(conn net.Conn) {
	p.wg.Add(1)
	p.addConnection(conn)

	go func() {
		defer p.wg.Done()
		defer p.removeConnection(conn)

		p.handleConnection(conn)
	}()
}

func (p *Proxy) addConnection(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.connections[conn] = struct{}{}
}

func (p *Proxy) removeConnection(conn net.Conn) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.connections, conn)
}

func (p *Proxy) closeConnections() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for conn := range p.connections {
		conn.Close()
	}
}

func (p *Proxy) handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Client connected:", conn.RemoteAddr())

	serverConn, serverData, err := p.connectToBackend()
	if err != nil {
		fmt.Println("Error connecting to a server", err)
		return
	}

	defer func() {
		p.balancer.decrementConnections(serverData.Address)
		serverConn.Close()
	}()

	fmt.Println("Connected to backend:", serverConn.RemoteAddr())

	clientTCP := conn.(*net.TCPConn)
	backendTCP := serverConn.(*net.TCPConn)

	done := make(chan error, 2)

	go func() {
		err := copyData(serverConn, conn)
		backendTCP.CloseWrite()
		done <- err
	}()

	go func() {
		err := copyData(conn, serverConn)
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

func (p *Proxy) connectToBackend() (net.Conn, Backend, error) {
	for {
		selected, err := p.balancer.chooseLeastConnectionsBackend()
		if err != nil {
			return nil, Backend{}, err
		}

		conn, err := net.DialTimeout("tcp", selected.Address, p.dialTimeout)
		if err == nil {
			return conn, selected, nil
		}

		fmt.Println("Backend unavailable:", selected.Address)
		p.balancer.decrementConnections(selected.Address)
		p.balancer.setBackendHealth(selected.Address, false)
	}
}

func (p *Proxy) wait() {
	p.wg.Wait()
}

func (p *Proxy) waitWithTimeout(timeout time.Duration) bool {
	done := make(chan struct{})

	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}
