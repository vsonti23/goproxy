package main

import (
	"fmt"
	"io"
	"net"
	"time"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Client connected:", conn.RemoteAddr())
	
	serverConn, serverData, err := connectToBackend()
	if err != nil {
		fmt.Println("Error connecting to a server", err)
		return
	}
	
	defer func() {
		decrementConnections(serverData.Address)
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

func connectToBackend() (net.Conn, Backend, error) {
	for {
		selected, err := chooseLeastConnectionsBackend()
		if err != nil {
			return nil, Backend{}, err
		}

		conn, err := net.DialTimeout("tcp", selected.Address, 2 * time.Second)
		if err == nil {
			return conn, selected, nil
		}

		fmt.Println("Backend unavailable:", selected.Address)
		decrementConnections(selected.Address)
		setBackendHealth(selected.Address, false)
	}
}