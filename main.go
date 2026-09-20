package main

import (
	"fmt"
	"io"
	"net"
)

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

	backend, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		fmt.Println("Failed to connect to backend", err)
		return
	}
	defer backend.Close()

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
